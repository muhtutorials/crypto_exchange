package order_book

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type Trade struct {
	IsBid     bool
	Price     float64
	Size      float64
	Timestamp int64
}

type OrderBook struct {
	bidLimitsList Limits
	askLimitsList Limits
	BidLimits     map[float64]*Limit
	AskLimits     map[float64]*Limit
	OrdersMu      sync.RWMutex
	Orders        map[string]*Order
	Trades        []*Trade
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		BidLimits: make(map[float64]*Limit),
		AskLimits: make(map[float64]*Limit),
		Orders:    make(map[string]*Order),
	}
}

func (ob *OrderBook) PlaceLimitOrder(order *Order, price float64) {
	var limit *Limit

	if order.IsBid {
		limit = ob.BidLimits[price]
	} else {
		limit = ob.AskLimits[price]
	}

	if limit == nil {
		limit = NewLimit(price, ob)
		if order.IsBid {
			ob.bidLimitsList = append(ob.bidLimitsList, limit)
			ob.BidLimits[price] = limit
		} else {
			ob.askLimitsList = append(ob.askLimitsList, limit)
			ob.AskLimits[price] = limit
		}
	}

	limit.AddOrder(order)

	ob.OrdersMu.Lock()
	ob.Orders[order.ID] = order
	ob.OrdersMu.Unlock()
}

func (ob *OrderBook) PlaceMarketOrder(order *Order) []Match {
	var (
		matches        []Match
		limitsToDelete []*Limit
	)

	if order.IsBid {
		if order.Size > ob.AsksTotalVolume() {
			panic(fmt.Sprintf("not enough volume [size: %.2f] for market order [size: %.2f]", ob.AsksTotalVolume(), order.Size))
		}
		for _, limit := range ob.AskLimitsList() {
			limitMatches := limit.Fill(order)
			matches = append(matches, limitMatches...)

			if len(limit.Orders) == 0 {
				limitsToDelete = append(limitsToDelete, limit)
			}

			if order.IsFilled() {
				break
			}
		}
	} else {
		if order.Size > ob.BidsTotalVolume() {
			panic(fmt.Sprintf("not enough volume [size: %.2f] for market order [size: %.2f]", ob.BidsTotalVolume(), order.Size))
		}
		for _, limit := range ob.BidLimitsList() {
			limitMatches := limit.Fill(order)
			matches = append(matches, limitMatches...)

			if len(limit.Orders) == 0 {
				limitsToDelete = append(limitsToDelete, limit)
			}

			if order.IsFilled() {
				break
			}
		}
	}

	for _, l := range limitsToDelete {
		ob.deleteLimit(l, !order.IsBid)
	}

	for _, match := range matches {
		trade := &Trade{
			IsBid:     order.IsBid,
			Size:      match.SizeFilled,
			Price:     match.Price,
			Timestamp: time.Now().UnixNano(),
		}

		ob.Trades = append(ob.Trades, trade)
	}

	return matches
}

func (ob *OrderBook) BidsTotalVolume() float64 {
	var totalVolume float64

	for i := 0; i < len(ob.bidLimitsList); i++ {
		totalVolume += ob.bidLimitsList[i].TotalVolume
	}

	return totalVolume
}

func (ob *OrderBook) AsksTotalVolume() float64 {
	var totalVolume float64

	for i := 0; i < len(ob.askLimitsList); i++ {
		totalVolume += ob.askLimitsList[i].TotalVolume
	}

	return totalVolume
}

func (ob *OrderBook) BidLimitsList() []*Limit {
	sort.Sort(ByBestBid(ob.bidLimitsList))
	return ob.bidLimitsList
}

func (ob *OrderBook) AskLimitsList() []*Limit {
	sort.Sort(ByBestAsk(ob.askLimitsList))
	return ob.askLimitsList
}

func (ob *OrderBook) deleteLimit(limit *Limit, isBid bool) {
	if isBid {
		for i, l := range ob.bidLimitsList {
			if limit == l {
				ob.bidLimitsList = append(ob.bidLimitsList[:i], ob.bidLimitsList[i+1:]...)
			}
		}
		delete(ob.BidLimits, limit.Price)
	} else {
		for i, l := range ob.askLimitsList {
			if limit == l {
				ob.askLimitsList = append(ob.askLimitsList[:i], ob.askLimitsList[i+1:]...)
			}
		}
		delete(ob.AskLimits, limit.Price)
	}
}

func (ob *OrderBook) CancelOrder(order *Order) {
	limit := order.Limit
	limit.DeleteOrder(order)

	if len(limit.Orders) == 0 {
		ob.deleteLimit(limit, order.IsBid)
	}

	delete(ob.Orders, order.ID)
}

type UserOrders struct {
	Bids []*Order
	Asks []*Order
}

func (ob *OrderBook) GetUserOrders(userID string) *UserOrders {
	ob.OrdersMu.RLock()
	defer ob.OrdersMu.RUnlock()

	userOrders := &UserOrders{}

	for _, order := range ob.Orders {
		if userID == order.UserID {
			if order.IsBid {
				userOrders.Bids = append(userOrders.Bids, order)
			} else {
				userOrders.Asks = append(userOrders.Asks, order)
			}
		}
	}

	return userOrders
}
