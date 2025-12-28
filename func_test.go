package flashbot

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/harpy-wings/flashbot/testutils/erc20ex"
	"github.com/stretchr/testify/require"
)

func TestSimulate(t *testing.T) {
	ethC, err := ethclient.Dial("https://cool-flashy-pallet.ethereum-sepolia.quiknode.pro/4145abfcfdc072adfcf91ec960c4b6c92579d096")
	require.NoError(t, err)

	const ERC20TokenAddress = "0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238"
	const ERC20Destanation = "0x4bfD011E2bE77b57A42882f2e854a235a7D18646"

	fb, err := New(context.Background(), WithChainID(SepoliaChainID), WithRelayURL(SepoliaRelayURL))
	require.NoError(t, err)
	ethWallet, err := crypto.HexToECDSA(os.Getenv("ETH_WALLET_PK"))
	require.NoError(t, err)
	ercWallet, err := crypto.HexToECDSA(os.Getenv("ERC_WALLET_PK"))
	require.NoError(t, err)
	fmt.Println("ETH Wallet Address:", crypto.PubkeyToAddress(ethWallet.PublicKey).Hex())
	fmt.Println("ERC Wallet Address:", crypto.PubkeyToAddress(ercWallet.PublicKey).Hex())

	var bundle = &Bundle{
		Transactions: make([]*types.Transaction, 0),
		CanRevert:    make([]bool, 0),
	}
	gasTip, err := ethC.SuggestGasTipCap(context.Background())
	require.NoError(t, err)
	gasPrice, err := ethC.SuggestGasPrice(context.Background())
	require.NoError(t, err)

	var tx2GasCost *big.Int // Will store the gas cost for tx2

	// tx 2
	{
		TokenAddress := common.HexToAddress(ERC20TokenAddress)
		// Get ABI and pack the transfer function call
		abi, err := erc20ex.Erc20exMetaData.GetAbi()
		require.NoError(t, err)
		transferData, err := abi.Pack("transfer", common.HexToAddress(ERC20Destanation), big.NewInt(1))
		require.NoError(t, err)

		ercWalletNounce, err := ethC.PendingNonceAt(context.Background(), crypto.PubkeyToAddress(ercWallet.PublicKey))
		require.NoError(t, err)

		ETAGas, err := ethC.EstimateGas(context.Background(), ethereum.CallMsg{
			From: crypto.PubkeyToAddress(ercWallet.PublicKey),
			To:   &TokenAddress,
			Data: transferData,
		})
		require.NoError(t, err)
		gas := ETAGas * 125 / 100

		// Calculate gas cost for tx2: gasLimit * gasPrice
		tx2GasCost = new(big.Int).Mul(big.NewInt(int64(gas)), gasPrice)

		tx := types.NewTx(&types.DynamicFeeTx{
			ChainID:   big.NewInt(SepoliaChainID),
			Nonce:     ercWalletNounce,
			To:        &TokenAddress,
			Value:     big.NewInt(0),
			Gas:       gas,
			GasFeeCap: gasPrice,
			GasTipCap: gasTip,
			Data:      transferData,
		})
		signer := types.LatestSignerForChainID(big.NewInt(SepoliaChainID))
		signedTx, err := types.SignTx(tx, signer, ercWallet)
		require.NoError(t, err)
		bundle.Transactions = append(bundle.Transactions, signedTx)
	}

	// tx 1 - sends ETH to ercWallet to cover gas costs for tx2
	{
		nonce, err := ethC.NonceAt(context.Background(), crypto.PubkeyToAddress(ethWallet.PublicKey), nil)
		require.NoError(t, err)

		To := crypto.PubkeyToAddress(ercWallet.PublicKey)
		tx := types.NewTx(&types.DynamicFeeTx{
			ChainID:   big.NewInt(SepoliaChainID),
			Nonce:     nonce,
			To:        &To,
			Value:     tx2GasCost, // Transfer the gas cost required for tx2
			Gas:       21000,
			GasFeeCap: gasPrice,
			GasTipCap: gasTip,
			Data:      []byte{},
		})
		singer := types.LatestSignerForChainID(big.NewInt(SepoliaChainID))
		signedTx, err := types.SignTx(tx, singer, ethWallet)
		require.NoError(t, err)
		bundle.Transactions = append([]*types.Transaction{signedTx}, bundle.Transactions...)
	}

	currentBlock, err := ethC.BlockNumber(context.Background())
	require.NoError(t, err)
	resp, err := fb.Broadcast(context.Background(), bundle, currentBlock+1)
	require.NoError(t, err)
	fmt.Println("Broadcast Response: ", resp)
}
