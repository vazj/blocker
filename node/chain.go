package node

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/vazj/blocker/crypto"
	"github.com/vazj/blocker/proto"
	"github.com/vazj/blocker/types"
)

const godSeed = "f7b2e105abbf7b30cefc49019386f498ecc40e1db5472b7875fa223ead7c9389"

type HeaderList struct {
	headers []*proto.Header
}

func NewHeaderList() *HeaderList {
	return &HeaderList{
		headers: make([]*proto.Header, 0),
	}
}

func (hl *HeaderList) Add(header *proto.Header) {
	hl.headers = append(hl.headers, header)
}

func (hl *HeaderList) Get(index int) *proto.Header {
	if index > hl.Height() {
		panic("index is greater than the height of the list")
	}
	return hl.headers[index]
}

func (hl *HeaderList) Len() int {
	return len(hl.headers)
}

// Height returns the height of the last header in the list
func (hl *HeaderList) Height() int {
	return hl.Len() - 1
}

type UTXO struct {
	Hash     string
	OutIndex int
	Amount   int64
	Spent    bool
}

type Chain struct {
	txStore    TXStorer
	utxStore   UTXOStorer
	blockStore BlockStorer
	headers    *HeaderList
}

func NewChain(blockStorer BlockStorer, txStore TXStorer) *Chain {
	chain := &Chain{
		txStore:    txStore,
		utxStore:   NewMemoryUTXOStore(), //TODO to pass as parameter
		blockStore: blockStorer,
		headers:    NewHeaderList(),
	}
	chain.addBlock(createGenesisBlock())
	return chain
}

// Height will always be at least 0 given the Genesis block
func (c *Chain) Height() int {
	return c.headers.Height()
}

func (c *Chain) AddBlock(b *proto.Block) error {
	if err := c.ValidateBlock(b); err != nil {
		return err
	}
	return c.addBlock(b)
}

func (c *Chain) addBlock(b *proto.Block) error {
	c.headers.Add(b.Header)
	for _, tx := range b.Transactions {
		fmt.Println("adding tx", hex.EncodeToString(types.HashTransaction(tx)))
		if err := c.txStore.Put(tx); err != nil {
			return err
		}
		for it, output := range tx.Outputs {
			hash := hex.EncodeToString(types.HashTransaction(tx))
			utxo := &UTXO{
				Hash:     hash,
				OutIndex: it,
				Amount:   output.Amount,
				Spent:    false,
			}
			if err := c.utxStore.Put(utxo); err != nil {
				return err
			}
		}
	}
	return c.blockStore.Put(b)
}

func (c *Chain) GetBlockByHeight(height int) (*proto.Block, error) {
	if height > c.Height() {
		return nil, fmt.Errorf("height[%d] is greater than the chain height[%d]", height, c.Height())
	}
	header := c.headers.Get(height)
	hash := types.HashHeader(header)
	return c.GetBlockByHash(hash)
}

func (c *Chain) GetBlockByHash(hash []byte) (*proto.Block, error) {
	hashHex := hex.EncodeToString(hash)
	return c.blockStore.Get(hashHex)
}

func (c *Chain) ValidateBlock(b *proto.Block) error {
	// validate the signature of the block
	if !types.VerifyBlock(b) {
		return fmt.Errorf("block's signature is invalid")
	}

	// validate if the prev hash of the block is the actual hash of the previous block
	currentBlock, err := c.GetBlockByHeight(c.Height())
	if err != nil {
		return err
	}
	hash := types.HashBlock(currentBlock)
	if !bytes.Equal(b.Header.PrevHash, hash) {
		return fmt.Errorf("block's previous hash doesn't match the current block's hash")
	}

	// validate the transactions
	for _, tx := range b.Transactions {
		if err := c.ValidateTransaction(tx); err != nil {
			return err
		}
	}
	return nil
}

func (c *Chain) ValidateTransaction(tx *proto.Transaction) error {
	// validate the signature of the transaction
	if !types.VerifyTransaction(tx) {
		return fmt.Errorf("transaction's signature is invalid")
	}
	// check if all inputs are unspent
	nInputs := len(tx.Inputs)
	sumInputs := int64(0)
	for i := 0; i < nInputs; i++ {
		prevHash := hex.EncodeToString(tx.Inputs[i].PrevTxHash)
		key := fmt.Sprintf("%s_%d", prevHash, i)
		utxo, err := c.utxStore.Get(key)
		if err != nil {
			return err
		}
		sumInputs += utxo.Amount
		if utxo.Spent {
			return fmt.Errorf("input at index %d of this transaction %s already spent", i, prevHash)
		}
	}

	// check if the sum of the inputs is greater than the sum of the outputs
	sumOutputs := int64(0)
	for _, output := range tx.Outputs {
		sumOutputs += output.Amount
	}

	if sumInputs < sumOutputs {
		return fmt.Errorf("insufficient balance got %d, spending %d", sumInputs, sumOutputs)
	}

	return nil
}

func createGenesisBlock() *proto.Block {
	privKey := crypto.NewPrivateKeyFromSeedStr(godSeed)
	block := &proto.Block{
		Header: &proto.Header{
			Version: 1,
		},
	}
	tx := &proto.Transaction{
		Version: 1,
		Inputs:  []*proto.TxInput{},
		Outputs: []*proto.TxOutput{
			{
				Amount:  1000,
				Address: privKey.PublicKey().Address().Bytes(),
			},
		},
	}

	block.Transactions = append(block.Transactions, tx)

	types.SignBlock(privKey, block)
	return block
}
