package miner

import (
	"fmt"
	"github.com/cfschilham/kophos/blockchain"
	"github.com/cfschilham/kophos/command"
	"github.com/cfschilham/kophos/models"
	"github.com/cfschilham/kophos/quadratic-sieve"
	"github.com/cfschilham/kophos/store"
	"github.com/cfschilham/kophos/tx"
	"github.com/cfschilham/kophos/wallet"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"math"
	"math/big"
	"math/rand"
	"time"
)

var CmdMine = command.Command{
	Run: runMine,
}

func runMine(args []string) {
	if len(args) < 2 {
		fmt.Printf("Usage:\n" +
			"kophos miner <walletId> (use \"kophos wallet list\" to see wallets")
		return
	}

	wallets := store.Get().Wallets
	wi, err := wallet.Lookup(wallets, args[1])
	if err != nil || wi == -1 {
		logrus.Fatalf("could not find wallet with id: %s", args[1])
	}

	minerWallet := wallets[wi]

	var diff uint64
	// 100 quadrillion
	flag.Uint64VarP(&diff, "difficulty", "d", 100000000000000000, "The blockchain mining difficulty")
	flag.Parse()

	chain := store.Get().Blocks
	if chain == nil {
		first := blockchain.Block{}
		chain = []blockchain.Block{first}
	}

	rand.Seed(int64(time.Now().Nanosecond()))

	logrus.Infof("starting miner difficulty: %v,  walletId: %s", diff, args[1])

	startTime := time.Now()
	numFactors := 0
	blockCount := 0
	go func() {
		for {
			t := time.Tick(time.Second * 10)
			<-t
			logrus.Infof(
				"avg. blocks/min: %.2f, %.2f MF/s",
				float64(blockCount)/time.Now().Sub(startTime).Minutes(),
				(float64(numFactors) / time.Now().Sub(startTime).Seconds()),
			)
		}
	}()
	childBlock := chain[len(chain)-1]
	cbHash := childBlock.MustHash()
	start := int64(diff)
	if len(chain) != 1 {
		start = childBlock.Nonce + 1
	}
	for nonce := start; true; nonce++ {
		n := big.NewInt(nonce)

		f1, f2 := quadratic_sieve.Factorize(n)
		if f1 == nil || f2 == nil {
			continue
		}
		//fmt.Printf("%d x %d = %d", f1, f2, f1.Int64()*f2.Int64())
		b := blockchain.Block{
			Seq:       childBlock.Seq + 1,
			Time:      uint64(time.Now().Unix()),
			Nonce:     nonce,
			Factors:   []int64{f1.Int64(), f2.Int64()},
			ChildHash: cbHash,
			Miner:     *minerWallet.Key.PublicKey.N,
			Txs:       []models.Tx{},
		}
		if !b.IsValid(int64(diff), &childBlock) {
			if nonce == math.MaxInt {
				nonce = 0
			}
			continue
		}
		logrus.Infof("found block with seq %v and nonce %d with factors %v at time %v", b.Seq, b.Nonce, b.Factors, time.Unix(int64(b.Time), 0))
		blockCount++
		var txs []*models.Tx
		for _, t := range store.Get().Txs {
			wallets := store.Get().Wallets
			wi, err := wallet.Lookup(wallets, t.Sender)
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
	}
}
