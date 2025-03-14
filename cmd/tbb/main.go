package main

import (
	"fmt"
	"os"

	"github.com/marc-watters/the-block-chain-bar/v2/fs"
	"github.com/spf13/cobra"
)

const (
	flagDataDir = "datadir"
	flagIP      = "ip"
	flagPort    = "port"
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
		trxCmd(),
		runCmd(),
		migrateCmd(),
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
