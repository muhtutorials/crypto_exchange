package server

import (
	"crypto/ecdsa"
	"crypto_exchange/order_book"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"math/big"
	"net/http"
)

const (
	ETH Market = "ETH"

	MarketOrder OrderType = "market"
	LimitOrder  OrderType = "limit"

	exchangePrivateKey = "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d"
)

type (
	Market    string
	OrderType string

	Exchange struct {
		orderBooks map[Market]*order_book.OrderBook
		Client     *ethclient.Client
		PrivateKey *ecdsa.PrivateKey
		Users      map[string]*User
	}

	User struct {
		ID         string
		PrivateKey *ecdsa.PrivateKey
	}

	Order struct {
		ID        string  `json:"id"`
		UserID    string  `json:"user_id"`
		IsBid     bool    `json:"is_bid"`
		Size      float64 `json:"size"`
		Price     float64 `json:"price"`
		Timestamp int64   `json:"timestamp"`
	}

	OrderBookRes struct {
		Bids            []*Order          `json:"bids"`
		Asks            []*Order          `json:"asks"`
		BidsTotalVolume float64           `json:"bids_total_volume"`
		AsksTotalVolume float64           `json:"asks_total_volume"`
		Orders          map[string]*Order `json:"orders"`
	}

	PlaceOrderReq struct {
		UserID    string `json:"user_id"`
		Market    `json:"market"`
		OrderType `json:"type"`
		IsBid     bool    `json:"is_bid"`
		Size      float64 `json:"size"`
		Price     float64 `json:"price"`
	}

	PlaceOrderRes struct {
		Message string `json:"message"`
		OrderID string `json:"order_id"`
	}

	Match struct {
		ID    string  `json:"id"`
		Size  float64 `json:"size"`
		Price float64 `json:"price"`
	}

	BestPrice struct {
		Price float64 `json:"price"`
	}

	UserOrders struct {
		Bids []*Order `json:"bids"`
		Asks []*Order `json:"asks"`
	}
)

func NewExchange(client *ethclient.Client, privateKey string) (*Exchange, error) {
	orderBooks := make(map[Market]*order_book.OrderBook)
	orderBooks[ETH] = order_book.NewOrderBook()

	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, err
	}

	return &Exchange{
		orderBooks: orderBooks,
		Client:     client,
		PrivateKey: pk,
		Users:      make(map[string]*User),
	}, nil
}

func NewUser(privateKey string) *User {
	id := uuid.NewString()

	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		panic(err)
	}

	return &User{
		ID:         id,
		PrivateKey: pk,
	}
}

func (ex *Exchange) handleGetOrderBook(c echo.Context) error {
	market := Market(c.Param("market"))

	orderBook, ok := ex.orderBooks[market]
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "market not found",
		})
	}

	var orderBookRes OrderBookRes
	for _, limit := range orderBook.BidLimitsList() {
		for _, order := range limit.Orders {
			bid := toOrder(order)
			orderBookRes.Bids = append(orderBookRes.Bids, bid)
		}
	}

	for _, limit := range orderBook.AskLimitsList() {
		for _, order := range limit.Orders {
			ask := toOrder(order)
			orderBookRes.Asks = append(orderBookRes.Asks, ask)
		}
	}

	orderBookRes.BidsTotalVolume = orderBook.BidsTotalVolume()
	orderBookRes.AsksTotalVolume = orderBook.AsksTotalVolume()

	orderBookRes.Orders = make(map[string]*Order)
	for id, order := range orderBook.Orders {
		orderBookRes.Orders[id] = toOrder(order)
	}

	return c.JSON(http.StatusOK, orderBookRes)
}

func (ex *Exchange) handlePlaceOrder(c echo.Context) error {
	var data PlaceOrderReq
	if err := c.Bind(&data); err != nil {
		return err
	}

	orderBook := ex.orderBooks[data.Market]
	order := order_book.NewOrder(data.UserID, data.Size, data.IsBid)

	if data.OrderType == LimitOrder {
		if err := ex.handlePlaceLimitOrder(orderBook, order, data.Price); err != nil {
			return err
		}
	}

	if data.OrderType == MarketOrder {
		matches, _ := ex.handlePlaceMarketOrder(orderBook, order)

		if err := ex.handleMatches(matches); err != nil {
			return err
		}
	}

	return c.JSON(http.StatusOK, PlaceOrderRes{
		Message: "Order placed",
		OrderID: order.ID,
	})
}

func (ex *Exchange) handlePlaceLimitOrder(orderBook *order_book.OrderBook, order *order_book.Order, price float64) error {
	orderBook.PlaceLimitOrder(order, price)
	return nil
}

func (ex *Exchange) handlePlaceMarketOrder(orderBook *order_book.OrderBook, order *order_book.Order) ([]order_book.Match, []*Match) {
	var isBid bool
	if order.IsBid {
		isBid = true
	}

	matches := orderBook.PlaceMarketOrder(order)
	matchesRes := make([]*Match, len(matches))

	var totalSizeFilled float64
	var sumPrice float64
	for i := 0; i < len(matchesRes); i++ {
		id := matches[i].Bid.ID
		if isBid {
			id = matches[i].Ask.ID
		}
		matchesRes[i] = &Match{
			ID:    id,
			Size:  matches[i].SizeFilled,
			Price: matches[i].Price,
		}
		totalSizeFilled += matches[i].SizeFilled
		sumPrice += matches[i].Price
	}

	averagePrice := sumPrice / float64(len(matches))

	fmt.Printf("Filled market order => ID: %s | size: %.2f | price: %.2f \n", order.ID, totalSizeFilled, averagePrice)

	return matches, matchesRes
}

func (ex *Exchange) handleMatches(matches []order_book.Match) error {
	for _, match := range matches {
		fromUser, ok := ex.Users[match.Ask.UserID]
		if !ok {
			return fmt.Errorf("from user not found: %s", match.Ask.UserID)
		}

		toUser, ok := ex.Users[match.Bid.UserID]
		if !ok {
			return fmt.Errorf("to user not found: %s", match.Bid.UserID)
		}

		amount := big.NewInt(int64(match.SizeFilled))

		err := transferETH(ex.Client, fromUser.PrivateKey, toUser.PrivateKey, amount)
		if err != nil {
			fmt.Println("transferETH:", err)
		}
	}

	return nil
}

func (ex *Exchange) handleCancelOrder(c echo.Context) error {
	orderID := c.Param("id")

	orderBook := ex.orderBooks[ETH]

	order, ok := orderBook.Orders[orderID]
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "Order not found",
		})
	}

	orderBook.CancelOrder(order)

	return c.JSON(http.StatusOK, map[string]any{
		"message":  "Order deleted",
		"order_id": order.ID,
	})
}

func (ex *Exchange) handleGetBestPrice(c echo.Context) error {
	market := Market(c.Param("market"))
	limitType := c.QueryParam("type")

	orderBook, ok := ex.orderBooks[market]
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "market not found",
		})
	}

	var bestPrice float64

	if limitType == "bid" {
		if len(orderBook.BidLimitsList()) == 0 {
			return fmt.Errorf("no bids in the order book")
		}
		bestPrice = orderBook.BidLimitsList()[0].Price
	} else if limitType == "ask" {
		if len(orderBook.AskLimitsList()) == 0 {
			return fmt.Errorf("no asks in the order book")
		}
		bestPrice = orderBook.AskLimitsList()[0].Price
	} else {
		return fmt.Errorf("wrong type of limit")
	}

	return c.JSON(http.StatusOK, BestPrice{Price: bestPrice})
}

func (ex *Exchange) handleGetUserOrders(c echo.Context) error {
	market := Market(c.Param("market"))
	userID := c.Param("userID")

	orderBook, ok := ex.orderBooks[market]
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "market not found",
		})
	}

	userOrders := orderBook.GetUserOrders(userID)
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "userID not found",
		})
	}

	userOrdersRes := &UserOrders{
		Bids: make([]*Order, len(userOrders.Bids)),
		Asks: make([]*Order, len(userOrders.Asks)),
	}

	for i, order := range userOrders.Bids {
		userOrdersRes.Bids[i] = toOrder(order)
	}

	for i, order := range userOrders.Asks {
		userOrdersRes.Asks[i] = toOrder(order)
	}

	return c.JSON(http.StatusOK, userOrdersRes)
}

func (ex *Exchange) handleGetTrades(c echo.Context) error {
	market := Market(c.Param("market"))

	orderBook, ok := ex.orderBooks[market]
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "market not found",
		})
	}
	return c.JSON(http.StatusOK, orderBook.Trades)
}
