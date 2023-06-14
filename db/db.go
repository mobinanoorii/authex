package db

import (
	"authex/model"
	"context"
	_ "embed"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/gommon/log"
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
			select {
			case t, ok := <-c.Transfers:
				if !ok {
					wg.Done()
					log.Debugf("closing balance handler")
				}
				tx, err := c.pool.BeginTx(context.Background(), pgx.TxOptions{})
				if err != nil {
					log.Errorf("error starting transaction: %v", err)
				}
				q := `INSERT INTO balances (address, token_address, balance) VALUES ($1, $2, $3) ON CONFLICT (address, token_address) DO UPDATE SET balance = balances.balance + $3`
				for _, delta := range t.Deltas {
					if _, err := tx.Exec(context.Background(), q, delta.Address, t.TokenAddress, delta.Amount); err != nil {
						log.Errorf("error updating the recipient balance: %v", err)
						tx.Rollback(context.Background())
						break
					}
				}
				// update token block number
				if _, err := tx.Exec(context.Background(), "UPDATE tokens SET last_block = $1 WHERE address = $2", t.BlockNumber, t.TokenAddress); err != nil {
					log.Errorf("error updating the token block number: %v", err)
					tx.Rollback(context.Background())
					break
				}

				tx.Commit(context.Background())
			}
		}
	}()

	// handle CLOB matches
	go func() {
		for {
			select {
			case match, ok := <-c.Matches:
				if !ok {
					wg.Done()
					log.Debugf("closing match handler")
				}
				c.handleMatch(match)
			}
		}
	}()

	wg.Wait()

}

func (c *Connection) handleMatch(m *model.Match) {
	// NOW UPDATE THE DATABASE
	_, err := c.pool.Exec(context.Background(), "INSERT INTO orders (id, symbol, side, price, size) VALUES ($1, $2, $3, $4, $5)",
		m.Request.Payload.ID, m.Request.Payload.Market, m.Request.Payload.Side, m.Request.Payload.Price, m.Request.Payload.Size)
	if err != nil {
		return
	}
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
func (c *Connection) SaveMarket(marketID string, base, quote *model.Token) error {
	tx, err := c.pool.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	_, err = tx.Exec(context.Background(),
		`INSERT INTO tokens(address, symbol, asset_type)
		VALUES ($1, $2, $3) ON CONFLICT (address) DO NOTHING`, base.Address, base.Symbol, base.AssetType)
	if err != nil {
		tx.Rollback(context.Background())
		return err
	}
	_, err = tx.Exec(context.Background(),
		`INSERT INTO tokens(address, symbol, asset_type)
			VALUES ($1, $2, $3) ON CONFLICT (address) DO NOTHING`, quote.Address, quote.Symbol, quote.AssetType)
	if err != nil {
		tx.Rollback(context.Background())
		return err
	}
	_, err = tx.Exec(context.Background(),
		`INSERT INTO markets (id, base_address, quote_address, recorded_at)
		VALUES ($1, $2, $3, $4)`, marketID, base.Address, quote.Address, time.Now().UTC())
	if err != nil {
		tx.Rollback(context.Background())
		return err
	}
	tx.Commit(context.Background())
	return nil
}

func (c *Connection) IsAuthorized(address string) (active bool) {
	if err := c.pool.QueryRow(context.Background(), "select active from accounts where address = $1", address).Scan(&active); err != nil {
		active = false
	}
	return
}

// GetTokenAddresses returns the list of tokens addresses currently in the database
func (c *Connection) GetTokenAddresses() ([]string, error) {
	rows, err := c.pool.Query(context.Background(), "SELECT address FROM tokens")
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

// GetOrder returns an order from the database
func (c *Connection) GetOrder(id string) (*model.Order, error) {
	var order model.Order
	err := c.pool.QueryRow(context.Background(), "SELECT id, symbol, side, price, size FROM orders WHERE id = $1", id).Scan(&order.ID, &order.Market, &order.Side, &order.Price, &order.Size)
	if err != nil {
		return nil, err
	}
	return &order, nil
}
