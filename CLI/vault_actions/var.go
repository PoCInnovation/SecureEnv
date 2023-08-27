package vault_actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func Secret_get(name string, mainUrl string, printable int) map[string]interface{} {

	url := mainUrl + "/" + name + "/var"

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var data map[string]interface{}
	_ = json.Unmarshal(body, &data)
	prettyJSON, _ := json.MarshalIndent(data, "", "  ")

	if printable == 1 {
		fmt.Println(string(prettyJSON))
	}
	return data
}

func Secret_create(name string, key string, value string, mainUrl string) {

	url := mainUrl + "/" + name + "/var/" + key

	bodyjson := map[string]interface{}{
		"Value": value,
	}
	jsonData, _ := json.Marshal(bodyjson)

	req, res := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	if res != nil {
		fmt.Println(res)
		return
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		fmt.Print("Failed to create secret \"", key, "\". Status code :", resp.StatusCode, "\n")
		return
	}

	fmt.Print("Secret \"", key, "\" created successfully\n")
}

func Secret_delete(name string, key string, mainUrl string) {

	url := mainUrl + "/" + name + "/var/" + key

	req, res := http.NewRequest("DELETE", url, nil)
	if res != nil {
		fmt.Println(res)
		return
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		fmt.Print("Failed to delete the secret \"", key, "\". Status code :", resp.StatusCode)
		return
	}

	fmt.Print("Secret \"", key, "\" deleted\n")
}

func secret_edit(name string, key string, value string, mainUrl string) {

	url := mainUrl + "/" + name + "/var/" + key

	bodyjson := map[string]interface{}{
		"Value": value,
	}
	jsonData, _ := json.Marshal(bodyjson)

	req, res := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	if res != nil {
		fmt.Println(res)
		return
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		fmt.Print("Failed to modify the secret \"", key, "\". Status code :", resp.StatusCode, "\n")
		return
	}

	fmt.Print("Secret \"", key, "\" modified successfully\n")
}
