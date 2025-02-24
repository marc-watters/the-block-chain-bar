package main

import (
	"fmt"
	"os"
	"tbb/v2/database"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func init() {
	database.AppFs = &afero.Afero{Fs: afero.NewOsFs()}
}

const flagDataDir = "datadir"

func main() {
	tbbCmd := &cobra.Command{
		Use:   "tbb",
		Short: "The Blockchain Bar CLI",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return incorrectUsageErr()
		},
		Run: func(cmd *cobra.Command, args []string) {},
	}

	tbbCmd.AddCommand(
		versionCmd,
		balancesCmd(),
		runCmd(),
	)

	err := tbbCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func addDefaultRequiredFlags(cmd *cobra.Command) {
	cmd.Flags().String(flagDataDir, "", "Absolute path to the node data dir where the DB will/is stored")
	if err := cmd.MarkFlagRequired(flagDataDir); err != nil {
		fmt.Fprintf(os.Stderr, "error marking %s flag required: %v", flagDataDir, err)
	}
}

func incorrectUsageErr() error {
	return fmt.Errorf("incorrect usage")
}
