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
	err := godotenv.Load("secure-env.env")
	if err != nil {
		fmt.Println("File reading error", err)
		return config
	}
	config.Project = os.Getenv("PROJECT")
	config.Host = os.Getenv("HOST")
	config.Port, err = strconv.Atoi(os.Getenv("PORT"))
	config.Token = os.Getenv("TOKEN")
	return config
}
