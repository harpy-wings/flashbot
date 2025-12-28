# Flashbot Go Library

[![Go Reference](https://pkg.go.dev/badge/github.com/harpy-wings/flashbot.svg)](https://pkg.go.dev/github.com/harpy-wings/flashbot)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A comprehensive Go library for interacting with [Flashbots](https://docs.flashbots.net), enabling developers to send transaction bundles to Flashbots relay for MEV (Maximal Extractable Value) protection and optimization on Ethereum.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [API Documentation](#api-documentation)
- [Examples](#examples)
- [Configuration](#configuration)
- [Advanced Usage](#advanced-usage)
- [Network Support](#network-support)
- [Authentication](#authentication)
- [TODO & Improvements](#todo--improvements)
- [Contributing](#contributing)
- [Support](#support)
- [License](#license)

## Features

- 🚀 **Bundle Simulation**: Test your transaction bundles before broadcasting using `mev_simBundle`
- 📡 **Bundle Broadcasting**: Send transaction bundles to Flashbots relay with `mev_sendBundle`
- 🔒 **MEV Protection**: Protect your transactions from frontrunning and sandwich attacks
- ⛽ **Gas Estimation**: Built-in gas price and tip estimation utilities
- 🔧 **Flexible Configuration**: Support for multiple builders, custom relay URLs, and network selection
- 📊 **OpenTelemetry Integration**: Built-in tracing support for observability
- 🎯 **MEV-Share Support**: Full support for the latest MEV-Share bundle format
- 🔐 **EIP-191 Authentication**: Secure request signing using Ethereum private keys

## Installation

```bash
go get github.com/harpy-wings/flashbot
```

## Quick Start

### Basic Example: Sending a Bundle

```go
package main

import (
    "context"
    "crypto/ecdsa"
    "fmt"
    "math/big"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/harpy-wings/flashbot"
)

func main() {
    ctx := context.Background()
    
    // Initialize Flashbot client for Sepolia testnet
    fb, err := flashbot.New(ctx,
        flashbot.WithChainID(flashbot.SepoliaChainID),
        flashbot.WithRelayURL(flashbot.SepoliaRelayURL),
    )
    if err != nil {
        panic(err)
    }
    
    // Connect to Ethereum node
    ethClient, err := ethclient.Dial("https://your-rpc-url")
    if err != nil {
        panic(err)
    }
    
    // Get current block number
    currentBlock, err := ethClient.BlockNumber(ctx)
    if err != nil {
        panic(err)
    }
    
    // Create your transactions
    privateKey, _ := crypto.HexToECDSA("your-private-key")
    fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
    
    // Example: Create a simple ETH transfer transaction
    nonce, _ := ethClient.PendingNonceAt(ctx, fromAddress)
    gasPrice, _ := ethClient.SuggestGasPrice(ctx)
    gasTip, _ := ethClient.SuggestGasTipCap(ctx)
    
    toAddress := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
    tx := types.NewTx(&types.DynamicFeeTx{
        ChainID:   big.NewInt(flashbot.SepoliaChainID),
        Nonce:     nonce,
        To:        &toAddress,
        Value:     big.NewInt(1000000000000000000), // 1 ETH
        Gas:       21000,
        GasFeeCap: gasPrice,
        GasTipCap: gasTip,
        Data:      nil,
    })
    
    signer := types.LatestSignerForChainID(big.NewInt(flashbot.SepoliaChainID))
    signedTx, _ := types.SignTx(tx, signer, privateKey)
    
    // Create bundle
    bundle := &flashbot.Bundle{
        Transactions: []*types.Transaction{signedTx},
        CanRevert:    []bool{false},
    }
    
    // Simulate bundle first (recommended)
    simResp, err := fb.Simulate(ctx, bundle, currentBlock+1)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Simulation successful! Profit: %s\n", simResp.Profit)
    
    // Broadcast bundle
    broadcastResp, err := fb.Broadcast(ctx, bundle, currentBlock+1)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Bundle broadcasted! Hash: %s\n", broadcastResp.BundleHash)
}
```

## API Documentation

### Core Interface: `IFlashbot`

The library provides a clean interface for interacting with Flashbots:

```go
type IFlashbot interface {
    // Simulate runs the bundle against Flashbots Relay to check for reverts
    Simulate(ctx context.Context, bundle *Bundle, targetBlock uint64, opts ...BundleOption) (*SimulateResponse, error)
    
    // Broadcast sends the bundle to configured builders
    Broadcast(ctx context.Context, bundle *Bundle, targetBlock uint64, opts ...BundleOption) (*BroadcastResponse, error)
    
    // SendPrivateTransaction sends a single transaction with frontrunning protection
    SendPrivateTransaction(ctx context.Context, signedTxHex string, expDurationBlocks uint64) error
    
    // GetGasPrice returns suggested gas price and tip
    GetGasPrice(ctx context.Context) (gasPrice *big.Int, tip *big.Int, err error)
    
    // EstimateGasBundle calculates total gas units for the bundle
    EstimateGasBundle(ctx context.Context, bundle *Bundle) (uint64, error)
    
    // GetUserStats checks your signing key's reputation on the relay
    GetUserStats(ctx context.Context, blockNumber *big.Int) (*UserStats, error)
}
```

### Bundle Structure

```go
type Bundle struct {
    Transactions   []*types.Transaction  // List of signed transactions
    CanRevert      []bool                 // Whether each tx can revert
    ReplacementUUID string                 // UUID for bundle replacement
    Builders       []string               // Target builders
    MinTimestamp   int64                  // Minimum timestamp validity
    MaxTimestamp   int64                  // Maximum timestamp validity
}
```

### Response Types

#### SimulateResponse

```go
type SimulateResponse struct {
    Success         bool          // Whether simulation succeeded
    StateBlock      string        // Block used for simulation
    MevGasPrice     string        // MEV gas price
    Profit          string        // Expected profit in wei
    RefundableValue string        // Refundable value
    GasUsed         string        // Total gas used
    Logs            []TxLogResult // Transaction logs
}
```

#### BroadcastResponse

```go
type BroadcastResponse struct {
    BundleHash string // Unique bundle identifier
    Smart      bool   // Whether smart routing was used
}
```

## Examples

### Example 1: Multi-Transaction Bundle (Gas Sponsorship)

This example demonstrates creating a bundle where one transaction sponsors gas for another:

```go
package main

import (
    "context"
    "crypto/ecdsa"
    "fmt"
    "math/big"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/harpy-wings/flashbot"
)

func createGasSponsoredBundle(ctx context.Context, ethClient *ethclient.Client, fb flashbot.IFlashbot) error {
    // Wallet 1: Sponsor (has ETH)
    sponsorKey, _ := crypto.HexToECDSA("sponsor-private-key")
    sponsorAddr := crypto.PubkeyToAddress(sponsorKey.PublicKey)
    
    // Wallet 2: User (needs gas for ERC20 transfer)
    userKey, _ := crypto.HexToECDSA("user-private-key")
    userAddr := crypto.PubkeyToAddress(userKey.PublicKey)
    
    gasPrice, _ := ethClient.SuggestGasPrice(ctx)
    gasTip, _ := ethClient.SuggestGasTipCap(ctx)
    
    // Transaction 2: ERC20 transfer (user's transaction)
    tokenAddress := common.HexToAddress("0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238")
    // ... pack ERC20 transfer data ...
    userNonce, _ := ethClient.PendingNonceAt(ctx, userAddr)
    
    tx2 := types.NewTx(&types.DynamicFeeTx{
        ChainID:   big.NewInt(flashbot.SepoliaChainID),
        Nonce:     userNonce,
        To:        &tokenAddress,
        Value:     big.NewInt(0),
        Gas:       65000,
        GasFeeCap: gasPrice,
        GasTipCap: gasTip,
        Data:      transferData,
    })
    signer := types.LatestSignerForChainID(big.NewInt(flashbot.SepoliaChainID))
    signedTx2, _ := types.SignTx(tx2, signer, userKey)
    
    // Calculate gas cost for tx2
    tx2GasCost := new(big.Int).Mul(big.NewInt(65000), gasPrice)
    
    // Transaction 1: Sponsor sends ETH to user for gas
    sponsorNonce, _ := ethClient.NonceAt(ctx, sponsorAddr, nil)
    tx1 := types.NewTx(&types.DynamicFeeTx{
        ChainID:   big.NewInt(flashbot.SepoliaChainID),
        Nonce:     sponsorNonce,
        To:        &userAddr,
        Value:     tx2GasCost, // Send exact gas cost
        Gas:       21000,
        GasFeeCap: gasPrice,
        GasTipCap: gasTip,
        Data:      nil,
    })
    signedTx1, _ := types.SignTx(tx1, signer, sponsorKey)
    
    // Create bundle (tx1 must execute before tx2)
    bundle := &flashbot.Bundle{
        Transactions: []*types.Transaction{signedTx1, signedTx2},
        CanRevert:    []bool{false, false},
    }
    
    // Simulate
    currentBlock, _ := ethClient.BlockNumber(ctx)
    simResp, err := fb.Simulate(ctx, bundle, currentBlock+1)
    if err != nil {
        return err
    }
    
    if !simResp.Success {
        return fmt.Errorf("simulation failed")
    }
    
    // Broadcast
    broadcastResp, err := fb.Broadcast(ctx, bundle, currentBlock+1)
    if err != nil {
        return err
    }
    
    fmt.Printf("Bundle hash: %s\n", broadcastResp.BundleHash)
    return nil
}
```

### Example 2: Bundle with Privacy Hints

```go
func sendBundleWithPrivacy(ctx context.Context, fb flashbot.IFlashbot, bundle *flashbot.Bundle, targetBlock uint64) error {
    // Configure privacy hints to share specific information with builders
    privacy := flashbot.MevSendBundlePrivacy{
        Hints: []string{
            "calldata",        // Share calldata
            "contract_address", // Share contract address
            "logs",            // Share event logs
        },
        Builders: []string{"builder0x69"}, // Target specific builder
    }
    
    // Broadcast with privacy configuration
    resp, err := fb.Broadcast(ctx, bundle, targetBlock,
        flashbot.WithPrivacy(privacy),
    )
    if err != nil {
        return err
    }
    
    fmt.Printf("Bundle sent with privacy hints: %s\n", resp.BundleHash)
    return nil
}
```

### Example 3: Bundle with Expiration

```go
func sendBundleWithExpiration(ctx context.Context, fb flashbot.IFlashbot, bundle *flashbot.Bundle, targetBlock uint64) error {
    // Bundle expires after 10 blocks
    resp, err := fb.Broadcast(ctx, bundle, targetBlock,
        flashbot.WithExpirationDurationInBlocks(10),
    )
    if err != nil {
        return err
    }
    
    fmt.Printf("Bundle expires in 10 blocks: %s\n", resp.BundleHash)
    return nil
}
```

### Example 4: Gas Price Estimation

```go
func estimateOptimalGas(ctx context.Context, fb flashbot.IFlashbot) error {
    gasPrice, tip, err := fb.GetGasPrice(ctx)
    if err != nil {
        return err
    }
    
    fmt.Printf("Suggested gas price: %s wei\n", gasPrice.String())
    fmt.Printf("Suggested tip: %s wei\n", tip.String())
    
    // Calculate total fee for EIP-1559
    totalFee := new(big.Int).Add(gasPrice, tip)
    fmt.Printf("Total fee: %s wei\n", totalFee.String())
    
    return nil
}
```

## Configuration

### Client Options

When creating a new Flashbot client, you can configure it with various options:

```go
fb, err := flashbot.New(ctx,
    flashbot.WithChainID(flashbot.MainnetChainID),        // Set chain ID
    flashbot.WithRelayURL(flashbot.MainnetRelayURL),      // Set relay URL
    flashbot.WithBuilders([]string{"builder0x69"}),        // Target specific builders
)
```

#### Available Options

- `WithChainID(chainID uint64)`: Set the Ethereum chain ID
- `WithRelayURL(url string)`: Set custom Flashbots relay URL
- `WithBuilders(builders []string)`: Specify target block builders

### Bundle Options

Bundle options allow you to customize bundle behavior:

```go
// Privacy configuration
privacy := flashbot.MevSendBundlePrivacy{
    Hints:    []string{"calldata", "logs"},
    Builders: []string{"builder0x69"},
}

// Validity configuration (refunds)
validity := flashbot.MevSendBundleValidity{
    Refund: []flashbot.MevSendBundleRefund{
        {BodyIdx: 0, Percent: 50.0}, // 50% refund to first transaction
    },
}

// Metadata
originID := "my-origin-id"
metadata := flashbot.MevSendBundleMetadata{
    OriginID: &originID,
}

// Apply options when broadcasting
resp, err := fb.Broadcast(ctx, bundle, targetBlock,
    flashbot.WithPrivacy(privacy),
    flashbot.WithValidity(validity),
    flashbot.WithMetadata(metadata),
    flashbot.WithExpirationDurationInBlocks(25),
)
```

#### Available Bundle Options

- `WithPrivacy(privacy MevSendBundlePrivacy)`: Configure privacy hints and builder targeting
- `WithValidity(validity MevSendBundleValidity)`: Configure refund settings
- `WithMetadata(metadata MevSendBundleMetadata)`: Add metadata to bundle
- `WithExpirationDurationInBlocks(duration uint64)`: Set expiration in blocks
- `WithExpirationBlock(block uint64)`: Set specific expiration block

## Advanced Usage

### Custom HTTP Client

The library uses the default HTTP client by default. You can customize it by modifying the `flashbot` struct after creation (future enhancement).

### OpenTelemetry Tracing

The library includes built-in OpenTelemetry tracing. Ensure you have a tracer provider configured:

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

// Configure your tracer provider
tracerProvider := // ... your tracer provider
otel.SetTracerProvider(tracerProvider)
```

### Error Handling

Always check errors and handle them appropriately:

```go
simResp, err := fb.Simulate(ctx, bundle, targetBlock)
if err != nil {
    // Handle simulation error
    log.Printf("Simulation failed: %v", err)
    return err
}

if !simResp.Success {
    // Bundle would revert or fail
    log.Printf("Bundle would fail in simulation")
    return fmt.Errorf("simulation indicates failure")
}
```

## Network Support

### Supported Networks

- **Mainnet**: Chain ID `1`, Relay URL: `https://relay.flashbots.net`
- **Sepolia Testnet**: Chain ID `11155111`, Relay URL: `https://relay-sepolia.flashbots.net`

### Constants

```go
const (
    MainnetChainID = 1
    SepoliaChainID = 11155111
    
    MainnetRelayURL = "https://relay.flashbots.net"
    SepoliaRelayURL = "https://relay-sepolia.flashbots.net"
)
```

## Authentication

The library automatically handles authentication using EIP-191 signing. By default, a new private key is generated for each client instance. The signature is included in the `X-Flashbots-Signature` header.

**Note**: The signing key is used for reputation tracking. Consider using a dedicated key for Flashbots operations.

## TODO & Improvements

This section outlines planned improvements and areas where contributions are welcome:

### High Priority

- [ ] **Private Transaction Support**: Complete implementation of `SendPrivateTransaction` method
- [ ] **User Stats API**: Implement `GetUserStats` to check reputation and statistics
- [ ] **Bundle Status Tracking**: Implement `GetBundleStats` to track bundle inclusion status
- [ ] **Custom Private Key Support**: Add `WithPrivateKey` option for custom signing keys
- [ ] **Ethereum Client Integration**: Add `WithEthClient` option for custom Ethereum clients
- [ ] **Custom Logger Support**: Add `WithLogger` option for custom logging
- [ ] **Retry Logic**: Implement automatic retry for failed requests
- [ ] **Rate Limiting**: Add rate limiting awareness

### Medium Priority

- [ ] **Bundle Cancellation**: Implement `eth_cancelBundle` support
- [ ] **Fee Refund APIs**: Implement all fee refund query methods
  - `flashbots_getFeeRefundTotalsByRecipient`
  - `flashbots_getFeeRefundsByRecipient`
  - `flashbots_getFeeRefundsByBundle`
  - `flashbots_getFeeRefundsByBlock`
  - `flashbots_setFeeRefundRecipient`
- [ ] **Delayed Refund APIs**: Implement builder network refund methods
  - `buildernet_getDelayedRefunds`
  - `buildernet_getDelayedRefundTotalsByRecipient`
- [ ] **MEV Refund APIs**: Implement MEV refund query methods
  - `flashbots_getMevRefundTotalByRecipient`
  - `flashbots_getMevRefundTotalBySender`
- [ ] **Transaction Status API**: Add support for checking transaction status
- [ ] **WebSocket Support**: Add WebSocket connection for real-time updates
- [ ] **Batch Operations**: Support for sending multiple bundles in one request

### Low Priority

- [ ] **Additional Network Support**: Add support for other EVM chains
- [ ] **Bundle Optimization**: Add utilities for optimizing bundle ordering
- [ ] **Gas Price Strategies**: Implement different gas price strategies (fast, standard, slow)
- [ ] **Metrics Export**: Add Prometheus metrics export
- [ ] **Context Timeout Handling**: Improve context timeout and cancellation handling
- [ ] **Documentation**: Add more code examples and use cases
- [ ] **Performance Optimization**: Optimize request serialization and parsing

### Testing & Quality

- [ ] **Integration Tests**: Add comprehensive integration tests
- [ ] **Mock Server**: Create mock Flashbots relay server for testing
- [ ] **Test Coverage**: Increase test coverage to >80%
- [ ] **Benchmark Tests**: Add benchmark tests for performance monitoring
- [ ] **Fuzzing**: Add fuzzing tests for edge cases

### Documentation

- [ ] **API Reference**: Generate and host comprehensive API documentation
- [ ] **Tutorial Series**: Create step-by-step tutorials for common use cases
- [ ] **Video Guides**: Create video tutorials for complex scenarios
- [ ] **Best Practices Guide**: Document best practices for MEV protection

## Contributing

We welcome contributions from the community! Here's how you can help:

### Getting Started

1. **Fork the Repository**: Click the "Fork" button on GitHub
2. **Clone Your Fork**: 
   ```bash
   git clone https://github.com/your-username/flashbot.git
   cd flashbot
   ```
3. **Create a Branch**: 
   ```bash
   git checkout -b feature/your-feature-name
   ```

### Development Workflow

1. **Make Your Changes**: Implement your feature or fix
2. **Write Tests**: Add tests for new functionality
3. **Run Tests**: 
   ```bash
   go test ./...
   ```
4. **Check Linting**: 
   ```bash
   golangci-lint run
   ```
5. **Commit Changes**: 
   ```bash
   git commit -m "feat: add your feature description"
   ```
6. **Push to Your Fork**: 
   ```bash
   git push origin feature/your-feature-name
   ```
7. **Create Pull Request**: Open a PR on GitHub

### Contribution Guidelines

- **Code Style**: Follow Go conventions and use `gofmt`
- **Commit Messages**: Use conventional commits format (feat:, fix:, docs:, etc.)
- **Testing**: Ensure all tests pass and add tests for new features
- **Documentation**: Update README and add code comments for public APIs
- **Issues**: Check existing issues before creating new ones
- **Discussions**: Use GitHub Discussions for questions and proposals

### Areas Where Help is Needed

- Implementing missing API methods (see TODO section)
- Writing comprehensive tests
- Improving documentation and examples
- Performance optimization
- Adding support for additional networks
- Creating tutorials and guides

### Code Review Process

1. All PRs require at least one approval
2. Maintainers will review code for quality, tests, and documentation
3. Address review comments promptly
4. Once approved, maintainers will merge your PR

## Support

### Getting Help

- **GitHub Issues**: Report bugs or request features via [GitHub Issues](https://github.com/harpy-wings/flashbot/issues)
- **Discussions**: Ask questions in [GitHub Discussions](https://github.com/harpy-wings/flashbot/discussions)
- **Documentation**: Check the [Flashbots Documentation](https://docs.flashbots.net) for API details

### Reporting Bugs

When reporting bugs, please include:

- Go version
- Library version
- Steps to reproduce
- Expected behavior
- Actual behavior
- Error messages or logs

### Feature Requests

For feature requests, please include:

- Use case description
- Proposed API design
- Benefits and potential drawbacks
- Implementation considerations

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Related Projects

- [Flashbots Documentation](https://docs.flashbots.net) - Official Flashbots documentation
- [go-ethereum](https://github.com/ethereum/go-ethereum) - Official Go Ethereum client
- [Flashbots Protect](https://protect.flashbots.net) - Flashbots Protect RPC endpoint

## Acknowledgments

- [Flashbots](https://flashbots.net) for building the MEV infrastructure
- The Ethereum community for continuous innovation
- All contributors who help improve this library

---

**Made with ❤️ for the Ethereum and MEV community**

