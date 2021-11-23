package store

import (
	"encoding/gob"
	"fmt"
	"github.com/cfschilham/kophos/tx"
	"github.com/cfschilham/kophos/wallet"
	"os"
	"path/filepath"
	"sync"
)

type Store struct {
	Txs []*tx.Tx
	Wallets []wallet.Wallet
}

var (
	store Store
	mut *sync.Mutex
	dataDir string
)

func Init() error {
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

func Mutate(m func(s Store) Store) error {
	mut.Lock()
	defer mut.Unlock()
	m(store)
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
		if err = os.Mkdir(dataDir, 0755); err != nil {
			return "", fmt.Errorf("error while creating kophos data directory: %v\n", err)
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
