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
	Inbound chan *model.OrderRequest
	// order matches
	Matches chan *model.Match
}

func NewPool(matches chan *model.Match) *Pool {
	return &Pool{
		markets:  make(map[string]*ob.OrderBook),
		balances: make(map[string]map[string]decimal.Decimal),
		Inbound:  make(chan *model.OrderRequest),
		Matches:  matches,
	}
}

func (p *Pool) Run() {
	for {
		select {
		case order := <-p.Inbound:
			p.handleOrder(order)
		default:
			// channel is closed
			return
		}
	}
}

func (p *Pool) handleOrder(r *model.OrderRequest) {
	// get the order book for the symbol
	orderBook, ok := p.markets[r.Order.Market]
	if !ok {
		orderBook = ob.NewOrderBook()
		p.markets[r.Order.Market] = orderBook
	}
	// NOW UPDATE THE DATABASE

	// if it is a cancel order, cancel it
	if r.Order.Side == model.CancelOrder {
		orderBook.CancelOrder(r.Order.ID)
		return
	}
	// check the side
	side := ob.Buy
	if r.Order.Side == model.SideAsk {
		side = ob.Sell
	}
	quantity := decimal.NewFromInt(int64(r.Order.Size))
	price, err := decimal.NewFromString(r.Order.Price)
	if err != nil {
		log.Error(err)
		return
	}
	var (
		done    []*ob.Order
		partial *ob.Order
	)
	if r.Order.Price == "" {
		// limit order
		if done, partial, _, _, err = orderBook.ProcessMarketOrder(side, quantity); err != nil {
			log.Error(err)
			return
		}
	} else {
		// market order
		if done, partial, _, err = orderBook.ProcessLimitOrder(side, r.Order.ID, quantity, price); err != nil {
			log.Error(err)
			return
		}
	}
	go p.ProcessMatches(r, done, partial)
}

func (p *Pool) ProcessMatches(r *model.OrderRequest, done []*ob.Order, partial *ob.Order) {
	// for _, order := range done {
	// 	p.Matches <- model.NewMatch(r, order.ID(), order.Price())
	// }
	// if partial != nil {
	// 	p.Matches <- model.NewMatch(r, partial.ID(), partial.Price())
	// }
}
