package client

import (
	"bytes"
	"crypto_exchange/order_book"
	"crypto_exchange/server"
	"encoding/json"
	"fmt"
	"net/http"
)

const baseURL = "http://localhost:3000"

type Client struct {
	*http.Client
}

func NewClient() *Client {
	return &Client{
		Client: http.DefaultClient,
	}
}

type PlaceOrderArgs struct {
	UserID string
	IsBid  bool
	Size   float64
	Price  float64
}

func (c *Client) PlaceLimitOrder(args *PlaceOrderArgs) (*server.PlaceOrderRes, error) {
	url := baseURL + "/order"

	data := &server.PlaceOrderReq{
		UserID:    args.UserID,
		Market:    server.ETH,
		OrderType: server.LimitOrder,
		IsBid:     args.IsBid,
		Size:      args.Size,
		Price:     args.Price,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	placeOrderRes := &server.PlaceOrderRes{}

	err = json.NewDecoder(res.Body).Decode(placeOrderRes)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = res.Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	return placeOrderRes, nil
}

func (c *Client) PlaceMarketOrder(args *PlaceOrderArgs) (*server.PlaceOrderRes, error) {
	url := baseURL + "/order"

	data := &server.PlaceOrderReq{
		UserID:    args.UserID,
		Market:    server.ETH,
		OrderType: server.MarketOrder,
		IsBid:     args.IsBid,
		Size:      args.Size,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	placeOrderRes := &server.PlaceOrderRes{}

	err = json.NewDecoder(res.Body).Decode(placeOrderRes)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = res.Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	return placeOrderRes, nil
}

func (c *Client) CancelOrder(orderID string) error {
	url := baseURL + "/order/" + orderID

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := c.Do(req)
	if err != nil {
		return err
	}

	fmt.Println("Order cancel:", res.Status)

	return nil
}

func (c *Client) GetBestPrice(market server.Market, limitType string) (float64, error) {
	url := fmt.Sprintf("%s/book/%s/best-price?type=%s", baseURL, market, limitType)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	res, err := c.Do(req)
	if err != nil {
		return 0, err
	}

	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("GetBestPrice: %s", res.Status)
	}

	bestPrice := &server.BestPrice{}

	err = json.NewDecoder(res.Body).Decode(bestPrice)
	if err != nil {
		return 0, err
	}
	defer func() {
		err = res.Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	return bestPrice.Price, nil
}

func (c *Client) GetUserOrders(market server.Market, userID string) (*server.UserOrders, error) {
	url := fmt.Sprintf("%s/users/%s/%s/orders", baseURL, market, userID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetUserOrders: %s", res.Status)
	}

	userOrders := &server.UserOrders{}

	err = json.NewDecoder(res.Body).Decode(userOrders)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = res.Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	return userOrders, nil
}

func (c *Client) GetTrades(market server.Market) ([]*order_book.Trade, error) {
	url := fmt.Sprintf("%s/trades/%s", baseURL, market)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetUserOrders: %s", res.Status)
	}

	var trades []*order_book.Trade

	err = json.NewDecoder(res.Body).Decode(&trades)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = res.Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	return trades, nil
}
