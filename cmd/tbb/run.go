package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
	node "github.com/marc-watters/the-block-chain-bar/v2/node"
)

func runCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Launches the TBB node and its HTTP API",
		Run: func(cmd *cobra.Command, args []string) {
			miner, err := cmd.Flags().GetString("miner")
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

			bootstrapIP, err := cmd.Flags().GetString(flagBootstrapIP)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			bootstrapPort, err := cmd.Flags().GetUint64(flagBootstrapPort)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			bootstrapAcc, err := cmd.Flags().GetString(flagBootstrapAcc)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			s, err := db.NewStateFromDisk(getDataDirFromCmd(cmd))
			if err != nil {
				fmt.Fprintf(os.Stderr, "error getting new state from disk: %v", err)
				os.Exit(1)
			}

			bootstrap := node.NewPeerNode(
				bootstrapIP,
				bootstrapPort,
				true,
				db.NewAccount(bootstrapAcc),
				false,
			)

			n := node.New(
				s,
				ip,
				port,
				db.NewAccount(miner),
				bootstrap)

			fmt.Println("Launching TBB node and its HTTP API...")
			if err := n.Run(context.Background()); err != nil {
				fmt.Fprintf(os.Stderr, "error launching node: %v", err)
				os.Exit(1)
			}
		},
	}

	addDefaultRequiredFlags(cmd)
	cmd.Flags().String(flagMiner, node.DefaultMiner, "miner account of this node to receive block awards")
	cmd.Flags().String(flagIP, node.DefaultIP, "exposed HTTP IP address for peer communications")
	cmd.Flags().Uint64(flagPort, node.DefaultHTTPort, "exposed HTTP port for peer communications")
	cmd.Flags().String(flagBootstrapIP, node.DefaultBootstrapIP, "default bootstrap server to interconnect peers")
	cmd.Flags().Uint64(flagBootstrapPort, node.DefaultBootstrapPort, "default bootstrap server port to interconnect peers")
	cmd.Flags().String(flagBootstrapAcc, node.DefaultBootstrapAcc, "default bootstrap account to interconnect peers")

	return cmd
}
