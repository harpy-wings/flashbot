package flashbot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"go.opentelemetry.io/otel/codes"
)

// Simulate runs the bundle against the Flashbots Relay to check for reverts.
// This should ALWAYS be called before Broadcast.
// blockNumber: The target block you want to land in.
// stateBlock: The block state to simulate on (usually target - 1).
func (f *flashbot) Simulate(ctx context.Context, bundle *Bundle, targetBlock uint64, opts ...BundleOption) (*SimulateResponse, error) {
	ctx, span := f.tracer.Start(ctx, "flashbot.Simulate")
	defer span.End()

	if len(bundle.Transactions) == 0 {
		span.SetStatus(codes.Error, "bundle is empty")
		return nil, fmt.Errorf("bundle cannot be empty")
	}

	// Convert transactions to hex strings
	body := make([]mevSendBundleBodyItem, 0, len(bundle.Transactions))
	for i, tx := range bundle.Transactions {
		bs, err := tx.MarshalBinary()
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			return nil, fmt.Errorf("failed to encode transaction: %w", err)
		}
		txHex := hexutil.Encode(bs)
		canRevert := false
		if i < len(bundle.CanRevert) {
			canRevert = bundle.CanRevert[i]
		}
		body = append(body, mevSendBundleBodyItem{
			Tx:        &txHex,
			CanRevert: &canRevert,
		})
	}

	// Build the mev_simBundle params
	var blockHex string
	if targetBlock == 0 {
		blockHex = "latest"
	} else {
		blockHex = "0x" + strconv.FormatUint(targetBlock, 16)
	}
	params := mevSimBundleParams{
		Version: "v0.1",
		Inclusion: mevSendBundleInclusion{
			Block: blockHex,
		},
		Body: body,
	}
	// Apply options
	for _, opt := range opts {
		err := opt(&params)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Create JSON-RPC request
	reqID := rand.Intn(1000000)
	reqBody := rpcReq{
		JsonRpc: jsonRPCVersion,
		Id:      reqID,
		Method:  methodMevSimBundle,
		Params:  []interface{}{params},
	}

	httpReq, err := f.newRequest(ctx, &reqBody)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Execute request
	resp, err := f.client.Do(httpReq)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Parse response
	var rpcResp struct {
		Id     int             `json:"id"`
		Result *MevSimResponse `json:"result,omitempty"`
		Error  *rpcError       `json:"error,omitempty"`
	}

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	err = resp.Body.Close()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, fmt.Errorf("failed to close response body: %w", err)
	}

	err = json.Unmarshal(bs, &rpcResp)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if rpcResp.Error != nil {
		span.SetStatus(codes.Error, rpcResp.Error.Message)
		return rpcResp.Result, fmt.Errorf("RPC error: %s (code: %d)", rpcResp.Error.Message, rpcResp.Error.Code)
	}

	if rpcResp.Result == nil {
		span.SetStatus(codes.Error, "empty result")
		return nil, fmt.Errorf("empty result from relay")
	}
	span.SetStatus(codes.Ok, "simulation completed successfully")
	return rpcResp.Result, nil
}

// Broadcast sends the bundle to the configured list of builders (Titan, Beaver, Flashbots, etc.).
// It returns the list of builders that accepted the request.
func (f *flashbot) Broadcast(ctx context.Context, bundle *Bundle, targetBlock uint64, opts ...BundleOption) (*BroadcastResponse, error) {
	ctx, span := f.tracer.Start(ctx, "flashbot.Broadcast")
	defer span.End()
	_, err := f.Simulate(ctx, bundle, targetBlock)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, fmt.Errorf("failed to simulate bundle: %w", err)
	}
	// Convert transactions to hex strings
	body := make([]mevSendBundleBodyItem, 0, len(bundle.Transactions))
	for i, tx := range bundle.Transactions {
		bs, err := tx.MarshalBinary()
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			return nil, fmt.Errorf("failed to encode transaction: %w", err)
		}
		txHex := hexutil.Encode(bs)
		canRevert := false
		if i < len(bundle.CanRevert) {
			canRevert = bundle.CanRevert[i]
		}
		body = append(body, mevSendBundleBodyItem{
			Tx:        &txHex,
			CanRevert: &canRevert,
		})
	}

	// Build the mev_simBundle params
	var blockHex string
	if targetBlock == 0 {
		blockHex = "latest"
	} else {
		blockHex = "0x" + strconv.FormatUint(targetBlock, 16)
	}
	maxBlock := "0x" + strconv.FormatUint(targetBlock+30, 16)
	params := mevSimBundleParams{
		Version: "v0.1",
		Inclusion: mevSendBundleInclusion{
			Block:    blockHex,
			MaxBlock: &maxBlock,
		},
		Body: body,
	}

	// Apply options
	for _, opt := range opts {
		err := opt(&params)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Create JSON-RPC request
	reqID := rand.Intn(1000000)
	reqBody := rpcReq{
		JsonRpc: jsonRPCVersion,
		Id:      reqID,
		Method:  methodMevSendBundle,
		Params:  []interface{}{params},
	}

	httpReq, err := f.newRequest(ctx, &reqBody)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Execute request
	resp, err := f.client.Do(httpReq)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Parse response
	var rpcResp struct {
		Jsonrpc string             `json:"jsonrpc"`
		ID      int                `json:"id"`
		Result  *BroadcastResponse `json:"result,omitempty"`
		Error   *rpcError          `json:"error,omitempty"`
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	fmt.Println(string(bs))
	err = resp.Body.Close()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, fmt.Errorf("failed to close response body: %w", err)
	}

	err = json.Unmarshal(bs, &rpcResp)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if rpcResp.Error != nil {
		span.SetStatus(codes.Error, rpcResp.Error.Message)
		return rpcResp.Result, fmt.Errorf("RPC error: %s (code: %d)", rpcResp.Error.Message, rpcResp.Error.Code)
	}

	if rpcResp.Result == nil {
		span.SetStatus(codes.Error, "empty result")
		return nil, fmt.Errorf("empty result from relay")
	}
	span.SetStatus(codes.Ok, "simulation completed successfully")
	return rpcResp.Result, nil

}

// SendPrivateTransaction sends a single transaction directly to builders (eth_sendPrivateTransaction).
// Useful for simple transfers where you don't need a full bundle but want frontrunning protection.
func (f *flashbot) SendPrivateTransaction(ctx context.Context, signedTxHex string, maxBlockHeight uint64) error {
	return fmt.Errorf("not implemented")
}

// --- Gas & Network Intelligence ---

// GetGasPrice returns the suggested gas price.
// For EIP-1559 chains, this should return the BaseFee + PriorityFee.
// You can implement this to return a "fast" price for aggressive inclusion.
// tip is the current optimal priority fee (bribe) based on network congestion.
// This helps you calculate the `minerBribe` dynamically rather than hardcoding 2 Gwei.
func (f *flashbot) GetGasPrice(ctx context.Context) (gasPrice *big.Int, tip *big.Int, err error) {
	ctx, span := f.tracer.Start(ctx, "flashbot.GetGasPrice")
	defer span.End()
	gasPrice, err = f.ethC.SuggestGasPrice(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, nil, fmt.Errorf("failed to get gas price: %w", err)
	}
	tip, err = f.ethC.SuggestGasTipCap(ctx)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, nil, fmt.Errorf("failed to get gas tip: %w", err)
	}
	span.SetStatus(codes.Ok, "gas tip retrieved successfully")
	return gasPrice, tip, nil
}

// EstimateGasBundle calculates the gas units required for the entire bundle.
// This is useful for calculating exactly how much "sponsorship" ETH to send the user.
func (f *flashbot) EstimateGasBundle(ctx context.Context, bundle *Bundle) (uint64, error) {
	ctx, span := f.tracer.Start(ctx, "flashbot.EstimateGasBundle")
	defer span.End()
	totalGas := uint64(0)
	for _, tx := range bundle.Transactions {
		totalGas += tx.Gas()
	}
	span.SetStatus(codes.Ok, "gas units calculated successfully")
	return totalGas, nil
}

// --- Utilities ---

// GetUserStats checks your signing key's reputation on the relay.
func (f *flashbot) GetUserStats(ctx context.Context, blockNumber *big.Int) (*UserStats, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetBundleStats checks the status of a specific bundle on the relay
// EXPERIMENTAL: This is not yet implemented.
func (f *flashbot) GetBundleStats(ctx context.Context, bundleHash string, blockNumber uint64) (*BundleStats, error) {
	return nil, fmt.Errorf("not implemented")
}

// INTERNAL METHODS
func (f *flashbot) newRequest(ctx context.Context, req *rpcReq) (*http.Request, error) {
	ctx, span := f.tracer.Start(ctx, "flashbot.newRequest")
	defer span.End()
	// Marshal request body
	jsonBody, err := req.ToJson()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Sign the request
	signature, err := f.signRequest(jsonBody)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, f.relayURL, bytes.NewReader(jsonBody))
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set(headerFlashbotSignature, signature)

	return httpReq, nil
}

// signRequest signs the request body using EIP-191 and returns the signature in the format "address:signature"
// The signature is calculated by taking the EIP-191 hash of the json body encoded as UTF-8 bytes.
func (f *flashbot) signRequest(body []byte) (string, error) {
	// Create EIP-191 message hash (accounts.TextHash prepends the Ethereum message prefix)
	// Apply EIP-191 directly on the JSON body bytes, not on a hash of the body
	hashedBody := crypto.Keccak256Hash(body).Hex()
	sig, err := crypto.Sign(accounts.TextHash([]byte(hashedBody)), f.pk)
	if err != nil {
		return "", fmt.Errorf("failed to sign request: %w", err)
	}
	// Get the address from the private key
	address := crypto.PubkeyToAddress(f.pk.PublicKey)

	// Return in format "address:signature"
	return fmt.Sprintf("%s:%s", address.Hex(), hexutil.Encode(sig)), nil
}
