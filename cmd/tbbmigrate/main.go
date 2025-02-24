package main

import (
	"encoding/json"
	"fmt"
	"os"
	"tbb/v2/database"
	"time"

	"github.com/spf13/afero"
)

func init() {
	database.AppFs = &afero.Afero{Fs: afero.NewOsFs()}
}

func main() {
	state, err := database.NewStateFromDisk("../database")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer state.Close()

	block0 := database.NewBlock(
		database.Hash{},
		state.LatestBlock().Header.Height+1,
		uint64(time.Now().Unix()),
		[]database.Tx{
			database.NewTx("andrej", "andrej", 3, ""),
		},
	)

	if err := state.AddBlock(block0); err != nil {
		fmt.Fprintf(os.Stderr, "error adding block: %v", err)
		os.Exit(1)
	}

	b0Hash, err := state.Persist()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error persisting block: %v", err)
		os.Exit(1)
	}

	block1 := database.NewBlock(
		b0Hash,
		state.LatestBlock().Header.Height+1,
		uint64(time.Now().Unix()),
		[]database.Tx{
			database.NewTx("andrej", "andrej", 700, "reward"),
			database.NewTx("andrej", "babayaga", 2000, ""),
			database.NewTx("andrej", "andrej", 100, "reward"),
			database.NewTx("babayaga", "andrej", 1, ""),
			database.NewTx("babayaga", "ceasar", 1000, ""),
			database.NewTx("babayaga", "andrej", 50, ""),
			database.NewTx("andrej", "andrej", 600, "reward"),
		},
	)

	if err := state.AddBlock(block1); err != nil {
		fmt.Fprintf(os.Stderr, "error adding block: %v", err)
		os.Exit(1)
	}

	b1Hash, err := state.Persist()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error persisting block: %v", err)
		os.Exit(1)
	}

	json, err := json.MarshalIndent(b1Hash, "", " ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error pretty printing hash: %v", err)
		os.Exit(1)
	}

	fmt.Printf("%s\n", json)
}
