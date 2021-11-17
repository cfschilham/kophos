package main

import (
	"encoding/csv"
	"github.com/cfschilham/kophos/dhtmp"
	"github.com/sirupsen/logrus"
	"math/big"
	"math/rand"
	"os"
	"os/signal"
	"time"
)

func main() {
	//rand.Seed(time.Now().UnixNano())
	//dist := map[string]int{}
	//c := make(chan os.Signal, 1)
	//signal.Notify(c, os.Interrupt, os.Kill)
	//go func() {
	//	t := time.Tick(time.Second*10)
	//	for {
	//		<-t
	//		logrus.Infof("sample size: %v hashes", len(dist))
	//	}
	//}()
	//for data := big.NewInt(0); true; data.Add(data, big.NewInt(1)){
	//	sum := dhtmp.Sum(data.Bytes())
	//	dist[big.NewInt(0).SetBytes(sum[:]).String()]++
	//	select {
	//	case <-c:
	//		f, err := os.OpenFile("/home/cfschilham/dullhash.csv", os.O_CREATE|os.O_RDWR, 0755)
	//		if err != nil {
	//			logrus.Fatalf("failed to create file: %v", err)
	//		}
	//		defer f.Close()
	//		csvData := [][]string{
	//			{"Value", "Amount"},
	//		}
	//		for k, v := range dist {
	//			//kbi, _ := big.NewInt(0).SetString(k, 10)
	//			//kbi.Div(kbi, big.NewInt(math.MaxInt64))
	//			csvData = append(csvData, []string{k, strconv.Itoa(v)})
	//		}
	//		csv.NewWriter(f).WriteAll(csvData)
	//		os.Exit(0)
	//	default:
	//	}
	//}

	rand.Seed(time.Now().UnixNano())
	dist := map[string]string{}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		t := time.Tick(time.Second*10)
		for {
			<-t
			logrus.Infof("sample size: %v hashes", len(dist))
		}
	}()
	for data := big.NewInt(0); true; data.Add(data, big.NewInt(1)){
		sum := dhtmp.Sum(data.Bytes())
		dist[data.String()] = big.NewInt(0).SetBytes(sum[:]).String()
		select {
		case <-c:
			f, err := os.OpenFile("/home/cfschilham/dullhash.csv", os.O_CREATE|os.O_RDWR, 0755)
			if err != nil {
				logrus.Fatalf("failed to create file: %v", err)
			}
			defer f.Close()
			csvData := [][]string{
				{"Input", "Output"},
			}
			for k, v := range dist {
				//kbi, _ := big.NewInt(0).SetString(k, 10)
				//kbi.Div(kbi, big.NewInt(math.MaxInt64))
				csvData = append(csvData, []string{k, v})
			}
			csv.NewWriter(f).WriteAll(csvData)
			os.Exit(0)
		default:
		}
	}

	//var diff uint
	//flag.UintVarP(&diff, "difficulty", "d", 100000, "The blockchain mining difficulty")
	//flag.Parse()
	//
	//var first = blockchain.Block{}
	//var chain = []blockchain.Block{first}
	//
	//rand.Seed(int64(time.Now().Nanosecond()))
	//
	//logrus.Infof("starting miner, difficulty: %v", diff)
	//
	//startTime := time.Now()
	//go func() {
	//	for {
	//		t := time.Tick(time.Second * 10)
	//		<-t
	//		logrus.Infof("avg. blocks/min: %.2f", float64(len(chain) - 1) / time.Now().Sub(startTime).Minutes())
	//	}
	//}()
	//for nonce := uint(0); true; nonce++ {
	//	childBlock := chain[len(chain)-1]
	//	b := blockchain.Block{
	//		Seq: childBlock.Seq+1,
	//		Time:      int(time.Now().Unix()),
	//		Nonce:     nonce,
	//		ChildHash: childBlock.Hash(),
	//	}
	//	if !b.IsValid(diff) {
	//		if nonce == math.MaxUint {
	//			nonce = uint(0)
	//		}
	//		continue
	//	}
	//	logrus.Infof("found block with seq %v and nonce %v at time %v", b.Seq, b.Nonce, b.Time)
	//	chain = append(chain, b)
	//	nonce = uint(0)
	//}
}
