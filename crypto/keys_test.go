package crypto

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePrivateKey(t *testing.T) {
	privKey := GeneratePrivateKey()
	assert.Equal(t, PrivKeyLen, len(privKey.Bytes()))
	pubKey := privKey.Public()
	assert.Equal(t, PubKeyLen, len(pubKey.Bytes()))
}

func TestGenerateKey(t *testing.T) {
	privKey, pubKey := GenerateKey()
	assert.Equal(t, PrivKeyLen, len(privKey.Bytes()))
	assert.Equal(t, PubKeyLen, len(pubKey.Bytes()))
}

func TestGenerateKeyFromSeed(t *testing.T) {
	seed := make([]byte, SeedLen)
	privKey, pubKey := GenerateKeyFromSeed(seed)
	assert.Equal(t, PrivKeyLen, len(privKey.Bytes()))
	assert.Equal(t, PubKeyLen, len(pubKey.Bytes()))
}

func TestNewPrivateKeyFromSeed(t *testing.T) {
	seed := make([]byte, SeedLen)
	privKey := NewPrivateKeyFromSeed(seed)
	assert.Equal(t, PrivKeyLen, len(privKey.Bytes()))
}

func TestNewPrivateKey(t *testing.T) {
	privKey := NewPrivateKey()
	assert.Equal(t, PrivKeyLen, len(privKey.Bytes()))
}

func TestPrivateKey_Bytes(t *testing.T) {
	privKey := NewPrivateKey()
	assert.Equal(t, PrivKeyLen, len(privKey.Bytes()))
}

func TestPrivateKey_Public(t *testing.T) {
	privKey := NewPrivateKey()
	pubKey := privKey.Public()
	assert.Equal(t, PubKeyLen, len(pubKey.Bytes()))
}

func TestPrivateKey_Sign(t *testing.T) {
	privKey := GeneratePrivateKey()
	pubKey := privKey.Public()
	msg := []byte("foo bar baz")
	sig := privKey.Sign(msg)
	assert.True(t, sig.Verify(pubKey, msg))
	assert.False(t, sig.Verify(pubKey, []byte("foo bar")))

	invalidPubKey := GeneratePrivateKey().Public()
	assert.False(t, sig.Verify(invalidPubKey, msg))
}

func TestPublicKeyToAddress(t *testing.T) {
	privKey := GeneratePrivateKey()
	pubKey := privKey.Public()
	addr := pubKey.Address()
	assert.Equal(t, AddressLen, len(addr.Bytes()))
}

func TestNewPrivateKeyFromString(t *testing.T) {
	var (
		seed       = "a127fa0a31994985c678ff53b0829dfd0d7e17b9fe65947769932127e927da17"
		privKey    = NewPrivateKeyFromString(seed)
		addressStr = "44d3cedd0e65fe6d1452566fc16ef724527f747f"
	)
	assert.Equal(t, PrivKeyLen, len(privKey.Bytes()))
	address := privKey.Public().Address()
	fmt.Print(address.String())
	assert.Equal(t, addressStr, address.String())

}
