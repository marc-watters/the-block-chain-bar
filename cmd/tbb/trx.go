package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	db "github.com/marc-watters/the-block-chain-bar/v2/database"
)

const (
	flagFrom  = "from"
	flagTo    = "to"
	flagValue = "value"
	flagData  = "data"
)

func trxCmd() *cobra.Command {
	trxCmd := &cobra.Command{
		Use:   "trx",
		Short: "Interact with trxs (add...)",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return incorrectUsage()
		},
		Run: func(cmd *cobra.Command, args []string) {},
	}

	trxCmd.AddCommand(trxAddCmd())

	return trxCmd
}

func trxAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Adds new trx to database",
		Run: func(cmd *cobra.Command, args []string) {
			from, err := cmd.Flags().GetString(flagFrom)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			to, err := cmd.Flags().GetString(flagTo)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			value, err := cmd.Flags().GetUint64(flagValue)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			data, err := cmd.Flags().GetString(flagData)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			trx := db.NewTrx(db.NewAccount(from), db.NewAccount(to), value, data)

			s, err := db.NewStateFromDisk()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer s.Close()

			err = s.Add(trx)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			_, err = s.Persist()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			fmt.Println("trx successfully persisted to the ledger")
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

	cmd.Flags().Uint64(flagValue, 0, "Number of tokens to send")
	if err := cmd.MarkFlagRequired(flagValue); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	cmd.Flags().String(flagData, "", "Possible values: 'reward'")

	return cmd
}
