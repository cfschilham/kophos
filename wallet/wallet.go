package wallet

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base32"
	"fmt"
	"math/big"
)

type Wallet struct {
	Key *rsa.PrivateKey
	Balance uint64
}

func New() (Wallet, error) {
	key, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		return Wallet{}, fmt.Errorf("error while generating keypair: %v", err)
	}
	return Wallet{Key: key}, nil
}

func Lookup(ws []Wallet, id string) (int, error) {
	n, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(id)
	if err != nil {
		return 0, fmt.Errorf("error while decoding wallet id: %v", err)
	}
	for i, w := range ws {
		if big.NewInt(0).SetBytes(n).Cmp(w.Key.PublicKey.N) == 0 {
			return i, nil
		}
	}
	return -1, nil
}
