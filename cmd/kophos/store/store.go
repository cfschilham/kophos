package store

import (
	"fmt"
	"github.com/cfschilham/kophos/cmd/kophos/base"
	"github.com/cfschilham/kophos/store"
	"log"
	"os"
)

var CmdStore = base.Command{
	Run: runStore,
}

func runStore(args []string) {
	if len(args) == 1 {
		fmt.Print("Usage:\n" +
			"	kophos store erase - Erase all kophos data")
		os.Exit(0)
	}
	switch args[1] {
	case "clear":
		if err := store.Init(); err != nil {
			fmt.Printf("error while initializing store: %v\n", err)
		}

		if err := store.Set(store.Store{}); err != nil {
			log.Fatalf("error while try to clear store: %v", err)
		}
		fmt.Printf("store succesfully cleared\n")
	default:
		fmt.Print("Usage:\n" +
			"	kophos store erase - Erase all kophos data")
	}
}
