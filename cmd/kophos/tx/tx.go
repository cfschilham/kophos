package tx

import (
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"github.com/cfschilham/kophos/blockchain/tx"
	"github.com/cfschilham/kophos/cmd/kophos/base"
	"github.com/cfschilham/kophos/store"
	"log"
	"os"
	"strconv"
	"strings"
)

var CmdTx = base.Command{
	Run: run,
}

func decodeAddr(addr string) ([64]byte, error) {
	addrBytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(addr)
	if err != nil {
		return [64]byte{}, err
	}
	addrArray := [64]byte{}
	copy(addrBytes[:], addrBytes)
	return addrArray, nil
}

func create(senderStr, recipStr, amountStr string) error {
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		fmt.Printf("invalid amount\n")
		os.Exit(1)
	}
	sender, err := decodeAddr(senderStr)
	if err != nil {
		return err
	}
	recip, err := decodeAddr(recipStr)
	if err != nil {
		return err
	}
	s := store.Get()
	prev := s.Blockchain.LastTxFrom(sender)
	t := &tx.Tx{
		Header: tx.Header{
			Sender:    sender,
			Recipient: recip,
			Amount:    uint64(amount),
			Seq:       prev.Seq + 1,
			PrevHash:  prev.Hash(),
		},
	}
	fmt.Printf("transaction created with hash %064x\n", t.Hash())
	return store.Mutate(func(s *store.Store) {
		s.Txs = append(s.Txs, t)
	})
}

func list(args []string) {
	s := store.Get()
	unconfirmed := s.Txs
	confirmed := []*tx.Tx{}
	for _, b := range s.Blockchain {
		confirmed = append(confirmed, b.Txs...)
	}
	txs := append(unconfirmed, confirmed...)
	if len(args) != 0 {
		if args[0] == "confirmed" {
			txs = confirmed
		} else if args[0] == "unconfirmed" {
			txs = unconfirmed
		} else {
			usage()
		}
	}
	for i, t := range txs {
		fmt.Printf("%03d: %064x\n", i, t.Hash())
	}
}

func lookup(txs []*tx.Tx, id string) int {
	for i, t := range txs {
		hash := t.Hash()
		if hex.EncodeToString(hash[:]) == strings.ToLower(id) {
			return i
		}
	}
	return -1
}

func status(id string) {
	s := store.Get()
	i := lookup(s.Txs, id)
	if i != -1 {
		if s.Txs[i].HasValidSig() {
			fmt.Printf("valid, unconfirmed\n")
		} else {
			fmt.Printf("invalid, unconfirmed\n")
		}
		os.Exit(0)
	}
	for _, b := range s.Blockchain {
		i = lookup(b.Txs, id)
		if i != -1 {
			if s.Txs[i].HasValidSig() {
				fmt.Printf("valid, confirmed\n")
			} else {
				fmt.Printf("invalid, confirmed (this should not be possible?? this is bad)\n")
			}
			os.Exit(0)
		}
	}
	log.Fatalf("could not find transaction with the specified id\n")
}

func usage() {
	fmt.Printf("Usage:\n" +
		"	kophos tx create <sender address> <recipient address> <amount> Create a new transaction\n" +
		"	kophos tx list [unconfirmed|confirmed]                         List transactions, optionally filter by unconfirmed or confirmed\n" +
		"	kophos status <transaction id>                                 Check status of a transaction\n",
	)
	os.Exit(0)
}

func run(args []string) {
	if len(args) == 1 {
		usage()
	}
	switch args[1] {
	case "create":
		if len(args) < 5 {
			usage()
		}
		err := create(args[2], args[3], args[4])
		if err != nil {
			log.Fatalf("an error occured while creating transaction: %v", err)
		}
	case "list":
		if len(args) < 2 {
			usage()
		}
		list(args[2:])
	case "status":
		if len(args) < 3 {
			usage()
		}
		status(args[2])
	default:
		usage()
	}
}
