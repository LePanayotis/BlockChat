package main

import (
	// "fmt"
	// "log"
	"myproject/bcc"

	"github.com/joho/godotenv"
)


func main() {
	godotenv.Load()	
	// s := os.Environ()
	// for _, i := range s {
	// 	log.Printf(i)
	// }
	// time.Sleep(20*time.Second)
	bcc.StartNode()
}