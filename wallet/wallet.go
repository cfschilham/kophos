package wallet

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"github.com/cfschilham/kophos/command"
	"github.com/cfschilham/kophos/store"
	"github.com/cfschilham/kophos/tx"
	"log"
	"math/big"
	"os"
	"strings"
)

var CmdWallet = command.Command{
	Run: runWallet,
}

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

func create() {
	ws := []Wallet{}
	if err := store.Load(&ws, "wallets"); err != nil {
		fmt.Printf("error while loading wallets: %v\n", err)
		os.Exit(1)
	}
	w, err := New()
	if err != nil {
		fmt.Printf("error while creating wallet: %v\n", err)
		os.Exit(1)
	}
	ws = append(ws, w)
	if err = store.Save(ws, "wallets"); err != nil {
		fmt.Printf("error while saving wallets: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf(
		"created wallet with address %v\n",
		base32.StdEncoding.WithPadding(base32.NoPadding).
			EncodeToString(w.Key.PublicKey.N.Bytes()),
	)
	os.Exit(0)
}

func remove(id string) {
	ws := []Wallet{}
	if err := store.Load(&ws, "wallets"); err != nil {
		fmt.Printf("error while loading wallets: %v\n", err)
		os.Exit(1)
	}
	i, err := lookup(ws, id)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	if i == -1 {
		fmt.Printf("could not find a wallet with the provided id")
		os.Exit(1)
	}
	fmt.Printf("this action cannot be undone, please type \"remove this wallet\" to confirm: ")
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		fmt.Printf("error while reading input: %v\n", err)
		os.Exit(1)
	}
	if input == "remove this wallet\n" {
		ws = append(ws[:i], ws[i+1:]...)
		if err = store.Save(ws, "wallets"); err != nil {
			fmt.Printf("error while saving wallets: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("removed successfully\n")
		os.Exit(0)
	}
	fmt.Printf("aborting\n")
	os.Exit(0)
}

func list() {
	ws := []Wallet{}
	if err := store.Load(&ws, "wallets"); err != nil {
		fmt.Printf("error while loading wallets: %v\n", err)
		os.Exit(1)
	}
	for i, w := range ws {
		fmt.Printf(
			"%03d: %v\n",
			i,
			base32.StdEncoding.WithPadding(base32.NoPadding).
				EncodeToString(w.Key.PublicKey.N.Bytes()),
		)
	}
	os.Exit(0)
}

func lookup(ws []Wallet, id string) (int, error) {
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

func sign(txID, id string) {
	if err := store.Mutate(func(s store.Store) store.Store {
		for _, t := range s.Txs {
			hash := t.Hash()
			if hex.EncodeToString(hash[:]) == strings.ToLower(txID) {
				i, err := lookup(s.Wallets, id)
				if err != nil {
					log.Fatalf("%v", err)
				}
				if i == -1 {
					log.Fatalf("could not find wallet with the specified id")
				}

				if _, err = t.Sign(s.Wallets[i].Key); err != nil {
					log.Fatalf("error while signing transaction: %v", err)
				}
				fmt.Printf("signed successfully\n")
				return s
			}
		}
		log.Fatalf("could not find transaction with the specified id")
		return s
	}); err != nil {
		log.Fatalf("%v", err)
	}
}

func runWallet(args []string) {
	if len(args) == 1 {
		// Print help
		os.Exit(0)
	}

	switch args[1] {
	case "create":
		create()
	case "list":
		list()
	case "remove":
		if len(args) < 3 {
			// Print help.
			os.Exit(0)
		}
		remove(args[2])
	case "sign":
		if len(args) < 4 {
			// Print help.
			os.Exit(0)
		}
		sign(args[2], args[3])
	default:
		// Print help.
		os.Exit(0)
	}
}
