package main

import (
	"fmt"
	"github.com/cfschilham/kophos/cmd/kophos/miner"
	internalstore "github.com/cfschilham/kophos/cmd/kophos/store"
	"github.com/cfschilham/kophos/cmd/kophos/tx"
	"github.com/cfschilham/kophos/cmd/kophos/wallet"
	"github.com/cfschilham/kophos/store"
	"log"
	"os"
)

func usage() {
	fmt.Print("Usage:\n" +
		"	kophos miner       Start the miner\n" +
		"	kophos wallet      Create and manipulate wallets\n" +
		"	kophos tx          Create and manipulate transactions\n" +
		"	kophos store clear Clear all data")
	os.Exit(0)
}

func main() {
	if len(os.Args) == 1 {
		fmt.Print("Usage:\n" +
			"	kophos miner - Start the miner\n" +
			"	kophos wallet - See all wallet options\n" +
			"	kophos tx - See all transaction options\n" +
			"	kophos store erase - To erase all data")
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
	case "store":
		internalstore.CmdStore.Run(os.Args[1:])
	default:
		usage()
	}
}
