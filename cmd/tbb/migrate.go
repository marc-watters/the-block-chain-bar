package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
	"github.com/marc-watters/the-block-chain-bar/v2/node"
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

			pendingBlock := node.NewPendingBlock(
				db.Hash{},
				state.NextBlockHeight(),
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
		},
	}

	addDefaultRequiredFlags(migrateCmd)

	return migrateCmd
}
