package flashbot

import (
	"encoding/json"
)

// method is a type that represents the method name of the RPC call.
// it is used to identify the method name of the RPC call.
type method string

type SimulateResponse = MevSimResponse

type BroadcastResult struct {
	Builder string
	Status  string
	Message string
}

type UserStats struct {
	Reputation float64
	Stats      map[string]interface{}
}

type BundleStats struct {
	BundleHash  string
	BlockNumber uint64
	Status      string
	Message     string
}

// JsonRpcRequest is the request body for a JSON-RPC request.

type rpcReq struct {
	JsonRpc string        `json:"jsonrpc"`
	Id      int           `json:"id"`
	Method  method        `json:"method"`
	Params  []interface{} `json:"params"` // Array of objects
}

func (r *rpcReq) ToJson() ([]byte, error) {
	return json.Marshal(r)
}

type JsonRpcResponse struct {
	Id     int             `json:"id"`
	Result *callBundleResp `json:"result,omitempty"`
	Error  *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type callBundleResp struct {
	BundleHash   string `json:"bundleHash"`
	CoinbaseDiff string `json:"coinbaseDiff"` // Miner Profit
	Results      []struct {
		TxHash string `json:"txHash"`
		Error  string `json:"error,omitempty"`
		Revert string `json:"revert,omitempty"`
	} `json:"results"`
}

// RPC Method Params

// EthSendBundleParams represents the parameters for eth_sendBundle.
// txs: Array of signed transactions to execute in an atomic bundle
// blockNumber: Hex-encoded block number for which this bundle is valid
// minTimestamp: (Optional) Minimum timestamp for which this bundle is valid, in seconds since unix epoch
// maxTimestamp: (Optional) Maximum timestamp for which this bundle is valid, in seconds since unix epoch
// revertingTxHashes: (Optional) Array of tx hashes that are allowed to revert
// replacementUuid: (Optional) UUID that can be used to cancel/replace this bundle
// builders: (Optional) Array of registered block builder names to share the bundle with
type EthSendBundleParams struct {
	Txs               []string `json:"txs"`
	BlockNumber       string   `json:"blockNumber"`
	MinTimestamp      *int64   `json:"minTimestamp,omitempty"`
	MaxTimestamp      *int64   `json:"maxTimestamp,omitempty"`
	RevertingTxHashes []string `json:"revertingTxHashes,omitempty"`
	ReplacementUuid   *string  `json:"replacementUuid,omitempty"`
	Builders          []string `json:"builders,omitempty"`
}

// MevSendBundleRefund represents a refund configuration in validity.
type MevSendBundleRefund struct {
	BodyIdx int     `json:"bodyIdx"`
	Percent float64 `json:"percent"`
}

// MevSendBundleRefundConfig represents a refund configuration by address.
type MevSendBundleRefundConfig struct {
	Address string  `json:"address"`
	Percent float64 `json:"percent"`
}

// MevSendBundleParams represents the parameters for mev_sendBundle.
// Uses a new bundle format to send bundles to MEV-Share.
type MevSendBundleParams struct {
	Version   string                  `json:"version"` // "v0.1"
	Inclusion mevSendBundleInclusion  `json:"inclusion"`
	Body      []mevSendBundleBodyItem `json:"body"`
	Validity  *MevSendBundleValidity  `json:"validity,omitempty"`
	Privacy   *MevSendBundlePrivacy   `json:"privacy,omitempty"`
	Metadata  *MevSendBundleMetadata  `json:"metadata,omitempty"`
}

// EthCallBundleParams represents the parameters for eth_callBundle.
// txs: Array of signed transactions to execute in an atomic bundle
// blockNumber: Hex-encoded block number for which this bundle is valid
// stateBlockNumber: Hex-encoded number or block tag (e.g., "latest") for which state to base simulation on
// timestamp: (Optional) Timestamp to use for bundle simulation, in seconds since unix epoch
type EthCallBundleParams struct {
	Txs              []string `json:"txs"`
	BlockNumber      string   `json:"blockNumber"`
	StateBlockNumber string   `json:"stateBlockNumber"`
	Timestamp        *int64   `json:"timestamp,omitempty"`
}

// --- MEV-Share Simulation Types ---

// MevSimArgs is the wrapper for params
type MevSimArgs struct {
	Version   string       `json:"version"` // Usually "v0.1"
	Inclusion MevInclusion `json:"inclusion"`
	Body      []MevBody    `json:"body"`
	Privacy   *MevPrivacy  `json:"privacy,omitempty"` // Optional
}

type MevInclusion struct {
	Block    string `json:"block"`              // Target Block (Hex)
	MaxBlock string `json:"maxBlock,omitempty"` // Optional
}

type MevBody struct {
	Tx        string `json:"tx,omitempty"`        // Signed Tx Hex
	Hash      string `json:"hash,omitempty"`      // Pending Tx Hash (for backrunning)
	CanRevert bool   `json:"canRevert,omitempty"` // Default false
}

type MevPrivacy struct {
	Hints []string `json:"hints,omitempty"` // e.g. ["calldata", "logs"]
}

// MevSimResponse captures the detailed output
type MevSimResponse struct {
	Success         bool          `json:"success"`
	StateBlock      string        `json:"stateBlock"`
	MevGasPrice     string        `json:"mevGasPrice"`
	Profit          string        `json:"profit"`
	RefundableValue string        `json:"refundableValue"`
	GasUsed         string        `json:"gasUsed"`
	Logs            []TxLogResult `json:"logs,omitempty"` // <--- The best part
}

// EthCancelBundleParams represents the parameters for eth_cancelBundle.
// replacementUuid: UUID of the bundle to cancel
type EthCancelBundleParams struct {
	ReplacementUuid string `json:"replacementUuid"`
}

// EthSendPrivateTransactionParams represents the parameters for eth_sendPrivateTransaction.
type EthSendPrivateTransactionParams struct {
	Tx             string  `json:"tx"`
	MaxBlockNumber *string `json:"maxBlockNumber,omitempty"`
}

// EthSendPrivateRawTransactionParams represents the parameters for eth_sendPrivateRawTransaction.
type EthSendPrivateRawTransactionParams struct {
	Tx             string  `json:"tx"`
	MaxBlockNumber *string `json:"maxBlockNumber,omitempty"`
}

// EthCancelPrivateTransactionParams represents the parameters for eth_cancelPrivateTransaction.
type EthCancelPrivateTransactionParams struct {
	TxHash string `json:"txHash"`
}

// FlashbotsGetFeeRefundTotalsByRecipientParams represents the parameters for flashbots_getFeeRefundTotalsByRecipient.
type FlashbotsGetFeeRefundTotalsByRecipientParams struct {
	Recipient string `json:"recipient"`
}

// FlashbotsGetFeeRefundsByRecipientParams represents the parameters for flashbots_getFeeRefundsByRecipient.
type FlashbotsGetFeeRefundsByRecipientParams struct {
	Recipient string `json:"recipient"`
}

// FlashbotsGetFeeRefundsByBundleParams represents the parameters for flashbots_getFeeRefundsByBundle.
type FlashbotsGetFeeRefundsByBundleParams struct {
	BundleHash string `json:"bundleHash"`
}

// FlashbotsGetFeeRefundsByBlockParams represents the parameters for flashbots_getFeeRefundsByBlock.
type FlashbotsGetFeeRefundsByBlockParams struct {
	BlockNumber string `json:"blockNumber"`
}

// FlashbotsSetFeeRefundRecipientParams represents the parameters for flashbots_setFeeRefundRecipient.
type FlashbotsSetFeeRefundRecipientParams struct {
	Recipient string `json:"recipient"`
}

// BuildernetGetDelayedRefundsParams represents the parameters for buildernet_getDelayedRefunds.
// recipient: Address that receives delayed refunds
// blockRangeFrom: (Optional) Hex-encoded block number for start of range (inclusive)
// blockRangeTo: (Optional) Hex-encoded block number for end of range (inclusive)
// cursor: (Optional) Cursor to continue from
// hash: (Optional) Bundle hash; if provided, must also set both blockRangeFrom and blockRangeTo
type BuildernetGetDelayedRefundsParams struct {
	Recipient      string  `json:"recipient"`
	BlockRangeFrom *string `json:"blockRangeFrom,omitempty"`
	BlockRangeTo   *string `json:"blockRangeTo,omitempty"`
	Cursor         *string `json:"cursor,omitempty"`
	Hash           *string `json:"hash,omitempty"`
}

// BuildernetGetDelayedRefundTotalsByRecipientParams represents the parameters for buildernet_getDelayedRefundTotalsByRecipient.
// recipient: Address to query for delayed refunds
// blockRangeFrom: (Optional) Hex-encoded block number for start of range (inclusive)
// blockRangeTo: (Optional) Hex-encoded block number for end of range (inclusive)
type BuildernetGetDelayedRefundTotalsByRecipientParams struct {
	Recipient      string  `json:"recipient"`
	BlockRangeFrom *string `json:"blockRangeFrom,omitempty"`
	BlockRangeTo   *string `json:"blockRangeTo,omitempty"`
}

// FlashbotsGetMevRefundTotalByRecipientParams represents the parameters for flashbots_getMevRefundTotalByRecipient.
// Returns the total amount of MEV refunds paid to a recipient address. Does not require authentication.
type FlashbotsGetMevRefundTotalByRecipientParams struct {
	Recipient string `json:"recipient"`
}

// FlashbotsGetMevRefundTotalBySenderParams represents the parameters for flashbots_getMevRefundTotalBySender.
// Returns the total amount of MEV refunds generated on transactions/bundles from a sender address.
// The sender is tx.origin for individual transactions or bundles of size 1,
// or the Flashbots signer for bundles of size > 1. Does not require authentication.
type FlashbotsGetMevRefundTotalBySenderParams struct {
	Sender string `json:"sender"`
}

type TxLogResult struct {
	TxLogs []LogEntry `json:"txLogs,omitempty"`
}

type LogEntry struct {
	Address string   `json:"address"`
	Topics  []string `json:"topics"`
	Data    string   `json:"data"`
}

type BroadcastResponse struct {
	BundleHash string `json:"bundleHash"`
	Smart      bool   `json:"smart"`
}
