package db

import (
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
