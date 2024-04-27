package order_book

import (
	"fmt"
	"reflect"
	"testing"
)

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%+v != %+v", a, b)
	}
}

func TestLimit(t *testing.T) {
	orderBook := NewOrderBook()
	limit := NewLimit(10_000, orderBook)
	buyOrderA := NewOrder("1", 5, true)
	buyOrderB := NewOrder("2", 8, true)
	buyOrderC := NewOrder("3", 10, true)

	limit.AddOrder(buyOrderA)
	limit.AddOrder(buyOrderB)
	limit.AddOrder(buyOrderC)

	limit.DeleteOrder(buyOrderB)
	limit.DeleteOrder(buyOrderC)

	fmt.Println(limit)
}

func TestPlaceLimitOrder(t *testing.T) {
	orderBook := NewOrderBook()

	sellOrderA := NewOrder("1", 10, false)
	sellOrderB := NewOrder("2", 5, false)

	orderBook.PlaceLimitOrder(sellOrderA, 10_000)
	orderBook.PlaceLimitOrder(sellOrderB, 9_000)

	assert(t, len(orderBook.AskLimitsList()), 2)
	assert(t, sellOrderA, orderBook.Orders[sellOrderA.ID])
	assert(t, sellOrderB, orderBook.Orders[sellOrderB.ID])
	fmt.Println(orderBook)
}

func TestPlaceMarketOrder(t *testing.T) {
	orderBook := NewOrderBook()

	sellOrderA := NewOrder("1", 15, false)
	orderBook.PlaceLimitOrder(sellOrderA, 10_000)

	sellOrderB := NewOrder("2", 10, false)
	orderBook.PlaceLimitOrder(sellOrderB, 11_000)

	sellOrderC := NewOrder("3", 8, false)
	orderBook.PlaceLimitOrder(sellOrderC, 12_000)

	buyOrder := NewOrder("4", 30, true)
	matches := orderBook.PlaceMarketOrder(buyOrder)

	assert(t, len(matches), 3)
	assert(t, len(orderBook.AskLimitsList()), 1)
	assert(t, matches[0].Ask, sellOrderA)
	assert(t, matches[0].Bid, buyOrder)
	assert(t, matches[0].SizeFilled, 15.0)
	assert(t, matches[0].Price, 10_000.0)
	assert(t, buyOrder.IsFilled(), true)

	fmt.Println(matches)
}

func TestPlaceMarketOrderMultiFill(t *testing.T) {
	orderBook := NewOrderBook()

	buyOrderA := NewOrder("1", 5, true)
	buyOrderB := NewOrder("2", 8, true)
	buyOrderC := NewOrder("3", 10, true)
	buyOrderD := NewOrder("4", 1, true)

	orderBook.PlaceLimitOrder(buyOrderA, 10_000)
	orderBook.PlaceLimitOrder(buyOrderB, 9_000)
	orderBook.PlaceLimitOrder(buyOrderC, 11_000)
	orderBook.PlaceLimitOrder(buyOrderD, 9_000)

	assert(t, orderBook.BidsTotalVolume(), 24.0)

	sellOrder := NewOrder("5", 20, false)

	matches := orderBook.PlaceMarketOrder(sellOrder)

	assert(t, len(matches), 3)
	assert(t, len(orderBook.BidLimitsList()), 1)
	assert(t, orderBook.BidsTotalVolume(), 4.0)

	fmt.Println(matches)
}

func TestCancelOrder(t *testing.T) {
	orderBook := NewOrderBook()

	buyOrder := NewOrder("1", 5, true)
	orderBook.PlaceLimitOrder(buyOrder, 10_000)
	assert(t, orderBook.BidsTotalVolume(), 5.0)

	orderBook.CancelOrder(buyOrder)
	assert(t, orderBook.BidsTotalVolume(), 0.0)

	_, ok := orderBook.Orders[buyOrder.ID]
	assert(t, ok, false)
}
