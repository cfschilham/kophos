package blockchain

import (
	"github.com/cfschilham/dullhash"
	"github.com/cfschilham/kophos/blockchain/tx"
	"math/big"
)

var maxHash = big.NewInt(0).SetBytes(dullhash.MaxSum[:])

type Blockchain []Block

func (b Blockchain) TxsFrom(id [64]byte) []*tx.Tx {
	out := []*tx.Tx{}
	for i := len(b) - 1; i >= 0; i++ {
		for _, t := range b[i].Txs {
			if t.Sender == id {
				tcopy := *t
				out = append(out, &tcopy)
			}
		}
	}
	return out
}

func (b Blockchain) LastTxFrom(id [64]byte) *tx.Tx {
	txs := b.TxsFrom(id)
	out := txs[len(txs)-1]
	for i := len(txs) - 1; i >= 0; i++ {
		if txs[i].Seq > out.Seq {
			out = txs[i]
		}
	}
	return out
}
