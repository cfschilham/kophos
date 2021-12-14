package miner

import (
	"fmt"
	"github.com/cfschilham/kophos/blockchain"
	tx2 "github.com/cfschilham/kophos/blockchain/tx"
	"github.com/cfschilham/kophos/cmd/kophos/base"
	"github.com/cfschilham/kophos/cmd/kophos/store"
	"github.com/cfschilham/kophos/cmd/kophos/tx"
	"github.com/cfschilham/kophos/cmd/kophos/wallet"
	"github.com/cfschilham/kophos/models"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"math"
	"math/rand"
	"time"
)

var CmdMine = base.Command{
	Run: runMine,
}

func runMine(args []string) {
	if len(args) < 2 {
		fmt.Printf("Usage:\n" +
			"kophos miner <walletId> (use \"kophos wallet list\" to see wallets")
	}

	wallets := store.Get().Wallets
	wi, err := wallet.lookup(wallets, args[1])
	if err != nil || wi == -1 {
		logrus.Fatalf("could not find wallet with id: %s", args[1])
	}

	minerWallet := wallets[wi]

	var diff uint64
	flag.Uint64VarP(&diff, "difficulty", "d", 100000, "The blockchain mining difficulty")
	flag.Parse()

	chain := store.Get().Blocks
	if chain == nil {
		first := blockchain.Block{}
		chain = []blockchain.Block{first}
	}

	rand.Seed(int64(time.Now().Nanosecond()))

	logrus.Infof("starting miner, difficulty: %v, walletId: %s", diff, args[1])

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
			Miner:     *minerWallet.Key.PublicKey.N,
			Txs:       []tx2.Tx{},
		}
		numHashes++
		if !b.IsValid(diff) {
			if nonce == math.MaxInt {
				nonce = 0
			}
			continue
		}
		logrus.Infof("found block with seq %v and nonce %v at time %v", b.Seq, b.Nonce, b.Time)
		var txs []*tx2.Tx
		for _, t := range store.Get().Txs {
			wallets := store.Get().Wallets
			wi, err := wallet.lookup(wallets, t.Sender)
			if err != nil || wi == -1 {
				logrus.Errorf("could not find sender wallet for transaction: %064X, transaction deleted",
					t.Hash())
				continue
			}

			switch tx.Status(*t, wallets[wi]) {
			case models.Invalid:
				txs = append(txs, t)
				continue
			default: // when transaction is valid
				logrus.Infof("processing transaction: %064X", t.Hash())
			}
			balance := wallet.GetWalletBalance(wallets[wi])
			if balance < t.Amount {
				logrus.Errorf("insufficient balance to process transaction, deleting..")
				continue
			}
			b.Txs = append(b.Txs, *t)
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
