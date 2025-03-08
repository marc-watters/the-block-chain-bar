package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
)

func migrateCmd() *cobra.Command {
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrates the blockchain db according to new business rules.",
		Run: func(cmd *cobra.Command, args []string) {
			state, err := db.NewStateFromDisk(getDataDirFromCmd(cmd))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer state.Close()

			block0 := db.NewBlock(
				db.Hash{},
				0,
				uint64(time.Now().Unix()),
				[]db.Trx{
					db.NewTrx("andrej", "andrej", 3, ""),
					db.NewTrx("andrej", "andrej", 700, "reward"),
				},
			)

			block0hash, err := state.AddBlock(block0)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			block1 := db.NewBlock(
				block0hash,
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

			block1hash, err := state.AddBlock(block1)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			block2 := db.NewBlock(
				block1hash,
				2,
				uint64(time.Now().Unix()),
				[]db.Trx{
					db.NewTrx("andrej", "andrej", 24700, "reward"),
				},
			)

			block2hash, err := state.AddBlock(block2)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}

			fmt.Printf("Final block hash: %x", block2hash)
		},
	}

	addDefaultRequiredFlags(migrateCmd)

	return migrateCmd
}
