package main

import (
	"fmt"
	"os"
	"blockchat.io/bcc"
	"github.com/spf13/cobra"
)

func main() {

	rootCmd := &cobra.Command{
		Use:   "blockchat",
		Short: "BlockChat is a simple CLI application",
		Long:  `BlockChat is a simple CLI application built using Cobra.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hi")
		},
	}

	rootCmd.AddCommand(
		bcc.StartCmd,
		bcc.BalanceCmd,
		bcc.GenerateWalletCmd,
		bcc.ShowBlockchain,
		bcc.StopCmd,
		bcc.StakeCmd,
		bcc.TransactionCmd,
		bcc.PrintWalletCmd,
		bcc.UseNodeWalletCmd,
		bcc.UseWalletCmd,
	)
	bcc.SetStartFlags()

	bcc.ConfigCmds()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
