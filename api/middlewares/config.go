package middlewares

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var adr_vault = ""

func getEnvFile() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	adr_vault = os.Getenv("SECURE_ENV_HOST")
}
