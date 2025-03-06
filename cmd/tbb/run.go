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
			dataDir, err := cmd.Flags().GetString(flagDataDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading dataDir flag: %v", err)
				os.Exit(1)
			}

			s, err := db.NewStateFromDisk(dataDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error getting new state from disk: %v", err)
				os.Exit(1)
			}

			fmt.Println("Launching TBB node and its HTTP API...")

			n := node.New(s)
			if err := n.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "error launching node: %v", err)
				os.Exit(1)
			}
		},
	}

	addDefaultRequiredFlags(cmd)

	return cmd
}
