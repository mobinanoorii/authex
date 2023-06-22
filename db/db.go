package db

import (
	"authex/helpers"
	"authex/model"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/gommon/log"
	"github.com/shopspring/decimal"

	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
)

//go:embed schema.sql
var dbSchema string

type Connection struct {
	pool      *pgxpool.Pool
	Matches   chan *model.Match
	Transfers chan *model.BalanceChange
}

// Close the connection and all channels
func (c *Connection) Close() {
	close(c.Matches)
	close(c.Transfers)
	c.pool.Close()
}

func NewConnection(options *model.Settings) (*Connection, error) {
	pool, err := pgxpool.New(context.Background(), options.DB.URI)
	pool.Config().AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		// enable decimal support
		pgxdecimal.Register(conn.TypeMap())
		return nil
	}
	if err != nil {
		return nil, err
	}
	return &Connection{
		pool:      pool,
		Matches:   make(chan *model.Match),
		Transfers: make(chan *model.BalanceChange),
	}, nil
}

func (c *Connection) Run() {
	// TODO: handle goroutines lifecycle properly
	wg := sync.WaitGroup{}
	wg.Add(2)

	// handle ERC20 transfers
	go func() {
		for {
			t, ok := <-c.Transfers
			if !ok {
				wg.Done()
				log.Debugf("closing balance handler")
			}
			tx, err := c.pool.BeginTx(context.Background(), pgx.TxOptions{})
			if err != nil {
				log.Errorf("error starting transaction: %v", err)
			}
			q := `INSERT INTO balances (address, asset_address, balance) VALUES ($1, $2, $3) ON CONFLICT (address, asset_address) DO UPDATE SET balance = balances.balance + $3`
			for _, delta := range t.Deltas {
				if _, err := tx.Exec(context.Background(), q, delta.Address, t.TokenAddress, delta.Amount); err != nil {
					log.Errorf("error updating the recipient balance: %v", err)
					tx.Rollback(context.Background())
					break
				}
			}
			// update token block number
			if _, err := tx.Exec(context.Background(), "UPDATE assets SET last_block = $1 WHERE address = $2", t.BlockNumber, t.TokenAddress); err != nil {
				log.Errorf("error updating the asset block number: %v", err)
				tx.Rollback(context.Background())
				break
			}

			tx.Commit(context.Background())
		}
	}()

	// handle CLOB matches
	go func() {
		for {
			match, ok := <-c.Matches
			if !ok {
				wg.Done()
				log.Debugf("closing match handler")
			}
			c.handleMatch(match)
		}
	}()

	wg.Wait()
}

func (c *Connection) handleMatch(m *model.Match) {
	tx, err := c.pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		log.Errorf("error starting transaction: %v", err)
		return
	}
	// insert into matches
	q := `INSERT INTO matches
	(id, order_id, price, size, side, matched_at, status)
	VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(context.Background(), q, m.ID, m.OrderID, m.Price, m.Size, m.Side, m.Time, m.Status)
	if err != nil {
		log.Errorf("error inserting match: %v", err)
		tx.Rollback(context.Background())
		return
	}
	// update orders
	quote := m.Price.Mul(m.Size)
	// update balances
	switch m.Side {
	case model.SideBid:
		q = `
		WITH order_details AS (
			SELECT o.from_address, m.base_address, m.quote_address
			FROM orders o JOIN markets m ON o.market_address = m.address
			WHERE o.id = $1
		)
		INSERT INTO balances (address, asset_address, balance)
		SELECT from_address, quote_address, $2
		FROM order_details
		ON CONFLICT (address, asset_address) DO UPDATE SET balance = balances.balance + $2`
	case model.SideAsk:
		q = `
		WITH order_details AS (
			SELECT o.from_address, m.base_address, m.quote_address
			FROM orders o JOIN markets m ON o.market_address = m.address
			WHERE o.id = $1
		)
		INSERT INTO balances (address, asset_address, balance)
		SELECT from_address, base_address, $2
		FROM order_details
		ON CONFLICT (address, asset_address) DO UPDATE SET balance = balances.balance + $2`
	default:
		log.Errorf("unknown side: %s", m.Side)
		tx.Rollback(context.Background())
		return
	}
	_, err = tx.Exec(context.Background(), q, m.OrderID, quote)
	if err != nil {
		log.Errorf("handleMatch - error updating balance: %v", err)
		tx.Rollback(context.Background())
	}
	tx.Commit(context.Background())
}

// Setup the database, open a connection and create the database schema
func Setup(options *model.Settings) error {
	conn, err := pgx.Connect(context.Background(), options.DB.URI)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())
	_, err = conn.Exec(context.Background(), dbSchema)
	return err
}

// SaveMarket saves a market to the database
func (c *Connection) SaveMarket(marketAddress string, base, quote *model.Asset) error {
	tx, err := c.pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	_, err = tx.Exec(context.Background(),
		`INSERT INTO assets(address, symbol, class)
		VALUES ($1, $2, $3) ON CONFLICT (address) DO NOTHING`, base.Address, base.Symbol, base.Class)
	if err != nil {
		tx.Rollback(context.Background())
		return err
	}
	_, err = tx.Exec(context.Background(),
		`INSERT INTO assets(address, symbol, class)
			VALUES ($1, $2, $3) ON CONFLICT (address) DO NOTHING`, quote.Address, quote.Symbol, quote.Class)
	if err != nil {
		tx.Rollback(context.Background())
		return err
	}
	_, err = tx.Exec(context.Background(),
		`INSERT INTO markets (address, base_address, quote_address, recorded_at)
		VALUES ($1, $2, $3, $4)`, marketAddress, base.Address, quote.Address, time.Now().UTC())
	if err != nil {
		tx.Rollback(context.Background())
		return err
	}
	tx.Commit(context.Background())
	return nil
}

// GetMarkets returns all markets from the database
func (c *Connection) GetMarkets() ([]*model.MarketInfo, error) {
	var markets = make([]*model.MarketInfo, 0)
	q := `
select m.address, m.recorded_at,
b.symbol bs, b.address ba, b.class bt,
q.symbol qs, q.address qa, q.class qt
from markets m join assets b on (m.base_address = b.address)
join assets q on (m.quote_address = q.address)
order by m.recorded_at desc`
	rows, err := c.pool.Query(context.Background(), q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var market model.MarketInfo
		if err := rows.Scan(
			&market.Address, &market.RecordedAt,
			&market.Base.Symbol, &market.Base.Address, &market.Base.Class,
			&market.Quote.Symbol, &market.Quote.Address, &market.Quote.Class,
		); err != nil {
			return nil, err
		}
		markets = append(markets, &market)
	}
	return markets, nil
}

// GetMarketByAddress returns a market from the database by its address
func (c *Connection) GetMarketByAddress(address string) (*model.MarketInfo, error) {
	var market model.MarketInfo
	q := `
select m.address, m.recorded_at,
b.symbol bs, b.address ba, b.class bt,
q.symbol qs, q.address qa, q.class qt
from markets m join assets b on (m.base_address = b.address)
join assets q on (m.quote_address = q.address)
where m.address = $1`
	err := c.pool.QueryRow(context.Background(), q, address).Scan(
		&market.Address, &market.RecordedAt,
		&market.Base.Symbol, &market.Base.Address, &market.Base.Class,
		&market.Quote.Symbol, &market.Quote.Address, &market.Quote.Class,
	)
	if err != nil {
		return nil, err
	}
	return &market, nil
}

func (c *Connection) IsAuthorized(address string) (active bool) {
	if err := c.pool.QueryRow(context.Background(), "select active from accounts where address = $1", address).Scan(&active); err != nil {
		active = false
	}
	return
}

// GetTokenAddresses returns the list of tokens addresses currently in the database
func (c *Connection) GetAssetAddressesByClass(class string) ([]string, error) {
	rows, err := c.pool.Query(context.Background(), "SELECT address FROM assets WHERE class = $1", class)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var addresses []string
	for rows.Next() {
		var address string
		err = rows.Scan(&address)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, address)
	}
	return addresses, nil
}

// ValidateOrder checks if an order is valid
// by checking if the market exists and if the account
// has enough balance to place the order
func (c *Connection) ValidateOrder(order *model.Order, from string, quote decimal.Decimal) error {
	market, err := c.GetMarketByAddress(order.Market)
	if err != nil {
		return fmt.Errorf("market not found")
	}

	var price decimal.Decimal
	if quote.IsZero() { // it's a limit order, calculate the total amount
		// it's a limit order, calculate the total amount
		size := decimal.NewFromInt(int64(order.Size))
		price, err := helpers.ParseAmount(order.Price)
		if err != nil {
			return fmt.Errorf("invalid price")
		}
		// calculate the total amount
		quote = price.Mul(size)
	}

	// open a transaction
	tx, err := c.pool.Begin(context.Background())
	if err != nil {
		return err
	}

	var (
		newBalance  decimal.Decimal
		targetAsset string
	)

	switch order.Side {
	case model.SideBid:
		targetAsset = market.Base.Address
	case model.SideAsk:
		targetAsset = market.Quote.Address
	default:
		return fmt.Errorf("invalid order side")
	}

	q := `UPDATE balances SET balance = balance - $1 WHERE address = $2 AND asset_address = $3 returning balance`
	err = tx.QueryRow(context.Background(), q, quote, from, targetAsset).Scan(&newBalance)
	if err != nil {
		tx.Rollback(context.Background())
		return err
	}
	if newBalance.IsNegative() {
		tx.Rollback(context.Background())
		return fmt.Errorf("insufficient %s balance", targetAsset)
	}

	// populate the order ID and RecordedAt
	order.ID = uuid.New().String()

	// if all is good insert the order
	q = `INSERT INTO orders (id, market_address, from_address, side, price, size, recorded_at, submitted_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err = tx.Exec(context.Background(), q, order.ID, market.Address, from, order.Side, price, order.Size, order.RecordedAt, order.SubmittedAt)
	if err != nil {
		tx.Rollback(context.Background())
		return err
	}

	tx.Commit(context.Background())
	return err
}

func (c *Connection) GetBalance(address, token string) (decimal.Decimal, error) {
	var b decimal.Decimal
	err := c.pool.QueryRow(context.Background(), "SELECT balance FROM balances WHERE address = $1 AND asset_address = $2", address, token).Scan(&b)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return b, nil
	}
	return b, err
}

// GetOrder returns an order from the database
func (c *Connection) GetOrder(id string) (order *model.Order, from, status string, err error) {
	q := `
	SELECT o.id, o.from_address, o.market_address, o.side, o.price, o.size, o.recorded_at, o.submitted_at,
	COALESCE(m.status, 'open') AS status
	FROM "orders" o
	LEFT JOIN (
		SELECT order_id, status
		FROM "matches"
		WHERE status = any($2)
	) m ON m.order_id = o.id
	WHERE o.id = $1
	`
	order = new(model.Order)
	var price decimal.Decimal
	err = c.pool.QueryRow(context.Background(), q, id, model.ClosedStatuses).Scan(
		&order.ID, &from, &order.Market, &order.Side, &price, &order.Size, &order.RecordedAt, &order.SubmittedAt, &status,
	)
	if err != nil {
		return
	}
	order.Price = price.String()
	return
}

// GetMarketPrice returns the current market price
// it uses the VWAP (Volume Weighted Average Price) formula
func (c *Connection) GetMarketPrice(market string) (price decimal.Decimal, err error) {
	qVWAP := `
	SELECT sum(m.price * m.size)::numeric / sum(m.size)::numeric as vwap
	FROM matches m JOIN orders o ON m.order_id = o.id
	WHERE m.status = 'filled' AND o.market_address = $1`
	err = c.pool.QueryRow(context.Background(), qVWAP, market).Scan(&price)
	return price, err
}
