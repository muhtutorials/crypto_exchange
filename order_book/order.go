package order_book

import (
	"fmt"
	"github.com/google/uuid"
	"time"
)

type Order struct {
	ID        string
	UserID    string
	Size      float64
	IsBid     bool
	Timestamp int64
	Limit     *Limit
}

func NewOrder(userID string, size float64, isBid bool) *Order {
	return &Order{
		ID:        uuid.NewString(),
		UserID:    userID,
		Size:      size,
		IsBid:     isBid,
		Timestamp: time.Now().UnixNano(),
	}
}

func (o *Order) String() string {
	var (
		orderType  string
		limitPrice float64
	)

	if o.IsBid {
		orderType = "bid"
	} else {
		orderType = "ask"
	}

	if o.Limit != nil {
		limitPrice = o.Limit.Price
	}

	return fmt.Sprintf("[type: %s, size: %.2f, price: %.2f]", orderType, o.Size, limitPrice)
}

func (o *Order) IsFilled() bool {
	return o.Size == 0
}

type Orders []*Order

func (o Orders) Len() int {
	return len(o)
}

func (o Orders) Less(i, j int) bool {
	return o[i].Timestamp < o[j].Timestamp
}

func (o Orders) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}
