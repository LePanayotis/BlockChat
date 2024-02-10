package main

import (
	"os"
	"github.com/joho/godotenv"
	"blockchat.io/blockchat"
)

func main() {

	godotenv.Load()
	rootCmd := blockchat.RootCmd
	blockchat.SetStartFlags()


	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
