package wallet

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
)

type Wallet struct {
	Key *rsa.PrivateKey
}

func New() (Wallet, error) {
	key, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		return Wallet{}, fmt.Errorf("error while generating keypair: %v", err)
	}
	return Wallet{Key: key}, nil
}
