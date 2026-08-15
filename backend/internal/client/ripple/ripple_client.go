package ripple

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"tracker/internal/client"
)

type jsonRPCRequest struct {
	Method string         `json:"method"`
	Params []jsonRPCParam `json:"params"`
}

type jsonRPCParam struct {
	Account     string `json:"account,omitempty"`
	Strict      bool   `json:"strict,omitempty"`
	LedgerIndex string `json:"ledger_index,omitempty"`
	Peer        string `json:"peer,omitempty"`
	Limit       int    `json:"limit,omitempty"`
}

type jsonRPCResponse struct {
	Result json.RawMessage `json:"result"`
	Error  string          `json:"error"`
	Status string          `json:"status"`
}

type accountInfoResult struct {
	AccountData struct {
		Balance string `json:"Balance"`
	} `json:"account_data"`
}

type accountLinesResult struct {
	Lines []struct {
		Account  string `json:"account"`
		Balance  string `json:"balance"`
		Currency string `json:"currency"`
	} `json:"lines"`
}

type RippleClient struct {
	rpcURL     string
	httpClient *http.Client
}

func NewRippleClient(rpcURL string) client.ChainProvider {
	return &RippleClient{
		rpcURL: rpcURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *RippleClient) Connect(ctx context.Context) error {
	// XRPL JSON-RPC is HTTP-based, no persistent connection needed
	return nil
}

func (c *RippleClient) Close() {
	// Nothing to close for HTTP client
}

// GetBalance returns the XRP balance (in XRP) for a given address.
// The XRPL API returns balance in drops (1 XRP = 1,000,000 drops).
func (c *RippleClient) GetBalance(ctx context.Context, address string) (float64, error) {
	if address == "" {
		return 0, errors.New("xrpl address is required")
	}
	var result accountInfoResult
	if err := c.call(ctx, "account_info", jsonRPCParam{
		Account:     address,
		Strict:      true,
		LedgerIndex: "validated",
	}, &result); err != nil {
		if strings.Contains(err.Error(), "actNotFound") {
			return 0, nil // unfunded / not found → 0 balance
		}
		return 0, err
	}

	// Parse the balance (drops → XRP)
	var drops float64
	if _, err := fmt.Sscanf(result.AccountData.Balance, "%f", &drops); err != nil {
		return 0, fmt.Errorf("failed to parse XRP balance: %w", err)
	}
	return drops / 1e6, nil
}

// GetTokenBalance returns the balance of an issued token (e.g. RLUSD) on the
// XRP Ledger.  tokenAddress should be the **issuer** account.
func (c *RippleClient) GetTokenBalance(ctx context.Context, address, tokenAddress string) (float64, error) {
	if address == "" || tokenAddress == "" {
		return 0, errors.New("address and tokenAddress (issuer) are required")
	}

	var result accountLinesResult
	if err := c.call(ctx, "account_lines", jsonRPCParam{
		Account: address,
		Peer:    tokenAddress,
		Limit:   400,
	}, &result); err != nil {
		return 0, err
	}

	// Sum all trust-line balances for the given issuer
	var total float64
	for _, line := range result.Lines {
		var bal float64
		if _, err := fmt.Sscanf(line.Balance, "%f", &bal); err != nil {
			continue
		}
		total += bal
	}
	return total, nil
}

// ValidateAddress performs basic structural validation of an XRPL address.
func (c *RippleClient) ValidateAddress(address, tokenAddress string) error {
	if !isValidXRPLAddress(address) {
		return fmt.Errorf("invalid xrpl address")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// call executes a JSON-RPC call against the configured XRPL node.
func (c *RippleClient) call(ctx context.Context, method string, params jsonRPCParam, dest interface{}) error {
	reqBody := jsonRPCRequest{
		Method: method,
		Params: []jsonRPCParam{params},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.rpcURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("rpc call: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	var rpcResp jsonRPCResponse
	if err := json.Unmarshal(respBytes, &rpcResp); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}

	if rpcResp.Error != "" {
		return fmt.Errorf("xrpl error: %s", rpcResp.Error)
	}
	if rpcResp.Status != "success" && rpcResp.Status != "" {
		return fmt.Errorf("xrpl status: %s", rpcResp.Status)
	}

	if dest != nil {
		if err := json.Unmarshal(rpcResp.Result, dest); err != nil {
			return fmt.Errorf("unmarshal result: %w", err)
		}
	}
	return nil
}

// XRPL classic address: starts with 'r', 25–35 base58 characters.
var xrplAddrRegex = regexp.MustCompile(`^r[1-9A-HJ-NP-Za-km-z]{24,34}$`)

func isValidXRPLAddress(addr string) bool {
	return xrplAddrRegex.MatchString(addr)
}
