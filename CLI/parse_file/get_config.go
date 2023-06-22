package parse_file

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func GetOriginURL() (string, error) {
	filePath := "../.git/config"
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "url = ") {
			parts := strings.Split(line, " = ")
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("URL not found in the config file")
}

func ParseGitHubURL(url string) string {
	parts := strings.SplitN(url, "github.com:", 2)
	if len(parts) > 1 {
		url = strings.TrimSpace(parts[1])
	} else {
		parts = strings.SplitN(url, "github.com=", 2)
		if len(parts) > 1 {
			url = strings.TrimSpace(parts[1])
		}
	}
	url = strings.TrimSuffix(url, ".git")
	url = strings.ReplaceAll(url, "/", "_")
	return url
}

func PrintURL() {
	originURL, err := GetOriginURL()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	parsedURL := ParseGitHubURL(originURL)
	fmt.Println("Default project name:", parsedURL)
}

func Get_Config() *ffcli.Command {

	pcreate := &ffcli.Command{
		Name:       "config",
		ShortUsage: "config",
		ShortHelp:  "Get the default project name from .git/config",
		Exec: func(_ context.Context, args []string) error {

			if n := len(args); n != 0 {
				return fmt.Errorf("config requires 0 arguments but you provided %d", n)
			}
			PrintURL()
			return nil
		},
	}
	return pcreate
}
