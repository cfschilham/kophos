package models

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/binary"
	"fmt"
	"github.com/cfschilham/dullhash"
)

type Header struct {
	Sender    string
	Recipient string
	Amount    uint64
	Seq		  uint64
	ChildHash [32]byte
}

type Tx struct {
	Header
	Sig []byte
}


func (tx *Tx) Sign(key *rsa.PrivateKey) (*Tx, error) {
	if tx.Sig != nil {
		return tx, fmt.Errorf("transaction already signed")
	}
	hash := tx.Hash()
	sig, err := key.Sign(rand.Reader, hash[:], crypto.Hash(0))
	if err != nil {
		return tx, err
	}
	tx.Sig = sig
	return tx, nil
}

func (tx *Tx) Hash() [32]byte {
	return dullhash.Sum(tx.Bytes())
}

func (tx *Tx) Bytes() []byte {
	out := []byte(tx.Sender)
	out = append(out, []byte(tx.Recipient)...)
	amountbe := make([]byte, 8)
	binary.BigEndian.PutUint64(amountbe, tx.Amount)
	out = append(out, amountbe...)
	out = append(out, tx.ChildHash[:]...)
	return out
}