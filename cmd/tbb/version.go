package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	Major  = "0"
	Minor  = "5"
	Fix    = "0"
	Verbal = "HTTP API"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Describes version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s.%s.%s-beta %s\n", Major, Minor, Fix, Verbal)
	},
}
