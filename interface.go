package flashbot

import (
	"context"
	"math/big"
)

// IFlashBot defines the standard behavior for a MEV/Flashbots client.
// It unifies Simulation (Relay) and Broadcasting (Builders).
type IFlashbot interface {

	// Simulate runs the bundle against the Flashbots Relay to check for reverts.
	// This should ALWAYS be called before Broadcast.
	// blockNumber: The target block you want to land in.
	// stateBlock: The block state to simulate on (usually target - 1).
	Simulate(ctx context.Context, bundle *Bundle, targetBlock uint64, opts ...BundleOption) (*SimulateResponse, error)

	// Broadcast sends the bundle to the configured list of builders (Titan, Beaver, Flashbots, etc.).
	// It returns the list of builders that accepted the request.
	Broadcast(ctx context.Context, bundle *Bundle, targetBlock uint64, opts ...BundleOption) (*BroadcastResponse, error)

	// SendPrivateTransaction sends a single transaction directly to builders (eth_sendPrivateTransaction).
	// Useful for simple transfers where you don't need a full bundle but want frontrunning protection.
	// expDurationBlocks: The expected duration of the transaction in blocks. max 25 blocks. default 25 blocks.
	SendPrivateTransaction(ctx context.Context, signedTxHex string, expDurationBlocks uint64) error

	// --- Gas & Network Intelligence ---

	// GetGasPrice returns the suggested gas price.
	// For EIP-1559 chains, this should return the BaseFee + PriorityFee.
	// You can implement this to return a "fast" price for aggressive inclusion.
	GetGasPrice(ctx context.Context) (gasPrice *big.Int, tip *big.Int, err error)

	// EstimateGasBundle calculates the gas units required for the entire bundle.
	// This is useful for calculating exactly how much "sponsorship" ETH to send the user.
	EstimateGasBundle(ctx context.Context, bundle *Bundle) (uint64, error)

	// --- Utilities ---

	// GetUserStats checks your signing key's reputation on the relay.
	GetUserStats(ctx context.Context, blockNumber *big.Int) (*UserStats, error)
}
