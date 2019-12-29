package main

import (
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	a := app{}
	a.Initialize(os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))

	a.Run(":5000")
}
