package blockchain

import (
	"bytes"
	"encoding/gob"
	"github.com/cfschilham/dullhash"
	"github.com/cfschilham/kophos/tx"
	"math/big"
)

var maxHash = big.NewInt(0).SetBytes(dullhash.MaxSum[:])

type Block struct {
	Seq       uint
	Time      uint64
	Miner     uint
	ChildHash [32]byte
	Nonce     uint64
	Txs       []tx.Tx
}

func (b Block) Bytes() ([]byte, error) {
	buf := &bytes.Buffer{}
	err := gob.NewEncoder(buf).Encode(&b)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (b Block) MustHash() [32]byte {
	hash, err := b.Hash()
	if err != nil {
		panic("error while hashing block")
	}
	return hash
}

func (b Block) Hash() ([32]byte, error) {
	buf, err := b.Bytes()
	if err != nil {
		return [32]byte{}, err
	}
	return dullhash.Sum(buf), nil
}

func (b Block) HashBigInt() *big.Int {
	hash := b.MustHash()
	return big.NewInt(0).SetBytes(hash[:])
}

func (b Block) IsValid(di uint64) bool {
	h, d := b.HashBigInt(), big.NewInt(0).SetUint64(di)
	return h.Cmp(d.Div(maxHash, d)) == -1
}
