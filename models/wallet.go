package models

import "crypto/rsa"

type Wallet struct {
	Key *rsa.PrivateKey
}


