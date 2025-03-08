package main

import (
	"fmt"
	"os"
	"time"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
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

	block0 := db.NewBlock(
		db.Hash{},
		0,
		uint64(time.Now().Unix()),
		[]db.Trx{
			db.NewTrx("andrej", "andrej", 3, ""),
			db.NewTrx("andrej", "andrej", 700, "reward"),
		},
	)

	block0Hash, err := s.AddBlock(block0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	block1 := db.NewBlock(
		block0Hash,
		1,
		uint64(time.Now().Unix()),
		[]db.Trx{
			db.NewTrx("andrej", "babayaga", 2000, ""),
			db.NewTrx("andrej", "andrej", 100, "reward"),
			db.NewTrx("babayaga", "andrej", 1, ""),
			db.NewTrx("babayaga", "caesar", 1000, ""),
			db.NewTrx("babayaga", "andrej", 50, ""),
			db.NewTrx("andrej", "andrej", 600, "reward"),
		},
	)

	block1Hash, err := s.AddBlock(block1)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Final block hash: %x", block1Hash)
}
