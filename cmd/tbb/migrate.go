package main

import (
	"fmt"
	"os"
	"tbb/v2/database"
	"time"

	"github.com/spf13/cobra"
)

var migrateCmd = func() *cobra.Command {
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrates the blockchain database according to new business rules.",
		Run: func(cmd *cobra.Command, args []string) {
			s, err := database.NewStateFromDisk(getDataDirFromCmd(cmd))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer s.Close()

			b0 := database.NewBlock(
				database.Hash{},
				s.NextBlockHeight(),
				uint64(time.Now().Unix()),
				[]database.Tx{
					database.NewTx("andrej", "andrej", 3, ""),
				},
			)

			b0Hash, err := s.AddBlock(b0)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			b1 := database.NewBlock(
				b0Hash,
				s.NextBlockHeight(),
				uint64(time.Now().Unix()),
				[]database.Tx{
					database.NewTx("andrej", "babayaga", 2000, ""),
					database.NewTx("andrej", "andrej", 100, "reward"),
					database.NewTx("babayaga", "andrej", 1, ""),
					database.NewTx("babayaga", "caesar", 1000, ""),
					database.NewTx("babayaga", "andrej", 50, ""),
					database.NewTx("andrej", "andrej", 600, "reward"),
				},
			)

			b1Hash, err := s.AddBlock(b1)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			b2 := database.NewBlock(
				b1Hash,
				s.NextBlockHeight(),
				uint64(time.Now().Unix()),
				[]database.Tx{
					database.NewTx("andrej", "andrej", 24700, "reward"),
				},
			)

			_, err = s.AddBlock(b2)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}

	addDefaultRequiredFlags(migrateCmd)

	return migrateCmd
}
