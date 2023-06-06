package model

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// Settings is the global configuration
type Settings struct {
	// Version is the version of the application
	Version string
	// Database connection string
	DB struct {
		// URI is the connection string for the postgres database
		// e.g. "postgres://user:password@localhost:5432/authex?sslmode=disable"
		URI string
		// MaxConnections is the maximum number of connections to the database
		// 0 means unlimited
		// TODO: unused
		MaxConnections int
	}
	// Network is the configuration for an Ethereum compatible network
	Network struct {
		// RPCEndpoint is the URL of the RPC endpoint
		RPCEndpoint string
		// WSSEndpoint is the URL of the Websocket endpoint to listen for events
		WSEndpoint string
		// ChainID is the ID of the target chain
		ChainID string
	}
	// Identity is the configuration for server on chain related identities
	Identity struct {
		// KeystorePath is the path to the local keystore directory
		KeystorePath string
		// KeyFile is the name of the encrypted private key file to import
		KeyFile string
		// Password is the password for the key file
		Password string
		// AccessContractAddress is the address of the access control contract
		AccessContractAddress string
	}
	// Web is the configuration for the web server
	Web struct {
		// ListenAddr is the address to listen for incoming connections
		ListenAddr string
		// RateLimit is the number of requests per second
		RateLimit int
		// BurstLimit is the number of requests that can be made in a burst
		BurstLimit int
		// Remember is the duration of time to remember a client
		Remember time.Duration
	}
}

type SignedMessage interface {
	// Signature is the signature of the order
	GetSignature() ([]byte, error)
	// From is the address of the client, populated by the server
	GetData() ([]byte, error)
	// GetFrom returns the address of the client
	GetFrom() string
	// SetFrom sets the address of the client
	SetFrom(address string)
}

type SignedRequest struct {
	// Signature is the signature of the order
	Signature string `json:"signature,omitempty"`
	// From is the address of the client, populated by the server
	From string `json:"-"`
}

func (r SignedRequest) GetSignature() ([]byte, error) {
	return hex.DecodeString(r.Signature)
}

func (r SignedRequest) GetData() ([]byte, error) {
	// serialize the order
	return json.Marshal(map[string]struct{}{})
}

func (r SignedRequest) SetFrom(address string) {
	r.From = address
}

func (r SignedRequest) GetFrom() string {
	return r.From
}

const (
	// SideBid is the bid side
	SideBid string = "bid"
	// SideAsk is the ask side
	SideAsk string = "ask"
	// CancelOrder is used in internally as side to cancel an order
	CancelOrder string = "del"
)

// OrderRequest is the request to place an order
// it contains the signature of the order and the order itself
// The order is signed by the client and the signature is used by the server
// to extract the client public key, calculate the client address and verify
// that it's part of the games
type OrderRequest struct {
	SignedRequest
	// Order is the order itself
	Order Order `json:"order,omitempty"`
}

// Order is the CLOB order
type Order struct {
	// ID is UUID of the order, populated by the server
	ID string `json:"id,omitempty"`
	// SubmittedAt is the time the order was submitted, populated by the client
	SubmittedAt time.Time `json:"submitted_at,omitempty"`
	// RecordedAt is the time the order was received, populated by the server
	RecordedAt time.Time `json:"recorded_at,omitempty"`
	// Market is the market of the order, in the form of "base/quote" e.g. "USD/ETH"
	// This should be something that is compatible with the trading pair supported by the exchange
	Market string `json:"market,omitempty"`
	// Size is the size of the order
	Size uint `json:"size,omitempty"`
	// Price is the price of the order, in the quote currency. If not specified, it's a market order
	Price string `json:"price,omitempty"`
	// Side is the side of the order, either "bid" or "ask"
	Side string `json:"side,omitempty"`
}

func (or OrderRequest) GetMessageData() ([]byte, error) {
	// serialize the order
	return json.Marshal(or.Order)
}

func (or OrderRequest) Validate() error {
	if or.Order.Market == "" {
		return fmt.Errorf("market must be set")
	}
	if or.Order.Side != SideBid && or.Order.Side != SideAsk {
		return fmt.Errorf("side is either bid or ask, got %s", or.Order.Side)
	}
	if or.Order.Size <= 0 {
		return fmt.Errorf("size must be positive, got %d", or.Order.Size)
	}
	if or.Order.ID != "" {
		return fmt.Errorf("the order ID must not be set as it's assigned by the exchange")
	}
	return nil
}

// Match is the result of a match between two orders
type Match struct {
	OrderRequest *OrderRequest     `json:"order_request,omitempty"`
	IDs          []string          `json:"id,omitempty"`
	Prices       []decimal.Decimal `json:"price,omitempty"`
}

func NewMatch(o *OrderRequest, ids []string, prices []decimal.Decimal) *Match {
	return &Match{
		OrderRequest: o,
		IDs:          ids,
		Prices:       prices,
	}
}

type Market struct {
	// BaseSymbol is the base currency of the market
	BaseSymbol string `json:"base,omitempty"`
	// BaseAddress is the ERC20 address of the base currency
	// if empty, it's assumed to be an off-chain asset
	BaseAddress string `json:"base_address,omitempty"`
	// QuoteSymbol is the quote currency of the market
	// if empty, it's assumed to be an off-chain asset
	QuoteSymbol string `json:"quote,omitempty"`
	// QuoteAddress is the ERC20 address of the quote currency
	QuoteAddress string `json:"quote_address,omitempty"`
}

func (m Market) String() string {
	return fmt.Sprintf("%s/%s", m.BaseSymbol, m.QuoteSymbol)
}

// CreateMarketRequest is the request to create a new market
type CreateMarketRequest struct {
	SignedRequest
	// Market is the market to create
	Market Market `json:"market,omitempty"`
}

func (r CreateMarketRequest) GetMessageData() ([]byte, error) {
	// serialize the order
	return json.Marshal(r.Market)
}
