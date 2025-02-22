package main

import (
	"fmt"
	"os"
	"tbb/v2/database"

	"github.com/spf13/cobra"
)

const (
	flagFrom  = "from"
	flagTo    = "to"
	flagValue = "value"
	flagData  = "data"
)

func txCmd() *cobra.Command {
	txsCmd := &cobra.Command{
		Use:   "tx",
		Short: "Interact with transactions (add...)",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("incorrect usage")
		},
		Run: func(cmd *cobra.Command, args []string) {},
	}

	txsCmd.AddCommand(txAddCmd())

	return txsCmd
}

func txAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Adds new TX to database",
		Run: func(cmd *cobra.Command, args []string) {
			from, err := cmd.Flags().GetString(flagFrom)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			to, err := cmd.Flags().GetString(flagTo)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			value, err := cmd.Flags().GetUint(flagValue)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			data, err := cmd.Flags().GetString(flagData)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			tx := database.NewTx(database.NewAccount(from), database.NewAccount(to), value, data)

			s, err := database.NewStateFromDisk()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer s.Close()

			err = s.Add(tx)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			err = s.Persist()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			fmt.Println("TX successfully added to the ledger")
		},
	}

	cmd.Flags().String(flagFrom, "", "Which account to send tokens from")
	if err := cmd.MarkFlagRequired(flagFrom); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	cmd.Flags().String(flagTo, "", "Which account to send tokens to")
	if err := cmd.MarkFlagRequired(flagTo); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	cmd.Flags().Uint(flagValue, 0, "Number of tokens to send")
	if err := cmd.MarkFlagRequired(flagValue); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	cmd.Flags().String(flagData, "", "Possibles values: 'reward'")

	return cmd
}
