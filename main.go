package main

import (
	"github.com/cfschilham/kophos/miner"
	"github.com/cfschilham/kophos/store"
	"github.com/cfschilham/kophos/tx"
	"github.com/cfschilham/kophos/wallet"
	"log"
	"os"
)

func main() {
	if len(os.Args) == 1 {
		// Print help
		os.Exit(0)
	}
	if err := store.Init(); err != nil {
		log.Fatalf("error while loading data: %v", err)
	}
	switch os.Args[1] {
	case "miner":
		miner.CmdMine.Run(os.Args[1:])
	case "wallet":
		wallet.CmdWallet.Run(os.Args[1:])
	case "tx":
		tx.CmdTx.Run(os.Args[1:])
	default:
		// Print help
		os.Exit(0)
	}
}
