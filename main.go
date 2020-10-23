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

var (
	Token      string
	GuildID    string
	TChannelID string
	VChannelID string
	Folder     string
	err        error
)

func main() {
	fmt.Println("Hello, Pondashi!")
}
