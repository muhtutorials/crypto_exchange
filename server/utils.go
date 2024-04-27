package server

import (
	"context"
	"crypto/ecdsa"
	"crypto_exchange/order_book"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
)

func transferETH(client *ethclient.Client, from *ecdsa.PrivateKey, to *ecdsa.PrivateKey, amount *big.Int) error {
	ctx := context.Background()

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fromAddress, err := getAddress(from)
	if err != nil {
		return err
	}
	toAddress, err := getAddress(to)
	if err != nil {
		return err
	}

	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return err
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatal(err)
	}
	gasLimit := uint64(21000)

	txData := &types.AccessListTx{
		ChainID:  chainID,
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &toAddress,
		Value:    amount,
	}

	tx := types.NewTx(txData)

	signedTx, err := types.SignTx(tx, types.NewLondonSigner(chainID), from)
	if err != nil {
		return err
	}

	return client.SendTransaction(ctx, signedTx)
}

func getAddress(privateKey *ecdsa.PrivateKey) (common.Address, error) {
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Address{}, fmt.Errorf("error casting public key to ECDSA")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return address, nil
}

func toOrder(order *order_book.Order) *Order {
	return &Order{
		ID:        order.ID,
		UserID:    order.UserID,
		IsBid:     order.IsBid,
		Size:      order.Size,
		Price:     order.Limit.Price,
		Timestamp: order.Timestamp,
	}
}
