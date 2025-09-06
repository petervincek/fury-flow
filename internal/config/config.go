package config

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadEnvVariables(filename string) {
	err := godotenv.Load(filename)
	if err != nil {
		log.Fatalf("Error while loading env file: %s, %v", filename, err)
	}
}
