package main

import (
	"myproject/bcc"
	"github.com/joho/godotenv"
)


func main() {
	//Loads environment variables from .env file
	godotenv.Load()	

	//Starts the node
	bcc.StartNode()
}