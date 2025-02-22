package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func balancesCmd() *cobra.Command {
	balancesCmd := &cobra.Command{
		Use:   "balances",
		Short: "Interact with balances",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("incorrect usage")
		},
		Run: func(cmd *cobra.Command, args []string) {},
	}

	return balancesCmd
}
