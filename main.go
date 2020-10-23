package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	if os.Getenv("GO_ENV") == "" {
		os.Setenv("GO_ENV", "dev")
	}

	err := godotenv.Load(".env%s", os.Getenv("GO_ENV"))
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	fmt.Println("Hello, Pondashi!")
}
