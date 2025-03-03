package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	Major  = "1"
	Minor  = "0"
	Fix    = "0"
	Verbal = "Add tbb command"
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
