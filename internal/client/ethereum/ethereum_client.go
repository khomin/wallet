package ethereum

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EthereumClient struct {
	rpcURL  string
	client  *ethclient.Client
	timeout time.Duration
}

func NewEthereumClient(rpcURL string) *EthereumClient {
	return &EthereumClient{
		rpcURL:  rpcURL,
		timeout: 10 * time.Second,
	}
}

func (c *EthereumClient) Connect(ctx context.Context) error {
	if c.client != nil {
		return nil
	}
	if c.rpcURL == "" {
		return errors.New("ethereum rpc url is not configured")
	}
	client, err := ethclient.DialContext(ctx, c.rpcURL)
	if err != nil {
		return err
	}
	c.client = client
	return nil
}

func (c *EthereumClient) Close() {
	if c.client != nil {
		c.client.Close()
	}
}

func (c *EthereumClient) GetBalance(ctx context.Context, address string) (float64, error) {
	if err := c.Connect(ctx); err != nil {
		return 0, err
	}
	if c.client == nil {
		return 0, errors.New("ethereum client is not initialized")
	}
	account := common.HexToAddress(address)
	balance, err := c.client.BalanceAt(ctx, account, nil)
	if err != nil {
		return 0, err
	}
	balanceFloat := new(big.Float).SetInt(balance)
	etherValue := new(big.Float).Quo(balanceFloat, big.NewFloat(math.Pow10(18)))
	f, _ := etherValue.Float64()
	return f, nil
}

func (c *EthereumClient) GetTokenBalance(ctx context.Context, address, tokenAddress string) (float64, error) {
	if address == "0x51C72848c68a965f66FA7a88855F9f7784502a7F" {
		slog.Debug("ddd")
	}
	userAddressHex := common.HexToAddress(address)
	tokenAddressHex := common.HexToAddress(tokenAddress)

	// 1. Generate the function selector for standard ERC-20 "balanceOf(address)"
	// transfer/balanceOf function signature hash is: 0x70a08231
	transferFnSignature := []byte("balanceOf(address)")
	methodID := crypto.Keccak256(transferFnSignature)[:4]

	// 2. Pad the user's address to 32 bytes to pack it correctly into the data payload
	paddedUserAddress := common.LeftPadBytes(userAddressHex.Bytes(), 32)

	// 3. Combine the method ID and the padded address parameters
	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedUserAddress...)

	// 4. Construct the standard CallMsg execution request payload
	msg := ethereum.CallMsg{
		To:   &tokenAddressHex,
		Data: data,
	}

	// 5. Execute the read call to the node
	output, err := c.client.CallContract(ctx, msg, nil)
	if err != nil {
		return 0, err
	}

	// 6. Unpack the raw 32-byte hex return data into a standard Go big.Int
	balance := new(big.Int).SetBytes(output)
	// return balance, nil
	if err != nil {
		return 0, err
	}
	balanceFloat := new(big.Float).SetInt(balance)
	etherValue := new(big.Float).Quo(balanceFloat, big.NewFloat(math.Pow10(18)))
	f, _ := etherValue.Float64()
	return f, nil
}
