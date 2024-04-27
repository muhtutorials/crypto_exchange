package server

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/echo/v4"
	"log"
	"time"
)

func StartServer() {
	e := echo.New()

	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatal(err)
	}

	ex, err := NewExchange(client, exchangePrivateKey)
	if err != nil {
		log.Fatal(err)
	}

	user1 := NewUser("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	user1.ID = "1"
	ex.Users[user1.ID] = user1

	user2 := NewUser("59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d")
	user2.ID = "2"
	ex.Users[user2.ID] = user2

	user3 := NewUser("5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a")
	user3.ID = "3"
	ex.Users[user3.ID] = user3

	go checkBalances(ex.Client)

	e.GET("/book/:market", ex.handleGetOrderBook)
	e.GET("/book/:market/best-price", ex.handleGetBestPrice)
	e.POST("/order", ex.handlePlaceOrder)
	e.DELETE("/order/:id", ex.handleCancelOrder)
	e.GET("/users/:market/:userID/orders", ex.handleGetUserOrders)
	e.GET("/trades/:market", ex.handleGetTrades)

	e.Logger.Fatal(e.Start(":3000"))
}

func checkBalances(client *ethclient.Client) {
	time.Sleep(10 * time.Second)

	address1 := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	balance1, _ := client.BalanceAt(context.Background(), common.HexToAddress(address1), nil)
	fmt.Printf("user1 - [address: %s], [balance: %d]\n", address1, balance1)

	address2 := "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
	balance2, _ := client.BalanceAt(context.Background(), common.HexToAddress(address2), nil)
	fmt.Printf("user2 - [address: %s], [balance: %d]\n", address2, balance2)
}
