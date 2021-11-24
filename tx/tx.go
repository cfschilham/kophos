package tx

import (
	"crypto/rsa"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"github.com/cfschilham/kophos/command"
	"github.com/cfschilham/kophos/models"
	"github.com/cfschilham/kophos/store"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
)

var CmdTx = command.Command{
	Run: runTx,
}

func create(sender, recip, amountStr string) error {
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		fmt.Printf("invalid amount\n")
		os.Exit(1)
	}

	txs := getTransactionsFrom(sender)
	txSeq := uint64(1)
	txChildHash := [32]byte{0}
	if len(txs) > 0 {
		txSeq = txs[len(txs) - 1].Seq + 2
		txChildHash = txs[len(txs) - 1].Hash()
	}

	t := &models.Tx{
		Header: models.Header{
			Sender:    sender,
			Recipient: recip,
			Amount:    uint64(amount),
			Seq: txSeq - 1,
			ChildHash: txChildHash,
		},
	}
	fmt.Printf("transaction created with hash %x\n", t.Hash())
	return store.Mutate(func(s *store.Store) {
		s.Txs = append(s.Txs, t)
	})
}

func getTransactionsFrom(walletId string) []*models.Tx {
	allTxs := GetProcessed()
	var txs []*models.Tx
	for _, t := range allTxs {
		if t.Sender == walletId {
			txs = append(txs, &t)
		}
	}
	return txs
}

func getTransactionsTo(walletId string) []*models.Tx {
	allTxs := GetProcessed()
	var txs []*models.Tx
	for _, t := range allTxs {
		if t.Recipient == walletId {
			txs = append(txs, &t)
		}
	}
	return txs
}

func listQueued() {
	txs := store.Get().Txs
	if len(txs) == 0 {
		fmt.Print("no queued transactions")
		os.Exit(0)
	}
	for _, t := range txs {
		fmt.Printf("%03d: %064X\n", t.Seq, t.Hash())
	}
	os.Exit(0)
}

func GetProcessed() []models.Tx {
	blocks := store.Get().Blocks
	var txs []models.Tx
	for _, block := range blocks {
		if len(block.Txs) > 0 {
			txs = append(txs, block.Txs...)
		}
	}
	return txs
}

func listProcessed() {
	txs := GetProcessed()
	if len(txs) == 0 {
		fmt.Print("no processed transactions")
		os.Exit(0)
	}
	for _, t := range GetProcessed() {
		fmt.Printf("%03d: %064X\n", t.Seq, t.Hash())
	}
	os.Exit(0)
}

func listForWallet(id string) {
	txs := GetProcessed()
	if len(txs) == 0 {
		fmt.Print("no processed transactions")
		os.Exit(0)
	}
	for _, t := range GetProcessed() {
		if t.Sender == id {
			fmt.Printf("%03d: %064X\n", t.Seq, t.Hash())
		}
	}
	os.Exit(0)
}

func checkStatus(id string) {
	txs := store.Get().Txs
	for _, t := range txs {
		hash := t.Hash()
		if hex.EncodeToString(hash[:]) == strings.ToLower(id) {
			wallets := store.Get().Wallets
			n, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(id)
			if err != nil {
				log.Fatalf("error while decoding wallet id: %v", err)
			}
			var wallet *models.Wallet
			for _, w := range wallets {
				if big.NewInt(0).SetBytes(n).Cmp(w.Key.PublicKey.N) == 0 {
					wallet = &w
				}
			}
			if wallet == nil {
				log.Fatalf("error while finding wallet with id: %s", id)
			}
			fmt.Printf("transaction status: %v\n", Status(*t, *wallet))
			os.Exit(0)
		}
	}
	log.Fatalf("could not find transaction with the specified id\n")
}

func Status(tx models.Tx, sender models.Wallet) models.Status {
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
		return models.Invalid
	}

	return models.Valid
}


func runTx(args []string) {
	if len(args) == 1 {
		fmt.Print("Usage:\n" +
			"	kophos tx create <sourceWalletId> <destWalletId> <amount> - Create a new transaction\n" +
			"	kophos tx list <queue|processed> - List all queued or processed transaction\n" +
			"	kophos status <transactionId> - Check status of a transaction")
		os.Exit(0)
	}
	switch args[1] {
	case "create":
		if len(args) < 5 {
			fmt.Print("Usage:\n" +
				"	kophos tx create <sourceWalletId> <destWalletId> <amount> (use \"kophos wallet list\" to view all wallets)\n")
			os.Exit(0)
		}
		err := create(args[2], args[3], args[4])
		if err != nil {
			log.Fatalf("an error occured while creating transaction: %v", err)
		}
	case "list":
		if len(args) < 3 {
			fmt.Print("Usage:\n" +
				"	kophos tx list <queue|processed|wallet> <?walletId>")
			os.Exit(0)
		}
		switch args[2] {
		case "queue":
			listQueued()
		case "processed":
			listProcessed()
		case "wallet":
			if len(args) < 4 {
				fmt.Print("Usage:\n" +
					"	kophos tx list wallet <walletId> (use \"kophos wallet list\" to get all wallets)")
				os.Exit(0)
			}
		default:
			fmt.Print("Usage:\n" +
				"	kophos tx list <queue|processed> ")
			os.Exit(0)
		}
	case "status":
		if len(args) < 3 {
			fmt.Print("Usage:\n" +
				"	kophos status <transactionId> (use \"kophos tx list\" to view all transactions")
			os.Exit(0)
		}
		checkStatus(args[2])
	default:
		fmt.Print("Usage:\n" +
			"	kophos tx create <sourceWalletId> <destWalletId> <amount> - Create a new transaction\n" +
			"	kophos tx list <queue|processed> - List all queued or processed transaction\n" +
			"	kophos status <transactionId> - Check status of a transaction")
		os.Exit(0)
	}
}