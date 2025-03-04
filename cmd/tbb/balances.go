package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/marc-watters/the-block-chain-bar/v2/database"
	"github.com/spf13/cobra"
)

func balancesCmd() *cobra.Command {
	balancesCmd := &cobra.Command{
		Use:   "balances",
		Short: "Interact with balances",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return incorrectUsage()
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
		fmt.Printf("%s\n", strings.Repeat("=", 72))
		fmt.Printf("%[1]s Account Balances %[1]s\n", strings.Repeat(" ", 27))
		fmt.Printf("%s\n", strings.Repeat("=", 72))

		w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
		for k, v := range s.Balances {
			fmt.Fprintf(w, " |>\t%s\t%d\n", k, v)
		}
		w.Flush()

		fmt.Printf("%s", strings.Repeat("-", 72))
		fmt.Println()
	},
}
