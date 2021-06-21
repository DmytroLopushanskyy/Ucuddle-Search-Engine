package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"path"
	"strconv"
)

var NUM_LINKS_PER_CRAWLER int

func getGlobalVar() {
	fmt.Println("NUM_LINKS_PER_CRAWLER -- ", NUM_LINKS_PER_CRAWLER)
}


func main() {
	err := godotenv.Load(path.Join("shared_vars.env"))
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	NUM_LINKS_PER_CRAWLER, _ = strconv.Atoi(os.Getenv("NUM_LINKS_PER_CRAWLER"))

	getGlobalVar()
}