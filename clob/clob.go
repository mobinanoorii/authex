package clob

import (
	"authex/model"

	ob "github.com/i25959341/orderbook"
	"github.com/labstack/gommon/log"
	"github.com/shopspring/decimal"
)

type Pool struct {
	// order books, indexed by symbol
	markets map[string]*ob.OrderBook
	// account balances, indexed by address and symbol
	balances map[string]map[string]decimal.Decimal
	// incoming orders
	Inbound chan *model.SignedRequest[model.Order]
	// order matches
	Matches chan *model.Match
}

func NewPool(matches chan *model.Match) *Pool {
	return &Pool{
		markets:  make(map[string]*ob.OrderBook),
		balances: make(map[string]map[string]decimal.Decimal),
		Inbound:  make(chan *model.SignedRequest[model.Order]),
		Matches:  matches,
	}
}

func (p *Pool) Close() {
	close(p.Inbound)
}

func (p *Pool) Run() {
	for {
		order, ok := <-p.Inbound
		if !ok {
			// channel is closed
			return
		}
		p.handleOrder(order)
	}
}

func (p *Pool) OpenMarket(market string) {
	if _, ok := p.markets[market]; !ok {
		p.markets[market] = ob.NewOrderBook()
	}
}

func (p *Pool) handleOrder(r *model.SignedRequest[model.Order]) {
	// get the order book for the symbol
	orderBook, ok := p.markets[r.Payload.Market]
	if !ok {
		orderBook = ob.NewOrderBook()
		p.markets[r.Payload.Market] = orderBook
	}
	// if it is a cancel order, cancel it
	if r.Payload.Side == model.CancelOrder {
		orderBook.CancelOrder(r.Payload.ID)
		return
	}
	// check the side
	side := ob.Buy
	if r.Payload.Side == model.SideAsk {
		side = ob.Sell
	}
	quantity := decimal.NewFromInt(int64(r.Payload.Size))

	var (
		done    []*ob.Order
		partial *ob.Order
		err     error
	)
	if r.Payload.Price == "" {
		// limit order
		if done, partial, _, _, err = orderBook.ProcessMarketOrder(side, quantity); err != nil {
			log.Error(err)
			return
		}
	} else {
		price, err := decimal.NewFromString(r.Payload.Price)
		if err != nil {
			log.Error(err)
			return
		}
		// market order
		if done, partial, _, err = orderBook.ProcessLimitOrder(side, r.Payload.ID, quantity, price); err != nil {
			log.Error(err)
			return
		}
	}
	for _, order := range done {
		m := orderToMatch(r.Payload.ID, order, model.StatusFilled)
		log.Debugf("Matched order DONE %s price %s", order.ID(), order.Price())
		p.Matches <- m

	}
	if partial != nil {
		m := orderToMatch(r.Payload.ID, partial, model.StatusPartial)
		log.Debugf("Matched order PARTIAL %s price %s", partial.ID(), partial.Price())
		p.Matches <- m
	}
}

func orderToMatch(topID string, order *ob.Order, status string) *model.Match {
	side := model.SideBid
	if order.Side() == ob.Sell {
		side = model.SideAsk
	}
	m := &model.Match{
		ID:      topID,
		OrderID: order.ID(),
		Price:   order.Price(),
		Size:    order.Quantity(),
		Time:    order.Time(),
		Side:    side,
		Status:  status,
	}
	return m
}

func (p *Pool) GetOrderBook(market string) string {
	x := p.markets[market]
	a, b := x.Depth()
	for _, o := range a {
		log.Infof("ASK %s %s", o.Price, o.Quantity)
	}
	for _, o := range b {
		log.Infof("BID %s %s", o.Price, o.Quantity)
	}
	return p.markets[market].String()
}

// GetQuote returns the best bid and ask prices for a given market
func (p *Pool) GetQuote(market, side string, size decimal.Decimal) (price decimal.Decimal, err error) {
	orderBook, ok := p.markets[market]
	if !ok {
		err = model.ErrMarketNotFound
		return
	}
	obSide := ob.Buy
	if side == model.SideAsk {
		obSide = ob.Sell
	}
	price, err = orderBook.CalculateMarketPrice(obSide, size)
	return
}
