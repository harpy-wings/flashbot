package flashbot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/core/types"
)

// Bundle is a struct that represents a bundle of transactions.
type Bundle struct {
	// Transactions is a list of transactions to be included in the bundle.
	Transactions []*types.Transaction
	// CanRevert is a list of booleans indicating whether each transaction can revert. len(CanRevert) == len(Transactions).
	CanRevert []bool
	// ReplacementUUID is the UUID of the bundle to replace.
	ReplacementUUID string
	// Builders is a list of builders to send the bundle to.
	Builders []string
	// MinTimestamp is the minimum timestamp for which the bundle is valid.
	MinTimestamp int64
	// MaxTimestamp is the maximum timestamp for which the bundle is valid.
}

// BundleOption is a function that can be used to configure the bundle.
type BundleOption func(*mevSimBundleParams) error

// mevSimBundleParams represents the parameters for mev_simBundle.
// Similar structure to MevSendBundleParams but for simulation.
type mevSimBundleParams struct {
	Version   string                  `json:"version"`
	Inclusion mevSendBundleInclusion  `json:"inclusion"`
	Body      []mevSendBundleBodyItem `json:"body"`
	Validity  *MevSendBundleValidity  `json:"validity,omitempty"`
	Privacy   *MevSendBundlePrivacy   `json:"privacy,omitempty"`
	Metadata  *MevSendBundleMetadata  `json:"metadata,omitempty"`
}

// mevSendBundleInclusion represents the inclusion block parameters for mev_sendBundle.
type mevSendBundleInclusion struct {
	Block    string  `json:"block"`              // Hex-encoded number
	MaxBlock *string `json:"maxBlock,omitempty"` // Hex-encoded number
}

// mevSendBundleBodyItem represents a single item in the body array for mev_sendBundle.
// It can be one of: { hash: string }, { tx: string, canRevert: boolean }, or { bundle: MevSendBundleParams }
type mevSendBundleBodyItem struct {
	Hash      *string              `json:"hash,omitempty"`
	Tx        *string              `json:"tx,omitempty"`
	CanRevert *bool                `json:"canRevert,omitempty"`
	Bundle    *MevSendBundleParams `json:"bundle,omitempty"`
}

// MevSendBundleValidity represents validity configuration for mev_sendBundle.
type MevSendBundleValidity struct {
	Refund       []MevSendBundleRefund       `json:"refund,omitempty"`
	RefundConfig []MevSendBundleRefundConfig `json:"refundConfig,omitempty"`
}

// MevSendBundlePrivacy represents privacy configuration for mev_sendBundle.
type MevSendBundlePrivacy struct {
	Hints    []string `json:"hints,omitempty"` // "calldata", "contract_address", "logs", "function_selector", "hash", "tx_hash", "full"
	Builders []string `json:"builders,omitempty"`
}

// MevSendBundleMetadata represents metadata for mev_sendBundle.
type MevSendBundleMetadata struct {
	OriginID *string `json:"originId,omitempty"`
}

// BUNDLE OPTIONS

// WithValidity sets the validity configuration for the bundle.
func WithValidity(validity MevSendBundleValidity) BundleOption {
	return func(params *mevSimBundleParams) error {
		params.Validity = &validity
		return nil
	}
}

// WithPrivacy sets the privacy configuration for the bundle.
func WithPrivacy(privacy MevSendBundlePrivacy) BundleOption {
	return func(params *mevSimBundleParams) error {
		params.Privacy = &privacy
		return nil
	}
}

// WithMetadata sets the metadata for the bundle.
func WithMetadata(metadata MevSendBundleMetadata) BundleOption {
	return func(params *mevSimBundleParams) error {
		params.Metadata = &metadata
		return nil
	}
}

func WithExpirationDurationInBlocks(duration uint64) BundleOption {
	return func(params *mevSimBundleParams) error {
		block, err := strconv.ParseUint(strings.TrimPrefix(params.Inclusion.Block, "0x"), 16, 64)
		if err != nil {
			return fmt.Errorf("failed to parse block number: %w", err)
		}
		block += duration
		maxBlock := "0x" + strconv.FormatUint(block, 16)
		params.Inclusion.MaxBlock = &maxBlock
		return nil
	}
}
func WithExpirationBlock(block uint64) BundleOption {
	return func(params *mevSimBundleParams) error {
		maxBlock := "0x" + strconv.FormatUint(block, 16)
		params.Inclusion.MaxBlock = &maxBlock
		return nil
	}
}
