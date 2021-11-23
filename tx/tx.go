package tx

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/cfschilham/dullhash"
	"github.com/cfschilham/kophos/store"
	"github.com/cfschilham/kophos/command"
	"math/big"
	"os"
	"strconv"
	"strings"
)

var CmdTx = command.Command{
	Run: runTx,
}

type Header struct {
	Sender    string
	Recipient string
	Amount    uint64
	ChildHash [32]byte
}

type Tx struct {
	Header
	Sig []byte
}

func (tx *Tx) Sign(key *rsa.PrivateKey) (*Tx, error) {
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

func (tx *Tx) Validate() bool {
	hash := tx.Hash()
	sender, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(tx.Sender)
	if err != nil {
		// TODO
		panic(err)
	}
	err = rsa.VerifyPKCS1v15(&rsa.PublicKey{
		N: big.NewInt(0).SetBytes(sender),
		E: 65537,
	}, 0, hash[:], tx.Sig)
	return err == nil
}

func create(sender, recip, amountStr string) {
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		fmt.Printf("invalid amount\n")
		os.Exit(1)
	}

	txs := []*Tx{}
	if err = store.Load(&txs, "txs"); err != nil {
		fmt.Printf("error while loading txs: %v\n", err)
		os.Exit(1)
	}
	tx := &Tx{
		Header: Header{
			Sender:    sender,
			Recipient: recip,
			Amount:    uint64(amount),
			ChildHash: [32]byte{0},
		},
	}
	txs = append(txs, tx)
	if err = store.Save(txs, "txs"); err != nil {
		fmt.Printf("error while saving txs: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("transaction created with hash %x\n", tx.Hash())
	os.Exit(0)
}

func list() {
	txs := []*Tx{}
	if err := store.Load(&txs, "txs"); err != nil {
		fmt.Printf("error while loading txs: %v\n", err)
		os.Exit(1)
	}
	for i, tx := range txs {
		fmt.Printf("%03d: %064X\n", i, tx.Hash())
	}
	os.Exit(0)
}

func validate(id string) {
	txs := []*Tx{}
	if err := store.Load(&txs, "txs"); err != nil {
		fmt.Printf("error while loading txs: %v\n", err)
		os.Exit(1)
	}
	for _, tx := range txs {
		hash := tx.Hash()
		if hex.EncodeToString(hash[:]) == strings.ToLower(id) {
			if tx.Validate() {
				fmt.Printf("valid\n")
			} else {
				fmt.Printf("invalid\n")
			}
			os.Exit(0)
		}
	}
	fmt.Printf("could not find transaction with the specified id\n")
	os.Exit(1)
}

func runTx(args []string) {
	if len(args) == 1 {
		// Print help
		os.Exit(0)
	}
	switch args[1] {
	case "create":
		if len(args) < 5 {
			// Print help
			os.Exit(0)
		}
		create(args[2], args[3], args[4])
	case "list":
		list()
	case "validate":
		if len(args) < 3 {
			os.Exit(0)
		}
		validate(args[2])
	default:
		// Print help
		os.Exit(0)
	}
}
