package controller

import (
	"fmt"
	"net/http"
)

func CheckAPIStatus(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("ERROR : Failed to connect to the API : %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusInternalServerError {
		return fmt.Errorf("FAILED : API returned internal server error : %d", resp.StatusCode)
	}
	return nil
}
