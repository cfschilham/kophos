package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
)

func SHA256Sum(v interface{}) [32]byte {
	buff := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buff).Encode(v); err != nil {
		log.Fatalf("error while encoding block: %v\n", err)
	}
	return sha256.Sum256(buff.Next(buff.Len()))
}
