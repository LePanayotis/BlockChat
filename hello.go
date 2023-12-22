package main

import (
	"myproject/bcc"

	"github.com/joho/godotenv"
)


func main() {
	godotenv.Load()	
	bcc.StartNode()
}