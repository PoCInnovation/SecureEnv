package parse_file

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func FindNearestGitDir() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		gitDir := filepath.Join(currentDir, ".git")
		_, err := os.Stat(gitDir)
		if err == nil {
			return gitDir, nil
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			break // Root directory reach
		}

		currentDir = parentDir
	}

	return "", fmt.Errorf("no .git directory found in any parent directory")
}

func GetOriginURL() (string, error) {
	gitDir, err := FindNearestGitDir()
	if err != nil {
		return "", err
	}
	filePath := filepath.Join(gitDir, "config")
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

func Get_URL() (string, error) {
	originURL, err := GetOriginURL()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return "", err
	}
	parsedURL := ParseGitHubURL(originURL)
	return parsedURL, nil
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
