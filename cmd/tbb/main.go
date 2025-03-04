package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

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
		balancesCmd(),
		trxCmd(),
		versionCmd(),
	)

	err := tbbCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
