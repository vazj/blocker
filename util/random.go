package util

import (
	randc "crypto/rand"
	"math/rand"
	"time"

	"github.com/vazj/blocker/proto"
)

func RandomHash() []byte {
	hash := make([]byte, 32)
	randc.Read(hash)
	return hash
}

func RandomBlock() *proto.Block {
	header := &proto.Header{
		Version:   1,
		Height:    int32(rand.Intn(1000) + 1),
		PrevHash:  RandomHash(),
		RootHash:  RandomHash(),
		Timestamp: time.Now().UnixNano(),
	}

	return &proto.Block{Header: header}
}
