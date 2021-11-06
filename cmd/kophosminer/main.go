package main

import (
	"github.com/cfschilham/kophos/blockchain"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"math"
	"math/rand"
	"time"
)



func main() {
	//data := make([]byte, 59)
	//_, _ = rand.Read(data)
	////dhtmp.Sum([]byte("hello world"))
	//dhtmp.Sum(data)
	//os.Exit(0)

	var diff uint
	flag.UintVarP(&diff, "difficulty", "d", 100000, "The blockchain mining difficulty")
	flag.Parse()

	var first = blockchain.Block{}
	var chain = []blockchain.Block{first}

	rand.Seed(int64(time.Now().Nanosecond()))

	logrus.Infof("starting miner, difficulty: %v", diff)

	startTime := time.Now()
	go func() {
		for {
			t := time.Tick(time.Second * 10)
			<-t
			logrus.Infof("avg. blocks/min: %.2f", float64(len(chain) - 1) / time.Now().Sub(startTime).Minutes())
		}
	}()
	for nonce := uint(0); true; nonce++ {
		childBlock := chain[len(chain)-1]
		b := blockchain.Block{
			Seq: childBlock.Seq+1,
			Time:      int(time.Now().Unix()),
			Nonce:     nonce,
			ChildHash: childBlock.Hash(),
		}
		if !b.IsValid(diff) {
			if nonce == math.MaxUint {
				nonce = uint(0)
			}
			continue
		}
		logrus.Infof("found block with seq %v and nonce %v at time %v", b.Seq, b.Nonce, b.Time)
		chain = append(chain, b)
		nonce = uint(0)
	}
}
