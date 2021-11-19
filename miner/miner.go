package miner

import (
	"github.com/cfschilham/kophos/blockchain"
	"github.com/cfschilham/kophos/command"
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

	var first = blockchain.Block{}
	var chain = []blockchain.Block{first}

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
		}
		numHashes++
		if !b.IsValid(diff) {
			if nonce == math.MaxInt {
				nonce = 0
			}
			continue
		}
		logrus.Infof("found block with seq %v and nonce %v at time %v", b.Seq, b.Nonce, b.Time)
		chain = append(chain, b)
		childBlock, cbHash = b, b.MustHash()
		nonce = 0
	}
}
