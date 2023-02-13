package types

import (
	"crypto/sha256"

	pb "github.com/golang/protobuf/proto"
	"github.com/vazj/blocker/crypto"
	"github.com/vazj/blocker/proto"
)

func SignTransaction(pk *crypto.PrivateKey, tx *proto.Transaction) *crypto.Signature {
	return pk.Sign(HashTransaction(tx))
}

func HashTransaction(tx *proto.Transaction) []byte {
	b, err := pb.Marshal(tx)
	if err != nil {
		panic(err)
	}
	hash := sha256.Sum256(b)
	return hash[:]
}

func VerifyTransaction(tx *proto.Transaction) bool {
	for _, input := range tx.Inputs {
		if len(input.PublicKey) != crypto.PubKeyLen ||
			len(input.Signature) != crypto.SignatureLen {
			panic("invalid public key or signature length")
			//return false
		}
		var (
			sig    = crypto.SignatureFromBytes(input.Signature)
			pubKey = crypto.PublicKeyFromBytes(input.PublicKey)
		)

		// TODO: We need to set the signature to nil before hashing the transaction
		// because the signature is part of the transaction.
		// We don't want to hash the signature.
		// make sure we don't run into problems with this
		input.Signature = nil
		if !sig.Verify(pubKey, HashTransaction(tx)) {
			return false
		}
	}
	return true
}
