package store

import (
	"encoding/gob"
	"fmt"
	"github.com/cfschilham/kophos/blockchain"
	"github.com/cfschilham/kophos/blockchain/tx"
	"github.com/cfschilham/kophos/blockchain/wallet"
	"os"
	"path/filepath"
	"sync"
)

type Store struct {
	Txs        []*tx.Tx
	Wallets    []wallet.Wallet
	Blockchain blockchain.Blockchain
}

var (
	store   Store
	mut     *sync.Mutex
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

// Set overwrites the store to the provided store. Should only be used if it
// doesn't depend on data from the store itself. In that case, use Mutate
// instead.
func Set(s Store) error {
	mut.Lock()
	defer mut.Unlock()
	store = s
	return save(store)
}

// Get returns the current store.
func Get() Store {
	mut.Lock()
	defer mut.Unlock()
	return store
}

// Mutate runs function m on the store to change data in the store based on data
// already in the store. Concurrency safe.
func Mutate(m func(s *Store)) error {
	mut.Lock()
	defer mut.Unlock()
	m(&store)
	return save(store)
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
	s := Store{}
	if err = gob.NewDecoder(f).Decode(&s); err != nil {
		return Store{}, fmt.Errorf("error while decoding data file: %v", err)
	}
	return s, nil
}
