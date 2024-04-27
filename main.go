package main

import (
	"crypto_exchange/client"
	"crypto_exchange/server"
	"fmt"
	"math"
	"time"
)

const maxOrders = 3

var (
	dur = 2 * time.Second
)

func main() {
	go server.StartServer()

	time.Sleep(time.Second)

	cl := client.NewClient()

	seedMarket(cl)

	go makeMarketSimple(cl)

	time.Sleep(time.Second)

	marketOrderPlacer(cl)
}

func seedMarket(cl *client.Client) {
	bid := &client.PlaceOrderArgs{
		UserID: "1",
		IsBid:  true,
		Size:   2_000_000,
		Price:  3500,
	}
	_, err := cl.PlaceLimitOrder(bid)
	if err != nil {
		panic(err)
	}

	ask := &client.PlaceOrderArgs{
		UserID: "1",
		Size:   2_000_000,
		Price:  3600,
	}
	_, err = cl.PlaceLimitOrder(ask)
	if err != nil {
		panic(err)
	}
}

func makeMarketSimple(cl *client.Client) {
	for range time.Tick(dur) {
		orders, err := cl.GetUserOrders(server.ETH, "2")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("user orders:", orders)

		bestBid, err := cl.GetBestPrice(server.ETH, "bid")
		if err != nil {
			fmt.Println(err)
		}
		bestAsk, err := cl.GetBestPrice(server.ETH, "ask")
		if err != nil {
			fmt.Println(err)
		}

		spread := math.Abs(bestBid - bestAsk)
		fmt.Println("Exchange spread:", spread)

		if len(orders.Bids) < maxOrders {
			bidOrderArgs := &client.PlaceOrderArgs{
				UserID: "2",
				IsBid:  true,
				Size:   1000,
				Price:  bestBid + 100,
			}
			bidOrderID, err := cl.PlaceLimitOrder(bidOrderArgs)
			if err != nil {
				fmt.Println(bidOrderID)
			}
		}

		if len(orders.Asks) < maxOrders {
			askOrderArgs := &client.PlaceOrderArgs{
				UserID: "2",
				Size:   1000,
				Price:  bestAsk - 100,
			}
			askOrderID, err := cl.PlaceLimitOrder(askOrderArgs)
			if err != nil {
				fmt.Println(askOrderID)
			}
		}

		fmt.Println("Best bid price:", bestBid)
		fmt.Println("Best ask price:", bestAsk)
	}
}

func marketOrderPlacer(cl *client.Client) {
	for range time.Tick(dur * 2) {
		trades, err := cl.GetTrades(server.ETH)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("trades:", trades)

		marketAskOrderArgs := &client.PlaceOrderArgs{
			UserID: "3",
			Size:   1000,
		}
		askOrderID, err := cl.PlaceMarketOrder(marketAskOrderArgs)
		if err != nil {
			fmt.Println(askOrderID)
		}

		marketBidOrderArgs := &client.PlaceOrderArgs{
			UserID: "3",
			IsBid:  true,
			Size:   1000,
		}
		bidOrderID, err := cl.PlaceMarketOrder(marketBidOrderArgs)
		if err != nil {
			fmt.Println(bidOrderID)
		}
	}
}
