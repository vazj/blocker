package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vazj/blocker/crypto"
	"github.com/vazj/blocker/proto"
	"github.com/vazj/blocker/util"
)

// my balance is 100 coins
// 2 outputs
// I want to send 5 coins to someone else
// I want to keep 95 coins (send 95 to ourselves)

func TestNewTransaction(t *testing.T) {
	fromPrivKey := crypto.GeneratePrivateKey()
	fromAddress := fromPrivKey.Public().Address().Bytes()

	toPrivKey := crypto.GeneratePrivateKey()
	toAddress := toPrivKey.Public().Address().Bytes()

	input := &proto.TxInput{
		PrevTxHash:   util.RandomHash(),
		PrevOutIndex: 0,
		PublicKey:    fromPrivKey.Public().Bytes(),
	}

	output1 := &proto.TxOutput{
		Amount:  5,
		Address: toAddress,
	}

	output2 := &proto.TxOutput{
		Amount:  95,
		Address: fromAddress,
	}

	tx := proto.Transaction{
		Version: 1,
		Inputs:  []*proto.TxInput{input},
		Outputs: []*proto.TxOutput{output1, output2},
	}

	// Sign the transaction
	sig := SignTransaction(fromPrivKey, &tx)
	input.Signature = sig.Bytes()

	assert.True(t, VerifyTransaction(&tx))
}
