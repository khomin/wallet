package ethereum

import (
	"context"
	"errors"
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
	userAddressHex := common.HexToAddress(address)
	tokenAddressHex := common.HexToAddress(tokenAddress)

	// --- STEP 1: AUTO-DETECT DECIMALS ---
	// Keccak256 signature hash for "decimals()" is 0x313ce567
	decimalsSignature := []byte("decimals()")
	decimalsMethodID := crypto.Keccak256(decimalsSignature)[:4]

	decimalsMsg := ethereum.CallMsg{
		To:   &tokenAddressHex,
		Data: decimalsMethodID,
	}

	var decimals int64 = 18 // Default fallback to 18
	decimalsOutput, err := c.client.CallContract(ctx, decimalsMsg, nil)
	if err == nil && len(decimalsOutput) > 0 {
		// Unpack the 32-byte hex return into a standard integer
		decimals = new(big.Int).SetBytes(decimalsOutput).Int64()
	}

	// --- STEP 2: GET RAW BALANCE ---
	transferFnSignature := []byte("balanceOf(address)")
	balanceMethodID := crypto.Keccak256(transferFnSignature)[:4]
	paddedUserAddress := common.LeftPadBytes(userAddressHex.Bytes(), 32)

	var balanceData []byte
	balanceData = append(balanceData, balanceMethodID...)
	balanceData = append(balanceData, paddedUserAddress...)

	balanceMsg := ethereum.CallMsg{
		To:   &tokenAddressHex,
		Data: balanceData,
	}

	balanceOutput, err := c.client.CallContract(ctx, balanceMsg, nil)
	if err != nil {
		return 0, err
	}

	// --- STEP 3: CONVERT DYNAMICALLY ---
	balanceInt := new(big.Int).SetBytes(balanceOutput)
	balanceFloat := new(big.Float).SetInt(balanceInt)

	// Create the divisor directly by passing the integer directly to Pow10
	// We cast the int64 'decimals' to an int
	divisorFloat := math.Pow10(int(decimals))
	divisor := new(big.Float).SetFloat64(divisorFloat)

	etherValue := new(big.Float).Quo(balanceFloat, divisor)

	f, _ := etherValue.Float64()
	return f, nil
}
