package web

import (
	"bytes"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"authex/clob"
	"authex/db"
	h "authex/helpers"
	"authex/model"
	"authex/network"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/shopspring/decimal"
)

//go:embed index.html
var indexHTML string

// AuthexServer is the main server for the CLOB
type AuthexServer struct {
	opts    *model.Settings
	echo    *echo.Echo
	runtime *Runtime
	clobCli *clob.Pool
	nodeCli *network.NodeClient
	dbCli   *db.Connection
}

// Endpoint is a REST endpoint
type Endpoint struct {
	Help   string
	Path   string
	Routes []Route
}

type Route struct {
	Method  string
	Path    string
	Help    string
	Handler func(c echo.Context) error
}

const (
	keyRequestID = "request-id"
	keyOrderID   = "order-id"
	valError     = "error"
	valSuccess   = "ok"
)

// NewAuthexServer creates a new CLOB server
func NewAuthexServer(opts *model.Settings, clobCli *clob.Pool, nodeCli *network.NodeClient, dbCli *db.Connection) (AuthexServer, error) {
	var err error

	r := AuthexServer{
		opts:    opts,
		clobCli: clobCli,
		nodeCli: nodeCli,
		dbCli:   dbCli,
	}
	r.echo = echo.New()
	r.echo.HideBanner = true
	r.echo.Use(middleware.Logger())
	r.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	// assign a unique request ID to each request
	r.echo.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(keyRequestID, h.IID())
			return next(c)
		}
	})
	r.echo.Logger.SetOutput(os.Stdout)
	r.echo.Logger.SetLevel(log.INFO)
	r.echo.Logger.Infof("starting authex server on chain %s@%s", opts.Network.ChainID, opts.Network.RPCEndpoint)

	// start listening for events

	r.runtime = NewRuntime(opts.Version)

	groups := []Endpoint{
		{
			Help: "Administration endpoints",
			Path: "/admin",
			Routes: []Route{
				{
					Path:    "/markets",
					Method:  http.MethodPost,
					Handler: r.registerMarket,
					Help:    "Register a new market (requires admin privileges)",
				},
				{
					Path:    "/accounts/fund",
					Method:  http.MethodPost,
					Handler: r.fund,
					Help:    "Fund an account (requires admin privileges)",
				},
				{
					Path:    "/accounts/allow",
					Method:  http.MethodPost,
					Handler: r.fund,
					Help:    "Add an account to the allowed list (requires admin privileges)",
				},
				{
					Path:    "/accounts/block",
					Method:  http.MethodPost,
					Handler: r.fund,
					Help:    "Remove an account from the allowed list (requires admin privileges)",
				},
			},
		},
		{
			Help: "Query endpoints",
			Path: "/query",
			Routes: []Route{
				{
					Path:    "/markets",
					Method:  http.MethodGet,
					Handler: r.getMarkets,
					Help:    "Get all markets",
				},
				{
					Path:    "/markets/:address",
					Method:  http.MethodGet,
					Handler: r.getMarketByAddress,
					Help:    "Get a market by address",
				},
				{
					Path:    "/markets/:address/quote/:side/:size",
					Method:  http.MethodGet,
					Handler: r.getMarketQuote,
					Help:    "Get a market quote",
				},
				{
					Path:    "/orders/:id",
					Method:  http.MethodGet,
					Handler: r.getOrder,
					Help:    "Get an order by id",
				},
				{
					Path:    "/markets/:address/price",
					Method:  http.MethodGet,
					Handler: r.getMarketPrice,
					Help:    "Get all orders",
				},
			},
		},
		{
			Help: "Account endpoints",
			Path: "/account",
			Routes: []Route{
				{

					Path:    "/orders",
					Method:  http.MethodPost,
					Handler: r.postOrder,
					Help:    "Post a new buy or sell order",
				},
				{
					Path:    "/orders/cancel",
					Method:  http.MethodPost,
					Handler: r.cancelOrder,
					Help:    "Cancel an order",
				},
				{
					Path:    "/withdraw",
					Method:  http.MethodPost,
					Handler: r.withdraw,
					Help:    "Withdraw funds from the CLOB",
				},
				{
					Path:    "/orders/:id",
					Method:  http.MethodGet,
					Handler: r.getOrder,
					Help:    "Get an order by id",
				},
				{
					Path:    "/:address/orders",
					Method:  http.MethodGet,
					Handler: r.getOrder,
					Help:    "Get all orders for an account",
				},
				{
					Path:    "/:address/balance/:symbol",
					Method:  http.MethodGet,
					Handler: r.getOrder,
					Help:    "Get the balance of an account for a symbol",
				},
			},
		},
	}

	indexTemplate, err := template.New("index").Parse(indexHTML)
	if err != nil {
		return r, err
	}

	r.echo.GET("/", func(c echo.Context) error {
		return index(c, indexTemplate, r.runtime, groups)
	})

	for _, group := range groups {
		g := r.echo.Group(group.Path)
		for _, endpoint := range group.Routes {
			g.Match([]string{endpoint.Method}, endpoint.Path, endpoint.Handler)
		}
	}

	return r, nil
}

// Start the server
func (r AuthexServer) Start() error {
	r.echo.Logger.Infof("listening on %s", r.opts.Web.ListenAddr)
	return r.echo.Start(r.opts.Web.ListenAddr)
}

func withMsg(msg string) func(map[string]any) {
	return func(m map[string]any) {
		m["message"] = msg
	}
}

func withData(key string, data any) func(map[string]any) {
	return func(m map[string]any) {
		m[key] = data
	}
}

func rsp(reqID, status string, opts ...func(map[string]any)) map[string]any {
	data := map[string]any{}
	data["status"] = status
	data[keyRequestID] = reqID
	for _, o := range opts {
		o(data)
	}
	return data
}

func ok(reqID string, opts ...func(map[string]any)) map[string]any {
	return rsp(reqID, valSuccess, opts...)
}

func er(reqID, msg string, opts ...func(map[string]any)) map[string]any {
	return rsp(reqID, valError, append(opts, withMsg(msg))...)
}

// extractAddress verifies the signature of an order
// https://goethereumbook.org/signature-verify/
// // TODO: the signature verification may be weak since tbe messages can be replayed
// probably would be good to require a time stamp in the message and verify that is not older than a few seconds
func extractAddress(signature string, payload model.Serializable) (address string, err error) {
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return
	}
	// serialize the msg
	msgData, err := payload.Serialize()
	if err != nil {
		return
	}
	// verify the signature
	hash := crypto.Keccak256Hash(msgData)
	sigPublicKeyECDSA, err := crypto.SigToPub(hash.Bytes(), sigBytes)
	if err != nil {
		return

	}
	address = crypto.PubkeyToAddress(*sigPublicKeyECDSA).Hex()
	return
}

func (r AuthexServer) isAuthorized(address string) (err error) {
	if !r.opts.Web.Permissioned {
		return nil
	}
	if is := r.dbCli.IsAuthorized(address); !is {
		err = fmt.Errorf("account %s is not authorized", address)
	}
	return
}

func (r AuthexServer) isAdmin(c echo.Context, signature string, payload model.Serializable) error {
	requestID := c.Get(keyRequestID).(string)
	sender, err := extractAddress(signature, payload)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, er(requestID, err.Error()))
	}
	isAdmin, err := r.nodeCli.IsAdmin(sender)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, er(requestID, err.Error()))
	}
	if !isAdmin {
		return c.JSON(http.StatusUnauthorized, er(requestID, "only the admin can perform this action"))
	}
	return nil
}

func (r AuthexServer) registerMarket(c echo.Context) error {
	// generate a new request id to be used in logging
	requestID := c.Get(keyRequestID).(string)

	cmr := &model.SignedRequest[model.Market]{}
	if err := c.Bind(cmr); err != nil {
		log.Errorf("error binding request: %v, [incident: %s]", err, requestID)
		return c.JSON(http.StatusBadRequest, er(requestID, "invalid market request"))
	}
	// admin only
	if err := r.isAdmin(c, cmr.Signature, cmr.Payload); err != nil {
		log.Errorf("error registering market: %v, [incident: %s]", err, requestID)
		return c.JSON(http.StatusUnauthorized, er(requestID, "unauthorized"))
	}

	// helper function to parse a token
	parseToken := func(symbol string, address string) (token *model.Asset, err error) {
		if h.IsEmpty(symbol) {
			err = fmt.Errorf("missing base or quote symbol")
			return
		}
		// set the base and quote tokens
		token = model.NewOffChainAsset(symbol)
		if !h.IsEmpty(address) {
			token = model.NewERC20Token(symbol, address)
			isERC20, errERC := r.nodeCli.IsERC20(token.Address)
			if errERC != nil {
				err = fmt.Errorf("error checking if %s is an ERC20 token: %s", token.Address, errERC.Error())
				return
			}
			if !isERC20 {
				err = fmt.Errorf("%s is not an ERC20 token", token.Address)
				return
			}
		}
		return
	}

	// set the base and quote tokens
	base, err := parseToken(cmr.Payload.BaseSymbol, cmr.Payload.BaseAddress)
	if err != nil {
		log.Errorf("error parsing base token: %v, [incident: %s]", err, requestID)
		return c.JSON(http.StatusBadRequest, er(requestID, err.Error()))
	}
	quote, err := parseToken(cmr.Payload.QuoteSymbol, cmr.Payload.QuoteAddress)
	if err != nil {
		log.Errorf("error parsing quote token: %v, [incident: %s]", err, requestID)
		return c.JSON(http.StatusBadRequest, er(requestID, err.Error()))
	}

	// compute the market address
	marketAddr, err := h.ComputeMarketAddress(base.Address, quote.Address)
	if err != nil {
		log.Errorf("error computing market address: %s [incident: %s]", err.Error(), requestID)
		return c.JSON(http.StatusBadRequest, er(requestID, "invalid addresses for base or quote"))
	}
	log.Infof("new market address: %s", marketAddr)
	if err := r.dbCli.SaveMarket(marketAddr, base, quote); err != nil {
		log.Errorf("error saving market: %s [incident: %s]", err.Error(), requestID)
		return c.JSON(http.StatusInternalServerError, er(requestID, "error saving market"))
	}
	// open the market
	r.clobCli.OpenMarket(marketAddr)
	// start listening
	if base.IsERC20() {
		r.nodeCli.Tokens <- base.Address
	}
	if quote.IsERC20() {
		r.nodeCli.Tokens <- quote.Address
	}
	// Only the admin can register a new market
	return c.JSON(http.StatusOK, ok(requestID, withData("address", marketAddr)))
}

// getMarket returns the market details for a given market address
func (r AuthexServer) getMarketByAddress(c echo.Context) error {
	requestID := c.Get(keyRequestID).(string)
	marketID := c.Param("address")
	m, err := r.dbCli.GetMarketByAddress(marketID)
	if err != nil {
		log.Errorf("error getting market: %s [incident: %s]", err.Error(), requestID)
		return c.JSON(http.StatusNotFound, er(requestID, "market not found"))
	}
	// TODO: add the market dept
	m.OrderBook = r.clobCli.GetOrderBook(marketID)
	return c.JSON(http.StatusOK, ok(requestID, withData("market", m)))
}

// getMarkets returns the list of markets
func (r AuthexServer) getMarkets(c echo.Context) error {
	requestID := c.Get(keyRequestID).(string)
	markets, err := r.dbCli.GetMarkets()
	if err != nil {
		log.Errorf("error getting markets: %s [incident: %s]", err.Error(), requestID)
		return c.JSON(http.StatusInternalServerError, er(requestID, "error getting markets"))
	}
	return c.JSON(http.StatusOK, ok(requestID, withData("markets", markets)))
}

// fund adds funds to the an account
func (r AuthexServer) fund(c echo.Context) error {
	// generate a new request id to be used in logging
	requestID := c.Get(keyRequestID).(string)
	req := &model.SignedRequest[model.Funding]{}
	if err := c.Bind(req); err != nil {
		log.Errorf("error binding request: %v, [incident: %s]", err, requestID)
		return c.JSON(http.StatusBadRequest, er(requestID, "invalid funding request"))
	}
	// admin only
	if err := r.isAdmin(c, req.Signature, req.Payload); err != nil {
		log.Errorf("error funding account: %v, [incident: %s]", err, requestID)
		return c.JSON(http.StatusUnauthorized, er(requestID, "unauthorized"))
	}
	// extract the address from the signature

	amount, err := h.ParseAmount(req.Payload.Amount)
	if err != nil {
		log.Errorf("error parsing amount: %v, [incident: %s]", err, requestID)
		return c.JSON(http.StatusBadRequest, er(requestID, "invalid amount"))
	}
	bc := &model.BalanceChange{
		TokenAddress: req.Payload.Asset,
		Deltas: []*model.BalanceDelta{
			model.NewBalanceDelta(req.Payload.Account, amount),
		},
	}
	r.dbCli.Transfers <- bc
	return c.JSON(http.StatusOK, ok(requestID, withMsg("scheduled")))
}

// postOrder submits a new order to the CLOB
// it is required that the order is signed by the account
// the fields ID and RecordedAt are overwritten by the server
func (r AuthexServer) postOrder(c echo.Context) error {
	// generate a new request id to be used in logging
	requestID := c.Get(keyRequestID).(string)
	req := &model.SignedRequest[model.Order]{}
	if err := c.Bind(req); err != nil {
		log.Errorf("error binding request: %v, [incident: %s]", err, requestID)
		return c.JSON(http.StatusBadRequest, er(requestID, "invalid order request"))
	}
	// extract the address from the signature
	sender, err := extractAddress(req.Signature, req.Payload)
	if err != nil {
		log.Errorf("error extracting account address: %v, [incident: %s]", err, requestID)
		return c.JSON(http.StatusUnauthorized, er(requestID, "error extracting account address"))
	}
	if err := r.isAuthorized(sender); err != nil {
		log.Errorf("error authorizing address: %v, [incident: %s]", err, requestID)
		return c.JSON(http.StatusUnauthorized, er(requestID, "unauthorized"))
	}
	// update the from address
	req.From = sender
	// verify that is not older than a few seconds
	req.Payload.RecordedAt = time.Now().UTC()
	if req.Payload.SubmittedAt.IsZero() {
		req.Payload.SubmittedAt = req.Payload.RecordedAt
	}
	if req.Payload.RecordedAt.Sub(req.Payload.SubmittedAt) > 2*time.Second {
		err := fmt.Errorf("order is older than 2 seconds")
		log.Errorf("error validating order: %v, [incident: %s]", err, requestID)
		return c.JSON(http.StatusBadRequest, er(requestID, "order is older than 2 seconds"))
	}

	// validate the order
	if err := req.Payload.Validate(); err != nil {
		log.Errorf("error validating order: %v, [incident: %s]", err, requestID)
		return c.JSON(http.StatusBadRequest, er(requestID, "invalid order"))
	}
	// it's a market order or a limit order?
	var quote decimal.Decimal
	if h.IsEmpty(req.Payload.Price) {
		size := decimal.NewFromInt(int64(req.Payload.Size))
		quote, err = r.clobCli.GetQuote(req.Payload.Market, req.Payload.Side, size)
		if err != nil {
			log.Errorf("error getting quote: %v, [incident: %s]", err, requestID)
			return c.JSON(http.StatusBadRequest, er(requestID, "order cannot be processed"))
		}
	}
	// TODO this modifies the order (assign the ID), refactor
	if err := r.dbCli.ValidateOrder(&req.Payload, sender, quote); err != nil {
		log.Errorf("error validating order on db: %v, [incident: %s]", err, requestID)
		return c.JSON(http.StatusBadRequest, er(requestID, "invalid order"))
	}
	// queue the order for processing
	r.clobCli.Inbound <- req
	// reply with the order
	return c.JSON(http.StatusOK, ok(requestID, withData(keyOrderID, req.Payload.ID)))
}

func (r AuthexServer) cancelOrder(c echo.Context) error {
	requestID := c.Get(keyRequestID).(string)
	req := &model.SignedRequest[model.Order]{}
	if err := c.Bind(req); err != nil {
		log.Errorf("error binding request: %v, [incident: %s]", err, requestID)
		return c.JSON(http.StatusBadRequest, er(requestID, "invalid order request"))
	}
	// extract the address from the signature
	sender, err := extractAddress(req.Signature, req.Payload)
	if err != nil {
		log.Errorf("error extracting account address: %v, [incident: %s]", err, requestID)
		return c.JSON(http.StatusUnauthorized, er(requestID, "error extracting account address"))
	}
	if err := r.isAuthorized(sender); err != nil {
		log.Errorf("error authorizing address: %v, [incident: %s]", err, requestID)
		return c.JSON(http.StatusUnauthorized, er(requestID, "unauthorized"))
	}

	// handle cancel orders
	_, from, status, err := r.dbCli.GetOrder(req.Payload.ID)
	if err != nil {
		log.Errorf("error getting order: %v, [incident: %s]", err, requestID)
		return c.JSON(http.StatusBadRequest, er(requestID, "invalid order"))
	}
	if from != sender {
		log.Errorf("error order owner and request sender mismatch, [incident: %s]", err, requestID)
		return c.JSON(http.StatusUnauthorized, er(requestID, "unauthorized"))
	}
	if status != model.StatusOpen {
		log.Errorf("error order is filled, [incident: %s]", err, requestID)
		return c.JSON(http.StatusUnauthorized, er(requestID, "processed"))
	}
	// queue the order for processing
	r.clobCli.Inbound <- req
	return c.JSON(http.StatusOK, ok(requestID, withData(keyOrderID, req.Payload.ID), withMsg("scheduled")))
}

func (r AuthexServer) getOrder(c echo.Context) error {
	requestID := c.Get(keyRequestID).(string)
	orderID := c.Param("id")
	order, _, status, err := r.dbCli.GetOrder(orderID)
	if err != nil {
		log.Errorf("error getting order: %v, [incident: %s]", err, requestID)
		return c.JSON(http.StatusNotFound, er(requestID, "order not found"))
	}
	return c.JSON(http.StatusOK, ok(requestID,
		withData("order", order),
		withData("status", status),
	))
}

func (r AuthexServer) withdraw(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, er(c.Get(keyRequestID).(string), "not implemented"))
}

// getMarketQuote returns the current quote for a given market
func (r AuthexServer) getMarketQuote(c echo.Context) error {
	requestID := c.Get(keyRequestID).(string)
	market := c.Param("address")
	side := c.Param("side")

	size, err := h.ParseAmount(c.Param("size"))
	if err != nil {
		log.Errorf("error parsing size: %v, [incident: %s]", err, requestID)
		return c.JSON(http.StatusBadRequest, er(requestID, "invalid size"))
	}
	quote, err := r.clobCli.GetQuote(market, side, size)
	if err != nil {
		if errors.Is(err, model.ErrMarketNotFound) {
			log.Errorf("error getting quote: %v, [incident: %s]", err, requestID)
			return c.JSON(http.StatusNotFound, er(requestID, "market not found"))
		}
		log.Errorf("error getting quote: %v, [incident: %s]", err, requestID)
		return c.JSON(http.StatusNotFound, er(requestID, "quote not available"))
	}
	return c.JSON(http.StatusOK, ok(requestID,
		withData("quote", quote),
		withData("market", market),
		withData("side", side),
		withData("size", size)),
	)
}

// getMarketPrice returns the current price for a given market
func (r AuthexServer) getMarketPrice(c echo.Context) error {
	requestID := c.Get(keyRequestID).(string)
	market := c.Param("address")
	price, err := r.dbCli.GetMarketPrice(market)
	if err != nil {
		if errors.Is(err, model.ErrMarketNotFound) {
			log.Errorf("error getting price: %v, [incident: %s]", err, requestID)
			return c.JSON(http.StatusNotFound, er(requestID, "market not found"))
		}
		log.Errorf("error getting price: %v, [incident: %s]", err, requestID)
		return c.JSON(http.StatusNotFound, er(requestID, "price not available"))
	}
	return c.JSON(http.StatusOK, ok(requestID, withData("price", price)))
}

func index(c echo.Context, template *template.Template, runtime *Runtime, endpoints []Endpoint) error {
	var bb bytes.Buffer
	if err := template.Execute(&bb, struct {
		Runtime   *Runtime
		Endpoints []Endpoint
	}{runtime, endpoints}); err != nil {
		return c.HTML(http.StatusInternalServerError, "ERROR")
	}
	return c.HTML(http.StatusOK, bb.String())
}
