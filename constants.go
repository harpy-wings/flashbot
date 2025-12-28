package flashbot

const (
	MainnetChainID = 1
	SepoliaChainID = 11155111

	MainnetRelayURL = "https://relay.flashbots.net"
	SepoliaRelayURL = "https://relay-sepolia.flashbots.net"
)

const (
	// methodEthSendBundle can be used to send your bundles to the Flashbots builder.
	methodEthSendBundle method = "eth_sendBundle"
	// methodMevSendBundle uses a new bundle format to send bundles to MEV-Share.
	// See the Understanding Bundles page for more information.
	methodMevSendBundle method = "mev_sendBundle"
	// methodEthCallBundle can be used to simulate a bundle against a specific block number,
	// including simulating a bundle at the top of the next block.
	methodEthCallBundle method = "eth_callBundle"
	// methodMevSimBundle simulates a bundle using the MEV-Share format.
	methodMevSimBundle method = "mev_simBundle"
	// methodEthCancleBundle cancels a previously submitted bundle.
	methodEthCancleBundle method = "eth_cancelBundle"
	// methodEthSendPrivateTransaction sends a private transaction to the Flashbots relay.
	methodEthSendPrivateTransaction method = "eth_sendPrivateTransaction"
	// methodEthSendPrivateRawTransaction sends a raw private transaction to the Flashbots relay.
	methodEthSendPrivateRawTransaction method = "eth_sendPrivateRawTransaction"
	// methodEthCanclePrivateTransaction cancels a previously submitted private transaction.
	methodEthCanclePrivateTransaction method = "eth_cancelPrivateTransaction"
	// methodFlashbotGetFeeRefundTotalsByRecipient returns the total amount of fee refunds
	// that have been earned by a specific recipient address.
	methodFlashbotGetFeeRefundTotalsByRecipient method = "flashbots_getFeeRefundTotalsByRecipient"
	// methodFlashbotGetFeeRefundsByRecipient returns detailed information about fee refunds
	// for a specific recipient address.
	methodFlashbotGetFeeRefundsByRecipient method = "flashbots_getFeeRefundsByRecipient"
	// methodFlashbotGetFeeRefundsByBundle returns fee refund information for a specific bundle hash.
	methodFlashbotGetFeeRefundsByBundle method = "flashbots_getFeeRefundsByBundle"
	// methodFlashbotGetFeeRefundsByBlock returns fee refund information for a specific block number.
	methodFlashbotGetFeeRefundsByBlock method = "flashbots_getFeeRefundsByBlock"
	// methodFlashbotSetFeeRefundRecipient sets the fee refund recipient address for future transactions.
	methodFlashbotSetFeeRefundRecipient method = "flashbots_setFeeRefundRecipient"
	// methodBuildernetGetDelayedRefunds returns detailed information about delayed refunds,
	// including bundle hash, amount, block number, status, and recipient.
	methodBuildernetGetDelayedRefunds method = "buildernet_getDelayedRefunds"
	// methodBuildernetGetDelayedRefundTotalsByRecipient returns the total amount of delayed refunds
	// that have been earned by a specific address, including pending and received amounts.
	methodBuildernetGetDelayedRefundTotalsByRecipient method = "buildernet_getDelayedRefundTotalsByRecipient"
	// methodFlashbotsGetMevRefundTotalByRecipient returns the total amount of MEV refunds
	// that have been paid to a specific recipient address. This API does not require authentication.
	methodFlashbotsGetMevRefundTotalByRecipient method = "flashbots_getMevRefundTotalByRecipient"
	// methodFlashbotsGetMevRefundTotalBySender returns the total amount of MEV refunds
	// that have been generated on transactions or bundles from a specific sender address.
	// The sender is the tx.origin for individual transactions or bundles of size 1,
	// or the Flashbots signer for bundles of size > 1. This API does not require authentication.
	methodFlashbotsGetMevRefundTotalBySender method = "flashbots_getMevRefundTotalBySender"
)

const (
	headerFlashbotSignature = "X-Flashbots-Signature"
)

const (
	jsonRPCVersion = "2.0"
)
