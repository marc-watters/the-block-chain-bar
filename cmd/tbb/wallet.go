package main

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/console/prompt"
	"github.com/spf13/cobra"

	"github.com/marc-watters/the-block-chain-bar/v2/wallet"
)

func walletCmd() *cobra.Command {
	walletCmd := &cobra.Command{
		Use:   "wallet",
		Short: "Manages accounts, keys, and cryptography",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return incorrectUsage()
		},
		Run: func(cmd *cobra.Command, args []string) {},
	}

	walletCmd.AddCommand(walletNewAccountCmd())

	return walletCmd
}

func walletNewAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new-account",
		Short: "Creates a new account with a new set of elliptic-curve private + public keys",
		Run: func(cmd *cobra.Command, args []string) {
			password := getPassPhrase("Please enter a password to encrypt the new wallet: ", true)
			dataDir := getDataDirFromCmd(cmd)

			ks := keystore.NewKeyStore(wallet.GetKeystoreDirPath(dataDir), keystore.StandardScryptN, keystore.StandardScryptP)
			acc, err := wallet.NewKeyStoreAccount(dataDir, password)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Printf("New account created: %s\n", acc.Hex())
		},
	}

	addDefaultRequiredFlags(cmd)

	return cmd
}

func getPassPhrase(promptString string, confirmation bool) string {
	fmt.Println(promptString)
	password, err := prompt.Stdin.PromptPassword("Password: ")
	if err != nil {
		utils.Fatalf("Failed to read password: %v", err)
	}

	if confirmation {
		confirm, err := prompt.Stdin.PromptPassword("Confirm password: ")
		if err != nil {
			utils.Fatalf("Failed to confirm password: %v", err)
		}

		if password != confirm {
			utils.Fatalf("Password do not match")
		}
	}

	return password
}
