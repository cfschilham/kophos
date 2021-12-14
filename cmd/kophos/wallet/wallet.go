package wallet

import (
	"bufio"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"github.com/cfschilham/kophos/blockchain"
	"github.com/cfschilham/kophos/blockchain/wallet"
	"github.com/cfschilham/kophos/cmd/kophos/base"
	"github.com/cfschilham/kophos/cmd/kophos/tx"
	"github.com/cfschilham/kophos/store"
	"log"
	"math/big"
	"os"
	"strings"
)

var CmdWallet = base.Command{
	Run: run,
}

func lookup(ws []wallet.Wallet, id string) int {
	n, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(id)
	if err != nil {
		return -1
	}
	for i, w := range ws {
		if big.NewInt(0).SetBytes(n).Cmp(w.Key.PublicKey.N) == 0 {
			return i
		}
	}
	return -1
}

func create() {
	w, err := wallet.New()
	if err != nil {
		log.Fatalf("error while creating wallet: %v\n", err)
	}
	if err = store.Mutate(func(s *store.Store) { s.Wallets = append(s.Wallets, w) }); err != nil {
		log.Fatalf("error while saving wallets: %v\n", err)
	}
	fmt.Printf(
		"created wallet with address %v\n",
		base32.StdEncoding.WithPadding(base32.NoPadding).
			EncodeToString(w.Key.PublicKey.N.Bytes()),
	)
}

func remove(id string) {
	ws := store.Get().Wallets

	i := lookup(ws, id)
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
		if err = store.Mutate(func(s *store.Store) { s.Wallets = ws }); err != nil {
			fmt.Printf("error while saving wallets: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("removed successfully\n")
	} else {
		fmt.Printf("aborting\n")
		os.Exit(0)
	}
}

func list() {
	ws := store.Get().Wallets
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

//func GetWalletBalance(wallet wallet.Wallet) uint64 {
//	var bal uint64 = 0
//	for _, block := range store.Get().Blocks {
//		if block.Miner.Cmp(wallet.Key.PublicKey.N) == 0 {
//			bal += blockchain.BlockReward
//		}
//	}
//	walledIdStr := base32.StdEncoding.WithPadding(base32.NoPadding).
//		EncodeToString(wallet.Key.PublicKey.N.Bytes())
//	for _, t := range tx.GetProcessed() {
//		if t.Sender == walledIdStr {
//			bal -= t.Amount
//		}
//		if t.Recipient == walledIdStr {
//			bal += t.Amount
//		}
//	}
//	return bal
//}
//
//
//func checkBalance(id string) {
//	ws := store.Get().Wallets
//	i, err := lookup(ws, id)
//	if err != nil {
//		log.Fatalf("error while trying to find wallet: %v", err)
//	}
//	if i == -1 {
//		log.Fatalf("could not find wallet with the specified id")
//	}
//
//	fmt.Printf("wallet balance for wallet %s: %d", id, GetWalletBalance(ws[i]))
//}

func sign(txID, id string) {
	if err := store.Mutate(func(s *store.Store) {
		signed := false

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
				signed = true
				fmt.Printf("signed successfully\n")
			}
		}
		if !signed {
			log.Fatalf("could not find transaction with the specified id")
		}
	}); err != nil {
		log.Fatalf("%v", err)
	}
}

func run(args []string) {
	if len(args) == 1 {
		fmt.Print("Usage:\n" +
			"	kophos wallet create - Create a new wallet\n" +
			"	kophos wallet list - List all wallets\n" +
			"	kophos wallet remove <walletId> - Remove a wallet\n" +
			"	kophos wallet sign <transactionId> <walletId> - Sign a transactions\n" +
			"	kophos wallet bal <walletId> - Check wallet balance")
		os.Exit(0)
	}

	switch args[1] {
	case "create":
		create()
	case "list":
		list()
	case "remove":
		if len(args) < 3 {
			fmt.Print("Usage:\n" +
				"	kophos wallet remove <walletId> (use \"kophos wallet list\" to see wallets)")
			os.Exit(0)
		}
		remove(args[2])
	case "sign":
		if len(args) < 4 {
			fmt.Print("Usage:\n" +
				"	kophos wallet sign <transactionId> <walletId> (use \"kophos tx list\" to see transactions)")
			os.Exit(0)
		}
		sign(args[2], args[3])
	case "bal":
		if len(args) < 3 {
			fmt.Print("Usage:\n" +
				"	kophos wallet bal <walletId> (use \"kophos wallet list\" to see wallets)")
			os.Exit(0)
		}
		checkBalance(args[2])
	default:
		fmt.Print("Usage:\n" +
			"	kophos wallet create - Create a new wallet\n" +
			"	kophos wallet list - List all wallets\n" +
			"	kophos wallet remove <walletId> - Remove a wallet\n" +
			"	kophos wallet sign <transactionId> <walletId> - Sign a transactions\n" +
			"	kophos wallet bal <walletId> - Check wallet balance")
		os.Exit(0)
	}
}
