package store

import (
	"bufio"
	"encoding/base32"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/cfschilham/kophos/blockchain"
	"github.com/cfschilham/kophos/command"
	"github.com/cfschilham/kophos/tx"
	"github.com/cfschilham/kophos/wallet"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var CmdStore = command.Command{
	Run: runStore,
}

var CmdTx = command.Command{
	Run: runTx,
}

var CmdWallet = command.Command{
	Run: runWallet,
}

type Store struct {
	Txs []*tx.Tx
	TxSeq uint64
	LastTxHash [32]byte
	Wallets []wallet.Wallet
	Blocks []blockchain.Block
}

var (
	store Store
	mut *sync.Mutex
	dataDir string
)

func Init() error {
	mut = &sync.Mutex{}

	var err error
	dataDir, err = constructDataDirPath()
	if err != nil {
		return err
	}
	store, err = load()
	return err
}

func Set(s Store) error {
	mut.Lock()
	defer mut.Unlock()
	store = s
	return save(store)
}

func Get() Store {
	mut.Lock()
	defer mut.Unlock()
	return store
}

func Mutate(m func(s *Store)) error {
	mut.Lock()
	defer mut.Unlock()
	m(&store)
	return save(store)
}

func createTx(sender, recip, amountStr string) error {
	t := tx.Create(sender, recip, amountStr, Get().TxSeq, Get().LastTxHash)
	fmt.Printf("transaction created with hash %x\n", t.Hash())
	return Mutate(func(s *Store) {
		s.Txs = append(s.Txs, t)
		s.TxSeq++
		s.LastTxHash = t.Hash()
	})
}

func listQueuedTxs() {
	txs := Get().Txs
	if len(txs) == 0 {
		fmt.Print("no queued transactions")
		os.Exit(0)
	}
	for _, t := range txs {
		fmt.Printf("%03d: %064X\n", t.Seq, t.Hash())
	}
	os.Exit(0)
}

func getProcessedTxs() []tx.Tx {
	blocks := Get().Blocks
	var txs []tx.Tx
	for _, block := range blocks {
		if len(block.Txs) > 0 {
			txs = append(txs, block.Txs...)
		}
	}
	return txs
}

func listProcessedTxs() {
	txs := getProcessedTxs()
	if len(txs) == 0 {
		fmt.Print("no processed transactions")
		os.Exit(0)
	}
	for _, t := range getProcessedTxs() {
		fmt.Printf("%03d: %064X\n", t.Seq, t.Hash())
	}
	os.Exit(0)
}

func checkTxStatus(id string) {
	txs := Get().Txs
	for _, tx := range txs {
		hash := tx.Hash()
		if hex.EncodeToString(hash[:]) == strings.ToLower(id) {
			wallets := Get().Wallets
			wi, err := wallet.Lookup(wallets, tx.Sender)
			if err != nil {
				log.Fatalf("error while finding wallet %v", err)
			}
			if wi == -1 {
				log.Fatalf("could not find sender wallet")
			}
			fmt.Printf("transaction status: %v\n", tx.Status(wallets[wi]))
			os.Exit(0)
		}
	}
	log.Fatalf("could not find transaction with the specified id\n")
}

func createWallet() {
	w, err := wallet.New()
	if err != nil {
		log.Fatalf("error while creating wallet: %v\n", err)
	}
	if err = Mutate(func(s *Store) { s.Wallets = append(s.Wallets, w) }); err != nil {
		log.Fatalf("error while saving wallets: %v\n", err)
	}
	fmt.Printf(
		"created wallet with address %v\n",
		base32.StdEncoding.WithPadding(base32.NoPadding).
			EncodeToString(w.Key.PublicKey.N.Bytes()),
	)
}

func removeWallet(id string) {
	ws := Get().Wallets

	i, err := wallet.Lookup(ws, id)
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
		if err = Mutate(func(s *Store) { s.Wallets = ws }); err != nil {
			fmt.Printf("error while saving wallets: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("removed successfully\n")
	} else {
		fmt.Printf("aborting\n")
		os.Exit(0)
	}
}

func listWallets() {
	ws := Get().Wallets
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

func sign(txID, id string) {
	if err := Mutate(func(s *Store) {
		signed := false

		for _, t := range s.Txs {
			hash := t.Hash()
			if hex.EncodeToString(hash[:]) == strings.ToLower(txID) {
				i, err := wallet.Lookup(s.Wallets, id)
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

func save(s Store) error {
	if err := os.Remove(filepath.Join(dataDir, "data")); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error while removing data file: %v", err)
	}
	f, err := os.OpenFile(filepath.Join(dataDir, "data"), os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return fmt.Errorf("error while writing data file: %v", err)
	}
	defer f.Close()
	if err = gob.NewEncoder(f).Encode(&s); err != nil {
		return fmt.Errorf("error while encoding data file: %v", err)
	}
	return nil
}

func constructDataDirPath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("error while getting user cache directory: %v\n", err)
	}
	dir = filepath.ToSlash(filepath.Join(dir, "kophos"))
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		fmt.Printf("creating kophos data directory\n")
		if err = os.Mkdir(dir, 0755); err != nil {
			return "", fmt.Errorf("error while creating kophos data at %s directory: %v\n", dir, err)
		}
	} else if err != nil {
		return "", fmt.Errorf("error while retrieving kophos data directory info: %v\n", err)
	}
	return dir, nil
}

func load() (Store, error) {
	f, err := os.Open(filepath.Join(dataDir, "data"))
	if os.IsNotExist(err) {
		return Store{}, nil
	} else if err != nil {
		return Store{}, fmt.Errorf("error while opening data file: %v", err)
	}
	defer f.Close()
	s := Store{
		TxSeq: 0,
		LastTxHash: [32]byte{0},
	}
	if err = gob.NewDecoder(f).Decode(&s); err != nil {
		return Store{}, fmt.Errorf("error while decoding data file: %v", err)
	}
	return s, nil
}


func runStore(args []string) {
	if len(args) == 1 {
		fmt.Print("Usage:\n" +
			"	kophos store erase - Erase all kophos data")
		os.Exit(0)
	}
	switch args[1] {
	case "erase":
		err := os.RemoveAll(dataDir)
		if err != nil {
			log.Fatalf("error while try to delete store: %v", err)
		}
		fmt.Printf("store succesfully erased")
	default:
		fmt.Print("Usage:\n" +
			"	kophos store erase - Erase all kophos data")
		os.Exit(0)
	}
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
		err := createTx(args[2], args[3], args[4])
		if err != nil {
			log.Fatalf("an error occured while creating transaction: %v", err)
		}
	case "list":
		if len(args) < 3 {
			fmt.Print("Usage:\n" +
				"	kophos tx list <queue|processed> ")
			os.Exit(0)
		}
		switch args[2] {
		case "queue":
			listQueuedTxs()
		case "processed":
			listQueuedTxs()
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
		checkTxStatus(args[2])
	default:
		fmt.Print("Usage:\n" +
			"	kophos tx create <sourceWalletId> <destWalletId> <amount> - Create a new transaction\n" +
			"	kophos tx list <queue|processed> - List all queued or processed transaction\n" +
			"	kophos status <transactionId> - Check status of a transaction")
		os.Exit(0)
	}
}

func runWallet(args []string) {
	if len(args) == 1 {
		fmt.Print("Usage:\n" +
			"	kophos wallet create - Create a new wallet\n" +
			"	kophos wallet list - List all wallets\n" +
			"	kophos wallet remove <walletId> - Remove a wallet\n" +
			"	kophos wallet sign <transactionId> <walletId> - Sign a transactions")
		os.Exit(0)
	}

	switch args[1] {
	case "create":
		createWallet()
	case "list":
		listWallets()
	case "remove":
		if len(args) < 3 {
			fmt.Print("Usage:\n" +
				"	kophos wallet remove <walletId> (use \"kophos wallet list\" to see wallets)")
			os.Exit(0)
		}
		removeWallet(args[2])
	case "sign":
		if len(args) < 4 {
			fmt.Print("Usage:\n" +
				"	kophos wallet sign <transactionId> <walletId> (use \"kophos tx list\" to see transactions)")
			os.Exit(0)
		}
		sign(args[2], args[3])
	default:
		fmt.Print("Usage:\n" +
			"	kophos wallet create - Create a new wallet\n" +
			"	kophos wallet list - List all wallets\n" +
			"	kophos wallet remove <walletId> - Remove a wallet\n" +
			"	kophos wallet sign <transactionId> <walletId> - Sign a transactions")
		os.Exit(0)
	}
}
