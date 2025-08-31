package config

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadEnvVariables(filename string) {
	err := godotenv.Overload(filename)
	if err != nil {
		log.Fatalf("Error while loading env file: %s", filename)
	}
}
