package tx

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"github.com/cfschilham/dullhash"
	"github.com/cfschilham/kophos/wallet"
	"math/big"
	"os"
	"strconv"
)

type Status int
const (
	Valid Status = iota + 1
	NotSigned
	InsufficientBalance
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

func (tx *Tx) Status(sender wallet.Wallet) Status {
	hash := tx.Hash()
	s, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(tx.Sender)
	if err != nil {
		// TODO
		panic(err)
	}
	err = rsa.VerifyPKCS1v15(&rsa.PublicKey{
		N: big.NewInt(0).SetBytes(s),
		E: 65537,
	}, 0, hash[:], tx.Sig)

	if err != nil {
		return NotSigned
	}

	if tx.Amount > sender.Balance {
		return InsufficientBalance
	}

	return Valid
}

func Create(sender, recip, amountStr string, seq uint64, childHash [32]byte) *Tx {
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		fmt.Printf("invalid amount\n")
		os.Exit(1)
	}

	return &Tx{
		Header: Header{
			Sender:    sender,
			Recipient: recip,
			Amount:    uint64(amount),
			Seq: seq,
			ChildHash: childHash,
		},
	}
}

func (s Status) String() string {
	seasons := [...]string{"valid", "invalid", "insufficient balance"}
	if s < Valid || s > InsufficientBalance {
		return fmt.Sprintf("Status(%d)", int(s))
	}
	return seasons[s-1]
}