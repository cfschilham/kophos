package main

import (
	"fmt"
	"github.com/cfschilham/kophos/miner"
	"github.com/cfschilham/kophos/store"
	"log"
	"os"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Print("Usage:\n" +
			"	kophos miner - Start the miner\n" +
			"	kophos wallet - See all wallet options\n" +
			"	kophos tx - See all transaction options\n" +
			"	kophos store erase - To erase all data")
		os.Exit(0)
	}
	// Initialize store
	if err := store.Init(); err != nil {
		log.Fatalf("error while loading data: %v", err)
	}
	switch os.Args[1] {
	case "miner":
		miner.CmdMine.Run(os.Args[1:])
	case "wallet":
		store.CmdWallet.Run(os.Args[1:])
	case "tx":
		store.CmdTx.Run(os.Args[1:])
	case "store":
		store.CmdStore.Run(os.Args[1:])
	default:
		fmt.Print("Usage:\n" +
			"	kophos miner - Start the miner\n" +
			"	kophos wallet - See all wallet options\n" +
			"	kophos tx - See all transaction options\n" +
			"	kophos store erase - To erase all data")
		os.Exit(0)
	}
}
