package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	Major  = "1"
	Minor  = "3"
	Fix    = "0"
	Verbal = "Add trx command"
)

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Describes version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s.%s.%s-beta %s\n", Major, Minor, Fix, Verbal)
		},
	}
}
