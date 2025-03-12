package main

import (
	"context"
	"fmt"
	"os"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
	"github.com/marc-watters/the-block-chain-bar/v2/node"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	s, err := db.NewStateFromDisk(cwd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer s.Close()

	pendingBlock := node.NewPendingBlock(
		db.Hash{},
		s.NextBlockHeight(),
		db.NewAccount("andrej"),
		[]db.Trx{
			db.NewTrx("andrej", "andrej", 3, ""),
			db.NewTrx("andrej", "andrej", 700, "reward"),
			db.NewTrx("andrej", "babayaga", 2000, ""),
			db.NewTrx("andrej", "andrej", 100, "reward"),
			db.NewTrx("babayaga", "andrej", 1, ""),
			db.NewTrx("babayaga", "caesar", 1000, ""),
			db.NewTrx("babayaga", "andrej", 50, ""),
			db.NewTrx("andrej", "andrej", 600, "reward"),
		},
	)

	if _, err := node.Mine(context.Background(), pendingBlock); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
