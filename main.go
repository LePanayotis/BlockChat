package main

import (
	"blockchat.io/bcc"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"os"
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

	rootCmd.Flags().BoolP("version", "v", false, "Print the version number of BlockChat")
	//rootCmd.Flags().BoolP("help", "h", false, "Help for BlockChat Command Line Interface (CLI)")

	// Set up the function to handle flags
	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		cmd.Usage()
		return err
	})

	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Starts the node",
		Args:  cobra.NoArgs,
		Long:  "Starts the node based on the configuration at the .env file",
		Run: func(cmd *cobra.Command, args []string) {
			//Loads environment variables from .env file
			godotenv.Load()

			lockfilePath := "myapp.lock"
			if _, err := os.Stat(lockfilePath); err == nil {
				// Lock file exists, indicating another instance is running
				fmt.Println("Another instance is already running.")
				return
			}
			lockfile, err := os.Create(lockfilePath)
			if err != nil {
				// Unable to create lock file
				fmt.Println("Error creating lock file:", err)
				return
			}
			defer lockfile.Close()
				

			//Starts the node
			bcc.StartNode()
			
		},
	}

	// var isit bool
	// tempcmd := &cobra.Command{
	// 	Use:"temp",
	// 	Run: func(cmd *cobra.Command, args []string) {
	// 		if isit {
	// 			fmt.Println(len(args))
	// 		}
	// 	},
	// }
	// tempcmd.Flags().BoolVar(&isit, "isit", false,"Enable THIS")
	// rootCmd.AddCommand(tempcmd)
	startCmd.Flags().BoolVarP(&bcc.CLI,"detached","d",false,"Run or no the CLI")

	rootCmd.AddCommand(
		startCmd,
		bcc.BalanceCmd,
		bcc.GenerateWalletCmd,
		bcc.PrintWalletCmd,
		bcc.ShowBlockchain,
		bcc.StakeCmd,
		bcc.StakeCmd,
		bcc.TransactionCmd,
		bcc.PrintWalletCmd,
		bcc.UseNodeWalletCmd,
		bcc.UseWalletCmd,
	)


	bcc.ConfigCmds()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
