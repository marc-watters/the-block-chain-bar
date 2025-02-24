package main

import (
	"fmt"
	"os"
	"tbb/v2/node"

	"github.com/spf13/cobra"
)

func runCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Launcheds the TBB node and its HTTP API",
		Run: func(cmd *cobra.Command, args []string) {
			dataDir, err := cmd.Flags().GetString(flagDataDir)
			if err != nil {
				fmt.Fprint(os.Stderr, err)
				os.Exit(1)
			}

			fmt.Println("Launching TBB node and its HTTP API...")

			err = node.Run(dataDir)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	addDefaultRequiredFlags(runCmd)

	return runCmd
}
