package miner

import (
	"github.com/cfschilham/kophos/blockchain"
	"github.com/cfschilham/kophos/command"
	"github.com/cfschilham/kophos/store"
	"github.com/cfschilham/kophos/tx"
	"github.com/cfschilham/kophos/wallet"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"math"
	"math/rand"
	"time"
)

var CmdMine = command.Command{
	Run: runMine,
}

func runMine(args []string) {
	var diff uint64
	flag.Uint64VarP(&diff, "difficulty", "d", 100000, "The blockchain mining difficulty")
	flag.Parse()

	chain := store.Get().Blocks
	if chain == nil {
		first := blockchain.Block{}
		chain = []blockchain.Block{first}
	}

	rand.Seed(int64(time.Now().Nanosecond()))

	logrus.Infof("starting miner, difficulty: %v", diff)

	startTime := time.Now()
	numHashes := 0
	go func() {
		for {
			t := time.Tick(time.Second * 10)
			<-t
			logrus.Infof(
				"avg. blocks/min: %.2f, %.2f MH/s",
				float64(len(chain)-1)/time.Now().Sub(startTime).Minutes(),
				(float64(numHashes)/time.Now().Sub(startTime).Seconds())/1000000,
			)
		}
	}()
	childBlock := chain[len(chain)-1]
	cbHash := childBlock.MustHash()
	for nonce := uint64(0); true; nonce++ {
		b := blockchain.Block{
			Seq:       childBlock.Seq + 1,
			Time:      uint64(time.Now().Unix()),
			Nonce:     nonce,
			ChildHash: cbHash,
			Txs: []tx.Tx{},
		}
		numHashes++
		if !b.IsValid(diff) {
			if nonce == math.MaxInt {
				nonce = 0
			}
			continue
		}
		logrus.Infof("found block with seq %v and nonce %v at time %v", b.Seq, b.Nonce, b.Time)
		var txs []*tx.Tx
		for _, t := range store.Get().Txs {
			wallets := store.Get().Wallets
			wi, err := wallet.Lookup(wallets, t.Sender)
			if err != nil || wi == -1 {
				logrus.Errorf("could not find sender wallet for transaction: %064X, transaction deleted",
					t.Hash())
				continue
			}

			switch t.Status(wallets[wi]) {
			case tx.InsufficientBalance:
				logrus.Infof("transaction removed because of insufficient balance: %064X", t.Hash())
				continue
			case tx.NotSigned:
				txs = append(txs, t)
				continue
			default: // when transaction is valid
				b.Txs = append(b.Txs, *t)
				logrus.Infof("processed transaction: %064X", t.Hash())
			}
		}
		err := store.Mutate(func(s *store.Store) {
			s.Txs = txs
		})
		if err != nil {
			logrus.Errorf("an error occured while writing pending transactions to store")
		}
		chain = append(chain, b)
		err = store.Mutate(func(s *store.Store) {
			s.Blocks = chain
		})
		if err != nil {
			logrus.Errorf("an error occured while writing blockchain to store")
		}

		childBlock, cbHash = b, b.MustHash()
		nonce = 0
	}
}
