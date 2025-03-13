package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
	"github.com/marc-watters/the-block-chain-bar/v2/node"
	"github.com/marc-watters/the-block-chain-bar/v2/wallet"
)

func migrateCmd() *cobra.Command {
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrates the blockchain db according to new business rules.",
		Run: func(cmd *cobra.Command, args []string) {
			miner, err := cmd.Flags().GetString(flagMiner)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			ip, err := cmd.Flags().GetString(flagIP)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			port, err := cmd.Flags().GetUint64(flagPort)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			state, err := db.NewStateFromDisk(getDataDirFromCmd(cmd))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer state.Close()

			andrej := db.NewAccount(wallet.AndrejAccount)
			babayaga := db.NewAccount(wallet.BabayagaAccount)
			ceasar := db.NewAccount(wallet.CeasarAccount)

			pn := node.NewPeerNode(
				node.DefaultIP,
				node.DefaultHTTPort,
				true,
				andrej,
				false,
			)

			n := node.New(state, ip, port, db.NewAccount(miner), pn)
			n.AddPendingTrx(db.NewTrx(andrej, andrej, 3, ""), pn)
			n.AddPendingTrx(db.NewTrx(andrej, babayaga, 2000, ""), pn)
			n.AddPendingTrx(db.NewTrx(babayaga, andrej, 1, ""), pn)
			n.AddPendingTrx(db.NewTrx(babayaga, ceasar, 1000, ""), pn)
			n.AddPendingTrx(db.NewTrx(babayaga, andrej, 50, ""), pn)

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
			ticker := time.NewTicker(10 * time.Second)
			go func() {
				for range ticker.C {
					if !n.LatestBlockHash().IsEmpty() {
						ticker.Stop()
						cancel()
						return
					}
				}
			}()

			if err := n.Run(ctx); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}

	addDefaultRequiredFlags(migrateCmd)
	migrateCmd.Flags().String(flagMiner, node.DefaultMiner, "miner account of this node to receive block rewards")
	migrateCmd.Flags().String(flagIP, node.DefaultIP, "exposed IP for communication with peers")
	migrateCmd.Flags().Uint64(flagPort, node.DefaultHTTPort, "exposted HTTP port for communication with peers")

	return migrateCmd
}
