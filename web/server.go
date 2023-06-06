package web

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"authex/model"
	"authex/network"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

//go:embed index.html
var indexHTML string

type AuthexServer struct {
	opts          *model.Settings
	echo          *echo.Echo
	runtime       *Runtime
	endpoints     []RestEndpoint
	indexTemplate *template.Template
	orders        chan *model.OrderRequest
	nodeCli       *network.NodeClient
}

type RestEndpoint struct {
	Name    string
	Method  string
	Path    string
	Help    string
	Handler func(c echo.Context) error
}

const (
	GET  string = "GET"
	POST string = "POST"
)

func NewAuthexServer(opts *model.Settings, orders chan *model.OrderRequest, nodeCli *network.NodeClient) (AuthexServer, error) {
	var err error

	r := AuthexServer{
		opts:    opts,
		orders:  orders,
		nodeCli: nodeCli,
	}
	r.echo = echo.New()
	r.echo.HideBanner = true
	r.echo.Use(middleware.Logger())
	r.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	r.echo.Logger.SetOutput(os.Stdout)
	r.echo.Logger.SetLevel(log.INFO)
	r.echo.Logger.Infof("starting CLOB server on chain %s", opts.Network.ChainID)

	// start listening for events

	r.runtime = NewRuntime(opts.Version)
	r.indexTemplate, err = template.New("index").Parse(indexHTML)
	if err != nil {
		return r, err
	}

	r.endpoints = []RestEndpoint{
		{
			Name:    "pause",
			Path:    "/admin/pause",
			Method:  POST,
			Handler: r.pause,
			Help:    "Pause the CLOB, no new orders will be accepted",
		},
		{
			// register a new market
			Name:    "register",
			Path:    "/markets",
			Method:  POST,
			Handler: r.registerMarket,
			Help:    "Register a new market",
		},
		{

			Name:    "Post order",
			Path:    "/orders",
			Method:  POST,
			Handler: r.postOrder,
			Help:    "Post a new buy or sell order",
		},
		{
			Name:    "Cancel order",
			Path:    "/orders/cancel",
			Method:  POST,
			Handler: r.cancelOrder,
			Help:    "Cancel an order",
		},
		{
			Name:    "Withdraw",
			Path:    "/withdraw",
			Method:  POST,
			Handler: r.withdraw,
			Help:    "Withdraw funds from the CLOB",
		},
		{
			Name:    "Get order",
			Path:    "/orders/:id",
			Method:  GET,
			Handler: r.getOrder,
			Help:    "Get an order by id",
		},
		{
			Name:    "Get orders",
			Path:    "/account/:address/orders",
			Method:  GET,
			Handler: r.getOrder,
			Help:    "Get all orders for an account",
		},
		{
			Name:    "Get balance",
			Path:    "/account/:address/balance/:symbol",
			Method:  GET,
			Handler: r.getOrder,
			Help:    "Get the balance of an account for a symbol",
		},
	}

	r.echo.GET("/", r.index)

	for _, endpoint := range r.endpoints {
		r.echo.Match([]string{endpoint.Method}, endpoint.Path, endpoint.Handler)
	}

	return r, nil
}

func (r AuthexServer) Start() error {
	r.echo.Logger.Infof("listening on %s", r.opts.Web.ListenAddr)
	return r.echo.Start(r.opts.Web.ListenAddr)
}

// extractAddress verifies the signature of an order
// https://goethereumbook.org/signature-verify/
// // TODO: the signature verification may be weak since tbe messages can be replayed
// probably would be good to require a time stamp in the message and verify that is not older than a few seconds
func extractAddress(msg model.SignedMessage) (err error) {
	signature, err := msg.GetSignature()
	if err != nil {
		return
	}
	// serialize the msg
	msgData, err := msg.GetData()
	if err != nil {
		return
	}
	// verify the signature
	hash := crypto.Keccak256Hash(msgData)
	sigPublicKeyECDSA, err := crypto.SigToPub(hash.Bytes(), signature)
	if err != nil {
		return
	}
	msg.SetFrom(crypto.PubkeyToAddress(*sigPublicKeyECDSA).Hex())
	return
}

func authorize(msg model.SignedMessage) (err error) {
	// verify the signature
	if err = extractAddress(msg); err != nil {
		return
	}
	if active, found := participants[msg.GetFrom()]; !found || !active {
		err = fmt.Errorf("account %s is not authorized", msg.GetFrom())
		return
	}
	return
}

func (r AuthexServer) pause(c echo.Context) error {
	// Only the admin can pause the CLOB
	return c.JSON(http.StatusNotImplemented, map[string]string{"status": "not implemented"})
}

func (r AuthexServer) registerMarket(c echo.Context) error {
	cmr := &model.CreateMarketRequest{}
	if c.Bind(cmr) != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid market request"})
	}
	if err := extractAddress(cmr); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	isAdmin, err := r.nodeCli.IsAdmin(cmr.GetFrom())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if !isAdmin {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "only the admin can register a new market"})
	}
	// TODO: insert into the DB
	// TODO: start listening for events
	// Only the admin can register a new market
	return c.JSON(http.StatusNotImplemented, map[string]string{"status": "not implemented"})
}

// postOrder submits a new order to the CLOB
// it is required that the order is signed by the account
// the fields ID and RecordedAt are overwritten by the server
func (r AuthexServer) postOrder(c echo.Context) error {
	or := &model.OrderRequest{}
	if c.Bind(or) != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid order request"})
	}
	if err := authorize(or); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	// verify that is not older than a few seconds
	or.Order.RecordedAt = time.Now().UTC()
	if or.Order.RecordedAt.Sub(or.Order.SubmittedAt) > 2*time.Second {
		err := fmt.Errorf("order is older than 2 seconds")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	// populate the order ID and RecordedAt
	or.Order.ID = uuid.New().String()
	// validate the order
	if err := or.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	// TODO: shortcut hardcode the market
	// queue the order for processing
	r.orders <- or
	// reply with the order
	return c.JSON(http.StatusOK, map[string]any{"status": "ok", "order": or.Order})
}

func (r AuthexServer) cancelOrder(c echo.Context) error {
	or := &model.OrderRequest{}
	if c.Bind(or) != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid order request"})
	}
	if err := authorize(or); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}
	or.Order.Side = model.CancelOrder
	// queue the order for processing
	r.orders <- or
	return c.JSON(http.StatusNotImplemented, map[string]string{"status": "not implemented"})
}

func (r AuthexServer) getOrder(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"status": "not implemented"})
}

func (r AuthexServer) withdraw(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"status": "not implemented"})
}

func (r AuthexServer) index(c echo.Context) error {
	var bb bytes.Buffer
	if err := r.indexTemplate.Execute(&bb, struct {
		Runtime   *Runtime
		Endpoints []RestEndpoint
	}{r.runtime, r.endpoints}); err != nil {
		return c.HTML(http.StatusInternalServerError, "ERROR")
	}
	return c.HTML(http.StatusOK, bb.String())
}
