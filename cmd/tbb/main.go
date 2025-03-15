package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/marc-watters/the-block-chain-bar/v2/fs"
)

const (
	flagDataDir       = "datadir"
	flagMiner         = "miner"
	flagIP            = "ip"
	flagPort          = "port"
	flagBootstrapAcc  = "bootstrap-account"
	flagBootstrapIP   = "bootstrap-ip"
	flagBootstrapPort = "bootstrap-port"
)

func main() {
	tbbCmd := &cobra.Command{
		Use:   "tbb",
		Short: "The Blockchain Bar CLI",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return incorrectUsage()
		},
		Run: func(cmd *cobra.Command, args []string) {},
	}

	tbbCmd.AddCommand(
		balancesCmd(),
		runCmd(),
		walletCmd(),
		versionCmd(),
	)

	err := tbbCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func addDefaultRequiredFlags(cmd *cobra.Command) {
	cmd.Flags().String(flagDataDir, "", "Absolute path to the node data directory where the database will be stored")
	err := cmd.MarkFlagRequired(flagDataDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func getDataDirFromCmd(cmd *cobra.Command) string {
	dataDir, _ := cmd.Flags().GetString(flagDataDir)

	return fs.ExpandPath(dataDir)
}

func incorrectUsage() error {
	return fmt.Errorf("incorrect usage")
}
