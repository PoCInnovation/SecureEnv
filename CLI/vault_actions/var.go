package vault_actions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func secret_get(name string, mainUrl string) {

	url := mainUrl + "/" + name + "/var"

	println(url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	sb := string(body)
	log.Print(sb)
}

func secret_create(name string, key string, value string, mainUrl string) {

	url := mainUrl + "/" + name + "/var/" + key
	println(url)

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

	if resp.StatusCode < http.StatusBadRequest {
		fmt.Println("Secret créé avec succès.")
	} else {
		fmt.Println("La création du secret à échoué. Code de statut :", resp.StatusCode)
		return
	}
	body, _ := ioutil.ReadAll(resp.Body)
	println(body)
	defer resp.Body.Close()

	println("Secret created successfully")
}

func secret_delete(name string, key string, mainUrl string) {

	url := mainUrl + "/" + name + "/var/" + key
	println(url)
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

	if resp.StatusCode < http.StatusBadRequest {
		fmt.Println("La ressource a été supprimée avec succès.")
	} else {
		fmt.Println("La suppression de la ressource a échoué. Code de statut :", resp.StatusCode)
		return
	}
	body, _ := ioutil.ReadAll(resp.Body)
	println(body)
	defer resp.Body.Close()

	println("Secret deleted")
}

func secret_edit(name string, key string, value string, mainUrl string) {

	url := mainUrl + "/" + name + "/var/" + key
	println(url)

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

	if resp.StatusCode < http.StatusBadRequest {
		fmt.Println("Secret modifié avec succès.")
	} else {
		fmt.Println("La modification du secret à échoué. Code de statut :", resp.StatusCode)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	println(body)
	defer resp.Body.Close()

	println("Secret modify successfully")
}
