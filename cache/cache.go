package cache

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
)

var Dir string

func init() {
	var err error

	// Check if data directory exists, create one if not.
	Dir, err = os.UserCacheDir()
	if err != nil {
		fmt.Printf("error while getting user cache directory: %v\n", err)
		os.Exit(1)
	}
	Dir = filepath.ToSlash(filepath.Join(Dir, "kophos"))
	if _, err = os.Stat(Dir); os.IsNotExist(err) {
		fmt.Printf("creating kophos data directory\n")
		if err = os.Mkdir(Dir, 0755); err != nil {
			fmt.Printf("error while creating kophos data directory: %v\n", err)
			os.Exit(1)
		}
	} else if err != nil {
		fmt.Printf("error while retrieving kophos data: %v\n", err)
		os.Exit(1)
	}
}

func Save(a interface{}, name string) error {
	if err := os.Remove(filepath.Join(Dir, name)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error while removing %v file: %v", name, err)
	}
	f, err := os.OpenFile(filepath.Join(Dir, name), os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return fmt.Errorf("error while writing %v file: %v", name, err)
	}
	defer f.Close()
	if err = gob.NewEncoder(f).Encode(a); err != nil {
		return fmt.Errorf("error while encoding %v file: %v", name, err)
	}
	return nil
}

func Load(a interface{}, name string) error {
	f, err := os.Open(filepath.Join(Dir, name))
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("error while opening %v file: %v", name, err)
	}
	defer f.Close()
	if err = gob.NewDecoder(f).Decode(a); err != nil {
		return fmt.Errorf("error while decoding %v file: %v", name, err)
	}
	return nil
}
