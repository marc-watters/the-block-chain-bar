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

func main() {
	tbbCmd := &cobra.Command{
		Use:   "tbb",
		Short: "The Blockchain Bar CLI",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("incorrect usage")
		},
		Run: func(cmd *cobra.Command, args []string) {},
	}

	tbbCmd.AddCommand(
		versionCmd,
		balancesCmd(),
	)

	err := tbbCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
