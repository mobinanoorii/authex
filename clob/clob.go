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

func (p *Pool) Run() {
	for {
		select {
		case order, ok := <-p.Inbound:
			if !ok {
				// channel is closed
				return
			}
			p.handleOrder(order)
		}
	}
}

func (p *Pool) handleOrder(r *model.SignedRequest[model.Order]) {
	// get the order book for the symbol
	orderBook, ok := p.markets[r.Payload.Market]
	if !ok {
		orderBook = ob.NewOrderBook()
		p.markets[r.Payload.Market] = orderBook
	}
	// NOW UPDATE THE DATABASE

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
	price, err := decimal.NewFromString(r.Payload.Price)
	if err != nil {
		log.Error(err)
		return
	}
	var (
		done    []*ob.Order
		partial *ob.Order
	)
	if r.Payload.Price == "" {
		// limit order
		if done, partial, _, _, err = orderBook.ProcessMarketOrder(side, quantity); err != nil {
			log.Error(err)
			return
		}
	} else {
		// market order
		if done, partial, _, err = orderBook.ProcessLimitOrder(side, r.Payload.ID, quantity, price); err != nil {
			log.Error(err)
			return
		}
	}
	go p.ProcessMatches(r, done, partial)
}

func (p *Pool) ProcessMatches(r *model.SignedRequest[model.Order] , done []*ob.Order, partial *ob.Order) {
	log.Info("Processing matches")
	// for _, order := range done {
	// 	p.Matches <- model.NewMatch(r, order.ID(), order.Price())
	// }
	// if partial != nil {
	// 	p.Matches <- model.NewMatch(r, partial.ID(), partial.Price())
	// }
}
