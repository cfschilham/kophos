package blockchain

import (
	"github.com/cfschilham/kophos/tx"
	"github.com/cfschilham/kophos/util"
	"math/big"
)

var maxHash, _ = big.NewInt(0).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", 16)

type Block struct {
	Seq int
	Time  int
	Miner     uint
	ChildHash [32]byte
	Nonce     uint
	Txs   []tx.Tx
}

func (b Block) Hash() [32]byte {
	return util.SHA256Sum(b)
}

func (b Block) HashBigInt() *big.Int {
	hash := b.Hash()
	return big.NewInt(0).SetBytes(hash[:])
}

func (b Block) IsValid(di uint) bool {
	h, d := b.HashBigInt(), big.NewInt(0).SetUint64(uint64(di))
	return h.Cmp(d.Div(maxHash, d)) == -1
}
