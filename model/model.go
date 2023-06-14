package model

import (
	"authex/helpers"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// Asset is the asset type
const (
	AssetOffChain = "offchain"
	AssetERC20    = "erc20"
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
		// Permissioned if the system shall be closed to authorized accounts
		Permissioned bool
	}
}

// SignedRequest is a generic request with a signature
type SignedRequest[T Serializable] struct {
	Signature string `json:"signature,omitempty"`
	From      string `json:"-"`
	Payload   T      `json:"payload,omitempty"`
}

// SignatureBytes returns the signature bytes
func (r *SignedRequest[T]) SignatureBytes() ([]byte, error) {
	return hex.DecodeString(r.Signature)
}

type Serializable interface {
	Serialize() ([]byte, error)
}

const (
	// SideBid is the bid side
	SideBid string = "bid"
	// SideAsk is the ask side
	SideAsk string = "ask"
	// CancelOrder is used in internally as side to cancel an order
	CancelOrder string = "del"
)

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

func (o Order) Serialize() ([]byte, error) {
	return json.Marshal(o)
}

func (o Order) Validate() error {
	if o.Market == "" {
		return fmt.Errorf("market must be set")
	}
	if o.Side != SideBid && o.Side != SideAsk {
		return fmt.Errorf("side is either bid or ask, got %s", o.Side)
	}
	if o.Size <= 0 {
		return fmt.Errorf("size must be positive, got %d", o.Size)
	}
	if o.ID != "" {
		return fmt.Errorf("the order ID must not be set as it's assigned by the exchange")
	}
	return nil
}

// Match is the result of a match between two orders
type Match struct {
	Request *SignedRequest[Order] `json:"order_request,omitempty"`
	IDs     []string              `json:"id,omitempty"`
	Prices  []decimal.Decimal     `json:"price,omitempty"`
}

func NewMatch(o *SignedRequest[Order], ids []string, prices []decimal.Decimal) *Match {
	return &Match{
		Request: o,
		IDs:     ids,
		Prices:  prices,
	}
}

// Token is the token of the exchange
type Token struct {
	// Symbol is the symbol of the token
	Symbol string `json:"symbol,omitempty"`
	// Address is the address of the token
	// If the token is an ERC20 token, it's the address of the token contract
	// If the token is an off-chain token, it's the hash of the token symbol
	Address string `json:"address,omitempty"`
	// AssetType is the type of the asset
	// it will be either "erc20" or "offchain"
	AssetType string `json:"asset_type,omitempty"`
}

func (t Token) String() string {
	return fmt.Sprintf("%s:%s", t.Symbol, t.Address)
}

// IsERC20 returns true if the token is an ERC20 token
func (t *Token) IsERC20() bool {
	return t.AssetType == AssetERC20
}

// NewToken is a helper function to create a new token
func NewToken(symbol, address string, assetType string) *Token {
	return &Token{
		Symbol:    symbol,
		Address:   address,
		AssetType: assetType,
	}
}

// NewERC20Token is a helper function to create a new ERC20 token
func NewERC20Token(symbol, address string) *Token {
	return &Token{
		Symbol:    symbol,
		Address:   address,
		AssetType: AssetERC20,
	}
}

// NewOffChainAsset is a helper function to create a new off-chain token
func NewOffChainAsset(symbol string) *Token {
	return &Token{
		Symbol:    symbol,
		Address:   helpers.AsAddress(strings.ToLower(symbol)),
		AssetType: AssetOffChain,
	}
}

// Market is the market of the exchange
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

func (m Market) Serialize() ([]byte, error) {
	return json.Marshal(m)
}

type BalanceDelta struct {
	// Address is the address of the account
	Address string `json:"address,omitempty"`
	// Amount is the amount of the change
	Amount *big.Int `json:"amount,omitempty"`
}

func (bd BalanceDelta) String() string {
	return fmt.Sprintf("%s: %s", bd.Address, bd.Amount.String())
}

func NewBalanceDelta(address string, amount *big.Int) *BalanceDelta {
	return &BalanceDelta{
		Address: address,
		Amount:  amount,
	}
}

type BalanceChange struct {
	// BlockNumber is the block number of the transfer
	BlockNumber uint64 `json:"block_number,omitempty"`
	// TokenAddress is the address of the token
	TokenAddress string `json:"token_address,omitempty"`
	// Balances lists the balance updates
	Deltas []*BalanceDelta `json:"deltas,omitempty"`
}
