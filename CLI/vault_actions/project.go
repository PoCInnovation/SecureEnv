package vault_actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func project_list(mainUrl string) {

	url := mainUrl + "/"

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
	fmt.Println(string(prettyJSON))
}

func project_get(name string, mainUrl string) {

	url := mainUrl + "/" + name

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
	fmt.Println(string(prettyJSON))
}

func project_create(name string, mainUrl string) {

	url := mainUrl + "/"

	bodyjson := map[string]interface{}{
		"Value": name,
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
		fmt.Print("Failed to create project \"", name, "\". Status code: ", resp.StatusCode, "\n")
		return
	}

	fmt.Print("Project \"", name, "\" created successfully\n")
}

func project_delete(name string, mainUrl string) {

	url := mainUrl + "/" + name

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
		fmt.Print("Failed to delete the project \"", name, "\". Status code: ", resp.StatusCode, "\n")
		return
	}

	fmt.Print("Project \"", name, "\" deleted\n")
}

func project_edit(name string, value string, mainUrl string) {

	url := mainUrl + "/" + name

	bodyjson := map[string]interface{}{
		"Value": value,
	}
	jsonData, _ := json.Marshal(bodyjson)

	req, res := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
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
		fmt.Print("Failed to modify the project \"", name, "\". Status code: ", resp.StatusCode, "\n")
		return
	}

	fmt.Print("Project \"", name, "\" renamed successfully\n")
}

func project_update(name string, bodyjson map[string]interface{}, mainUrl string) {

	url := mainUrl + "/" + name

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
		fmt.Print("Failed to update the project \"", name, "\". Status code: ", resp.StatusCode, "\n")
		return
	}

	fmt.Print("Project \"", name, "\" updated successfully\n")
}
