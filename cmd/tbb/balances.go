package main

import (
	"fmt"
	"os"
	"tbb/v2/database"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func balancesCmd() *cobra.Command {
	balancesCmd := &cobra.Command{
		Use:   "balances",
		Short: "Interact with balances",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return incorrectUsageErr()
		},
		Run: func(cmd *cobra.Command, args []string) {},
	}

	balancesCmd.AddCommand(balancesListCmd)

	return balancesCmd
}

var balancesListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all balances",
	Run: func(cmd *cobra.Command, args []string) {
		s, err := database.NewStateFromDisk()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer s.Close()

		fmt.Println()
		fmt.Println("    Account Balances    ")
		fmt.Println("________________________")

		w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
		for a, b := range s.Balances {
			fmt.Fprintf(w, "* %s\t|\t%d\n", a, b)
		}
		w.Flush()

		fmt.Println("------------------------")
		fmt.Println()
	},
}
