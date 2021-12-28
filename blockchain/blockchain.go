package blockchain

import (
	"bytes"
	"encoding/gob"
	"github.com/cfschilham/dullhash"
	"github.com/cfschilham/kophos/models"
	"math/big"
)

const BlockReward uint64 = 10

var maxHash = big.NewInt(0).SetBytes(dullhash.MaxSum[:])

type Block struct {
	Seq       uint
	Time      uint64
	Miner     big.Int
	ChildHash [32]byte
	Nonce     int64
	Factors   []int64
	Txs       []models.Tx
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

func (b Block) IsValid(di int64, childBlock *Block) bool {
	num := b.Factors[0] * b.Factors[1]
	for _, factor := range b.Factors[2:] {
		num *= factor
	}
	return num == b.Nonce && num > di && num > childBlock.Nonce
}
