package node

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vazj/blocker/crypto"
	"github.com/vazj/blocker/proto"
	"github.com/vazj/blocker/types"
	"github.com/vazj/blocker/util"
)

func randomBlock(t *testing.T, chain *Chain) *proto.Block {
	privKey := crypto.GeneratePrivateKey()
	block := util.RandomBlock()
	prevBlock, err := chain.GetBlockByHeight(chain.Height())
	require.NoError(t, err)
	block.Header.PrevHash = types.HashBlock(prevBlock)
	types.SignBlock(privKey, block)
	return block
}

func TestNewChain(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
	assert.NotNil(t, chain)
	assert.Equal(t, 0, chain.Height())
	_, err := chain.GetBlockByHeight(0)
	assert.NoError(t, err)
}

// Height will always be at least 0 given the Genesis block
func TestChainHeight(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
	assert.Equal(t, 0, chain.Height())
	for i := 0; i < 10; i++ {
		block := randomBlock(t, chain)
		require.NoError(t, chain.AddBlock(block))
		require.Equal(t, chain.Height(), i+1)
	}
}

func TestGetBlockByHash(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
	block := util.RandomBlock()
	blockHash := types.HashBlock(block)
	_, err := chain.GetBlockByHash(blockHash)
	assert.Error(t, err)
}

func TestAddBlock(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
	for i := 0; i < 10; i++ {
		block := randomBlock(t, chain)
		blockHash := types.HashBlock(block)
		require.NoError(t, chain.AddBlock(block))

		fetchedBlock, err := chain.GetBlockByHash(blockHash)
		require.NoError(t, err)
		assert.Equal(t, block, fetchedBlock)

		fetchedBlockByHeight, err := chain.GetBlockByHeight(i + 1)
		require.NoError(t, err)
		require.Equal(t, block, fetchedBlockByHeight)
	}
}

func TestAddBlockWithTxInsufficientFunds(t *testing.T) {
	var (
		chain     = NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
		block     = randomBlock(t, chain)
		privKey   = crypto.NewPrivateKeyFromSeedStr(godSeed)
		pubKey    = privKey.Public()
		recipient = crypto.GeneratePrivateKey().Public().Address().Bytes()
	)

	prevTx, err := chain.txStore.Get("efb7e770f59b223434e32e41c2ed21c56ab6081fef99d7a16c9b1aac278c4fb5")
	require.NoError(t, err)

	inputs := []*proto.TxInput{
		{
			PrevTxHash:   types.HashTransaction(prevTx),
			PrevOutIndex: 0,
			PublicKey:    pubKey.Bytes(),
		},
	}
	outputs := []*proto.TxOutput{
		{
			Address: recipient,
			Amount:  10001,
		},
	}
	tx := &proto.Transaction{
		Inputs:  inputs,
		Outputs: outputs,
		Version: 1,
	}
	sig := types.SignTransaction(privKey, tx)
	tx.Inputs[0].Signature = sig.Bytes()
	block.Transactions = append(block.Transactions, tx)
	require.Error(t, chain.AddBlock(block))

}

func TestAddBlockWithTx(t *testing.T) {
	var (
		chain     = NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
		block     = randomBlock(t, chain)
		privKey   = crypto.NewPrivateKeyFromSeedStr(godSeed)
		pubKey    = privKey.Public()
		recipient = crypto.GeneratePrivateKey().Public().Address().Bytes()
	)

	prevTx, err := chain.txStore.Get("efb7e770f59b223434e32e41c2ed21c56ab6081fef99d7a16c9b1aac278c4fb5")
	require.NoError(t, err)

	inputs := []*proto.TxInput{
		{
			PrevTxHash:   types.HashTransaction(prevTx),
			PrevOutIndex: 0,
			PublicKey:    pubKey.Bytes(),
		},
	}
	outputs := []*proto.TxOutput{
		{
			Address: recipient,
			Amount:  100,
		},
		{
			Amount:  900,
			Address: privKey.Public().Address().Bytes(),
		},
	}
	tx := &proto.Transaction{
		Inputs:  inputs,
		Outputs: outputs,
		Version: 1,
	}
	sig := types.SignTransaction(privKey, tx)
	tx.Inputs[0].Signature = sig.Bytes()
	block.Transactions = append(block.Transactions, tx)
	types.SignBlock(privKey, block)
	require.NoError(t, chain.AddBlock(block))

}
