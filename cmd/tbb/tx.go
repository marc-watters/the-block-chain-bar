package main

import (
	"github.com/spf13/cobra"
)

func txCmd() *cobra.Command {
	txsCmd := &cobra.Command{
		Use:   "tx",
		Short: "Interact with transactions",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return incorrectUsageErr()
		},
		Run: func(cmd *cobra.Command, args []string) {},
	}

	return txsCmd
}
