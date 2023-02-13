package node

import (
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/vazj/blocker/proto"
	"github.com/vazj/blocker/types"
)

type UTXOStorer interface {
	Put(*UTXO) error
	Get(string) (*UTXO, error)
}

type MemoryUTXOStore struct {
	lock sync.RWMutex
	data map[string]*UTXO
}

func NewMemoryUTXOStore() *MemoryUTXOStore {
	return &MemoryUTXOStore{
		data: make(map[string]*UTXO),
	}
}

func (s *MemoryUTXOStore) Get(hash string) (*UTXO, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	utxo, ok := s.data[hash]
	if !ok {
		return nil, fmt.Errorf("utxo with hash[%s] doesn't exist", hash)
	}
	return utxo, nil
}

func (s *MemoryUTXOStore) Put(utxo *UTXO) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	key := fmt.Sprintf("%s_%d", utxo.Hash, utxo.OutIndex)
	s.data[key] = utxo
	return nil
}

type TXStorer interface {
	Put(*proto.Transaction) error
	Get(string) (*proto.Transaction, error)
}

type MemoryTXStore struct {
	lock sync.RWMutex
	txx  map[string]*proto.Transaction
}

func NewMemoryTXStore() *MemoryTXStore {
	return &MemoryTXStore{
		txx: make(map[string]*proto.Transaction),
	}
}

func (s *MemoryTXStore) Get(hash string) (*proto.Transaction, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	tx, ok := s.txx[hash]
	if !ok {
		return nil, fmt.Errorf("tx with hash[%s] doesn't exist", hash)
	}
	return tx, nil
}

func (s *MemoryTXStore) Put(tx *proto.Transaction) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.txx[hex.EncodeToString(types.HashTransaction(tx))] = tx
	return nil
}

type BlockStorer interface {
	Put(*proto.Block) error
	Get(string) (*proto.Block, error)
}

type MemoryBlockStore struct {
	lock   sync.RWMutex
	blocks map[string]*proto.Block
}

func NewMemoryBlockStore() *MemoryBlockStore {
	return &MemoryBlockStore{
		blocks: make(map[string]*proto.Block),
	}
}

func (s *MemoryBlockStore) Get(hash string) (*proto.Block, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	block, ok := s.blocks[hash]
	if !ok {
		return nil, fmt.Errorf("block with hash[%s] doesn't exist", hash)
	}
	return block, nil
}

func (s *MemoryBlockStore) Put(block *proto.Block) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	hash := hex.EncodeToString(types.HashBlock(block))
	s.blocks[hash] = block
	return nil
}
