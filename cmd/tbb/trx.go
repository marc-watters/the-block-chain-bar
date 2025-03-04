package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	flagFrom  = "from"
	flagTo    = "to"
	flagValue = "value"
)

func trxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trx",
		Short: "Interact with trxs",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("incorrect usage")
		},
		Run: func(cmd *cobra.Command, args []string) {},
	}

	return cmd
}
