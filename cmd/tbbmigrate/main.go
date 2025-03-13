package main

import (
	"context"
	"fmt"
	"os"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
	"github.com/marc-watters/the-block-chain-bar/v2/node"
	"github.com/marc-watters/the-block-chain-bar/v2/wallet"
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

	andrej := db.NewAccount(wallet.AndrejAccount)
	babayaga := db.NewAccount(wallet.BabayagaAccount)
	ceasar := db.NewAccount(wallet.CeasarAccount)

	pendingBlock := node.NewPendingBlock(
		db.Hash{},
		s.NextBlockHeight(),
		andrej,
		[]db.Trx{
			db.NewTrx(andrej, andrej, 3, ""),
			db.NewTrx(andrej, andrej, 700, "reward"),
			db.NewTrx(andrej, babayaga, 2000, ""),
			db.NewTrx(andrej, andrej, 100, "reward"),
			db.NewTrx(babayaga, andrej, 1, ""),
			db.NewTrx(babayaga, ceasar, 1000, ""),
			db.NewTrx(babayaga, andrej, 50, ""),
			db.NewTrx(andrej, andrej, 600, "reward"),
		},
	)

	if _, err := node.Mine(context.Background(), pendingBlock); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
