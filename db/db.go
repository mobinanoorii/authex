package db

import (
	"authex/helpers"
	"authex/model"
	"context"
	_ "embed"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var dbSchema string

type Connection struct {
	pool    *pgxpool.Pool
	Matches chan *model.Match
}

// Close the connection and all channels
func (c *Connection) Close() {
	close(c.Matches)
	c.pool.Close()
}

func NewConnection(dbUrl string) (*Connection, error) {
	pool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		return nil, err
	}
	return &Connection{
		pool:    pool,
		Matches: make(chan *model.Match),
	}, nil
}

func (c *Connection) Run() {
	for {
		select {
		case match := <-c.Matches:
			c.handleMatch(match)
		default:
			// channel is closed
			return
		}
	}
}

func (c *Connection) handleMatch(m *model.Match) {

	// NOW UPDATE THE DATABASE
	_, err := c.pool.Exec(context.Background(), "INSERT INTO orders (id, symbol, side, price, size) VALUES ($1, $2, $3, $4, $5)",
		m.OrderRequest.Order.ID, m.OrderRequest.Order.Market, m.OrderRequest.Order.Side, m.OrderRequest.Order.Price, m.OrderRequest.Order.Size)
	if err != nil {
		return
	}
}

// Setup the database, open a connection and create the database schema
func Setup(dbUrl string) error {
	conn, err := pgx.Connect(context.Background(), dbUrl)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())
	_, err = conn.Exec(context.Background(), dbSchema)
	return err
}

// SaveMarket saves a market to the database
func (c *Connection) SaveMarket(m *model.Market) error {
	_, err := c.pool.Exec(context.Background(),
		`INSERT INTO markets (id, base_symbol, quote_symbol, base_address, quote_address)
		VALUES ($1, $2, $3, $4, $5)`, m.String(), m.BaseSymbol, m.QuoteSymbol, m.BaseAddress, m.QuoteAddress)
	if err != nil {
		return err
	}
	// save the base token
	if !helpers.IsEmpty(m.BaseAddress) {
		_, err = c.pool.Exec(context.Background(),
			`INSERT INTO tokens(address, symbol)
			VALUES ($1, $2) ON CONFLICT (address) DO NOTHING`, m.BaseAddress, m.BaseSymbol)
		if err != nil {
			return err
		}
	}
	// save the quote token
	if !helpers.IsEmpty(m.QuoteAddress) {
		_, err = c.pool.Exec(context.Background(),
			`INSERT INTO tokens(address, symbol)
			VALUES ($1, $2) ON CONFLICT (address) DO NOTHING`, m.QuoteAddress, m.QuoteSymbol)
		if err != nil {
			return err
		}
	}
	return nil
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
