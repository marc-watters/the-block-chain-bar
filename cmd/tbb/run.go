package main

import (
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

			s, err := db.NewStateFromDisk(getDataDirFromCmd(cmd))
			if err != nil {
				fmt.Fprintf(os.Stderr, "error getting new state from disk: %v", err)
				os.Exit(1)
			}

			bootstrap := node.NewPeerNode(
				"127.0.0.1",
				8080,
				true,
				false,
			)

			n := node.New(s, ip, port, bootstrap)

			fmt.Println("Launching TBB node and its HTTP API...")
			if err := n.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "error launching node: %v", err)
				os.Exit(1)
			}
		},
	}

	addDefaultRequiredFlags(cmd)
	cmd.Flags().String(flagIP, node.DefaultIP, "exposed HTTP IP address for peer communications")
	cmd.Flags().Uint64(flagPort, node.DefaultHTTPort, "exposed HTTP port for peer communications")

	return cmd
}
