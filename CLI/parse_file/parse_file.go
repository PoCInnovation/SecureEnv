package parse_file

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Configuration struct {
	Project string
	Host    string
	Port    int
	Token   string
}

func Parsefile() Configuration {
	var config Configuration
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("File reading error", err)
		return config
	}
	config.Project = os.Getenv("SECURE_ENV_PROJECT")
	config.Host = os.Getenv("SECURE_ENV_HOST")
	config.Port, err = strconv.Atoi(os.Getenv("SECURE_ENV_PORT"))
	config.Token = os.Getenv("SECURE_ENV_TOKEN")
	return config
}
