package vault_actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"secureenv/parse_file"
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

func Project_pull(config parse_file.Configuration, bodyjson map[string]interface{}) {

	if err := os.Truncate(".env", 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	f, err := os.Create(".env")
	if err != nil {
		log.Fatal(err)
	}

	_, err = f.WriteString("SECURE_ENV_PROJECT_NAME=\"" + config.Project + "\"\n")
	if err != nil {
		log.Fatal(err)
	}

	_, err = f.WriteString("SECURE_ENV_TOKEN=\"" + config.Token + "\"\n")

	if err != nil {
		log.Fatal(err)
	}

	for key, value := range bodyjson {
		_, err = f.WriteString(key + "=\"" + value.(string) + "\"\n")
		if err != nil {
			log.Fatal(err)
		}
	}
}

func Project_clone(mainUrl string) (string, error) {
	name_project, err := parse_file.Get_URL()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return "", err
	}

	url := mainUrl + "/" + name_project
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	expectedJSON := `{"error":"Project not Found"}`
	if string(body) == expectedJSON {
		return "", fmt.Errorf("default project %s does not exist, please create the project before cloning", name_project)
	}

	bodyjson := Secret_get(name_project, mainUrl, 0)
	if err := os.Truncate(".env", 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	f, err := os.Create(".env")
	if err != nil {
		log.Fatal(err)
	}

	_, err = f.WriteString("SECURE_ENV_PROJECT_NAME=\"" + name_project + "\"\n")
	if err != nil {
		log.Fatal(err)
	}

	for key, value := range bodyjson {
		_, err = f.WriteString(key + "=\"" + value.(string) + "\"\n")
		if err != nil {
			log.Fatal(err)
		}
	}

	return "", nil
}
