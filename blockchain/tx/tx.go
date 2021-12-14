package tx

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/binary"
	"fmt"
	"github.com/cfschilham/dullhash"
	"math/big"
)

type Header struct {
	Sender    [64]byte
	Recipient [64]byte
	Amount    uint64
	Seq       uint64
	PrevHash  [32]byte
}

type Tx struct {
	Header
	Sig []byte
}

func (h Header) Hash() [32]byte {
	return dullhash.Sum(h.Bytes())
}

func (h Header) Bytes() []byte {
	out := h.Sender[:]
	amount, seq := make([]byte, 8), make([]byte, 8)
	binary.BigEndian.PutUint64(amount, h.Amount)
	binary.BigEndian.PutUint64(seq, h.Seq)
	out = append(out, h.Recipient[:]...)
	out = append(out, amount...)
	out = append(out, seq...)
	out = append(out, h.PrevHash[:]...)
	return out
}

func (tx *Tx) Bytes() []byte {
	return append(tx.Header.Bytes(), tx.Sig...)
}

func (tx *Tx) Hash() [32]byte {
	return dullhash.Sum(tx.Bytes())
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

func (tx *Tx) HasValidSig() bool {
	hash := tx.Hash()
	err := rsa.VerifyPKCS1v15(&rsa.PublicKey{
		N: big.NewInt(0).SetBytes(tx.Sender[:]),
		E: 65537,
	}, 0, hash[:], tx.Sig)
	return err == nil
}
