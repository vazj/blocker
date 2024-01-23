package types

import (
	"bytes"
	"crypto/sha256"

	"github.com/cbergoon/merkletree"
	pb "github.com/golang/protobuf/proto"
	"github.com/vazj/blocker/crypto"
	"github.com/vazj/blocker/proto"
)

type TxHash struct {
	hash []byte
}

func NewTxHash(hash []byte) TxHash {
	return TxHash{hash: hash}
}

func (t TxHash) CalculateHash() ([]byte, error) {
	return t.hash, nil
}

func (t TxHash) Equals(other merkletree.Content) (bool, error) {
	equals := bytes.Equal(t.hash, other.(TxHash).hash)
	return equals, nil
}

func VerifyBlock(b *proto.Block) bool {
	if len(b.Transactions) > 0 {
		if !VerifyRootHash(b) {
			return false
		}
	}

	if len(b.PublicKey) != crypto.PubKeyLen || len(b.Signature) != crypto.SignatureLen {
		return false
	}

	var (
		sig    = crypto.SignatureFromBytes(b.Signature)
		pubKey = crypto.PublicKeyFromBytes(b.PublicKey)
	)

	return sig.Verify(pubKey, HashBlock(b))
}

func VerifyRootHash(b *proto.Block) bool {
	t, err := GetMerkleTree(b)
	if err != nil {
		return false
	}

	valid, err := t.VerifyTree()
	if err != nil {
		return false
	}

	if len(b.Header.RootHash) == 0 {
		return false
	}

	if !valid {
		return false
	}

	return bytes.Equal(b.Header.RootHash, t.MerkleRoot())
}

func SignBlock(pk *crypto.PrivateKey, b *proto.Block) *crypto.Signature {
	if len(b.Transactions) > 0 {
		t, err := GetMerkleTree(b)
		if err != nil {
			panic(err)
		}

		b.Header.RootHash = t.MerkleRoot()
	}

	hash := HashBlock(b)
	sig := pk.Sign(hash)
	b.PublicKey = pk.Public().Bytes()
	b.Signature = sig.Bytes()

	return sig
}

func GetMerkleTree(b *proto.Block) (*merkletree.MerkleTree, error) {
	list := make([]merkletree.Content, len(b.Transactions))
	for i := 0; i < len(b.Transactions); i++ {
		list[i] = NewTxHash(HashTransaction(b.Transactions[i]))
	}
	t, err := merkletree.NewTree(list)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// Hashblock returns a single SHA256 of the block.
func HashBlock(block *proto.Block) []byte {
	return HashHeader(block.Header)
}

func HashHeader(header *proto.Header) []byte {
	b, err := pb.Marshal(header)
	if err != nil {
		panic(err)
	}
	hash := sha256.Sum256(b)
	return hash[:]
}
