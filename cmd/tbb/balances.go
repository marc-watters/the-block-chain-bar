package main

import "github.com/spf13/cobra"

func balancesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "balances",
		Short: "Interact with balances",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
}
