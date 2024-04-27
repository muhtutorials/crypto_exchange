package order_book

import (
	"fmt"
	"sort"
)

type Match struct {
	Ask        *Order
	Bid        *Order
	SizeFilled float64
	Price      float64
}

type Limit struct {
	Price float64
	Orders
	TotalVolume float64
	OrderBook   *OrderBook
}

func NewLimit(price float64, orderBook *OrderBook) *Limit {
	return &Limit{
		Price:     price,
		OrderBook: orderBook,
	}
}

func (l *Limit) String() string {
	return fmt.Sprintf("[price: %.2f | total volume: %.2f]", l.Price, l.TotalVolume)
}

func (l *Limit) AddOrder(order *Order) {
	order.Limit = l
	l.Orders = append(l.Orders, order)
	l.TotalVolume += order.Size
}

func (l *Limit) DeleteOrder(order *Order) {
	for i, o := range l.Orders {
		if order == o {
			l.Orders = append(l.Orders[:i], l.Orders[i+1:]...)
		}
	}

	// removes pointer
	order.Limit = nil
	l.TotalVolume -= order.Size

	sort.Sort(l.Orders)

	delete(l.OrderBook.Orders, order.ID)
}

func (l *Limit) Fill(order *Order) []Match {
	var (
		matches        []Match
		ordersToDelete []*Order
	)

	for _, o := range l.Orders {
		match := l.FillOrder(order, o)
		matches = append(matches, match)

		l.TotalVolume -= match.SizeFilled

		if o.IsFilled() {
			ordersToDelete = append(ordersToDelete, o)
		}

		if order.IsFilled() {
			break
		}
	}

	for _, o := range ordersToDelete {
		l.DeleteOrder(o)
	}

	return matches
}

func (l *Limit) FillOrder(a, b *Order) Match {
	var (
		bid        *Order
		ask        *Order
		sizeFilled float64
	)

	if a.IsBid {
		bid = a
		ask = b
	} else {
		bid = b
		ask = a
	}

	if a.Size >= b.Size {
		a.Size -= b.Size
		sizeFilled = b.Size
		b.Size = 0
	} else {
		b.Size -= a.Size
		sizeFilled = a.Size
		a.Size = 0
	}

	return Match{
		Bid:        bid,
		Ask:        ask,
		SizeFilled: sizeFilled,
		Price:      l.Price,
	}
}

type Limits []*Limit

type ByBestBid Limits

func (l ByBestBid) Len() int {
	return len(l)
}

func (l ByBestBid) Less(i, j int) bool {
	return l[i].Price > l[j].Price
}

func (l ByBestBid) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

type ByBestAsk Limits

func (l ByBestAsk) Len() int {
	return len(l)
}

func (l ByBestAsk) Less(i, j int) bool {
	return l[i].Price < l[j].Price
}

func (l ByBestAsk) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
