package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
)

const (
	// KeySize is the size of a public or private key in bytes.
	KeySize = ed25519.PublicKeySize
	// signatureLen is the size of a signature in bytes.
	SignatureLen = ed25519.SignatureSize // 64
	PrivKeyLen   = 64
	PubKeyLen    = 32
	SeedLen      = 32
	AddressLen   = 20
)

type PrivateKey struct {
	key ed25519.PrivateKey
}

type PublicKey struct {
	key ed25519.PublicKey
}

func GenerateKey() (*PrivateKey, *PublicKey) {
	priv := NewPrivateKey()
	pub := priv.Public()
	return priv, pub
}

func GeneratePrivateKey() *PrivateKey {
	seed := make([]byte, SeedLen)
	_, err := rand.Read(seed)
	if err != nil {
		panic(err)
	}
	return NewPrivateKeyFromSeed(seed)
}

func GenerateKeyFromSeed(seed []byte) (*PrivateKey, *PublicKey) {
	priv := NewPrivateKeyFromSeed(seed)
	pub := priv.Public()
	return priv, pub
}

func NewPrivateKeyFromString(s string) *PrivateKey {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return NewPrivateKeyFromSeed(b)
}

func NewPrivateKeyFromSeed(seed []byte) *PrivateKey {
	if len(seed) != SeedLen {
		panic("invalid seed length")
	}
	key := ed25519.NewKeyFromSeed(seed)
	return &PrivateKey{key: key}
}

func NewPrivateKeyFromSeedStr(s string) *PrivateKey {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return NewPrivateKeyFromSeed(b)
}

func NewPrivateKey() *PrivateKey {
	_, key, _ := ed25519.GenerateKey(nil)
	return &PrivateKey{key: key}
}

func (p *PrivateKey) Bytes() []byte {
	return p.key
}

func (p *PrivateKey) Sign(msg []byte) *Signature {
	return &Signature{value: ed25519.Sign(p.key, msg)}
}

func (p *PrivateKey) String() string {
	return string(p.key)
}

func (p *PrivateKey) Equals(other *PrivateKey) bool {
	return p.key.Equal(other.key)
}

func (p *PrivateKey) Public() *PublicKey {
	b := make([]byte, PubKeyLen)
	copy(b, p.key[PrivKeyLen-PubKeyLen:])
	return &PublicKey{key: b}
}

func (p *PrivateKey) PublicKey() *PublicKey {
	return &PublicKey{key: p.key.Public().(ed25519.PublicKey)}
}

func (p *PublicKey) Bytes() []byte {
	return p.key
}

func PublicKeyFromBytes(b []byte) *PublicKey {
	if len(b) != PubKeyLen {
		panic("invalid public key length")
	}
	return &PublicKey{key: ed25519.PublicKey(b)}
}

func (p *PublicKey) Verify(msg []byte, sig []byte) bool {
	return ed25519.Verify(p.key, msg, sig)
}

func (p *PublicKey) String() string {
	return string(p.key)
}

func (p *PublicKey) Equals(other *PublicKey) bool {
	return p.key.Equal(other.key)
}

func (p *PublicKey) Address() *Address {
	return &Address{value: p.key[PubKeyLen-AddressLen:]}
}

type Signature struct {
	value []byte
}

func (s *Signature) Bytes() []byte {
	return s.value
}

func (s *Signature) String() string {
	return string(s.value)
}

func SignatureFromBytes(b []byte) *Signature {
	if len(b) != SignatureLen {
		panic("invalid signature length")
	}
	return &Signature{value: b}
}

func (s *Signature) Verify(pubKey *PublicKey, msg []byte) bool {
	return ed25519.Verify(pubKey.key, msg, s.value)
}

type Address struct {
	value []byte
}

func (a *Address) Bytes() []byte {
	return a.value
}

func (a *Address) String() string {
	return hex.EncodeToString(a.value)
}

func AddressFromBytes(b []byte) *Address {
	if len(b) != AddressLen {
		panic("invalid address length")
	}
	return &Address{value: b}
}
