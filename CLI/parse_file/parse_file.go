package parse_file

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Configuration struct {
	Project string
	Token   string
}

func Parsefile() Configuration {
	var config Configuration
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("File reading error", err)
		return config
	}
	config.Project = os.Getenv("SECURE_ENV_PROJECT_NAME")
	config.Token = os.Getenv("SECURE_ENV_TOKEN")
	return config
}
