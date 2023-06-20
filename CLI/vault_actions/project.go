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

func project_get(name string, mainUrl string) {

	url := mainUrl + "/" + name

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

func project_create(name string, mainUrl string) {

	url := mainUrl + "/"
	println(url)

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

	if resp.StatusCode < http.StatusBadRequest {
		fmt.Println("Projet crée avec succès.")
	} else {
		fmt.Println("La création du projet à échoué. Code de statut :", resp.StatusCode)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	println(body)
	defer resp.Body.Close()

	println("Project created successfully")
}

func project_delete(name string, mainUrl string) {

	url := mainUrl + "/" + name
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
		fmt.Println("Le project a été supprimée avec succès.")
	} else {
		fmt.Println("La suppression du projet a échoué. Code de statut :", resp.StatusCode)
		return
	}
	body, _ := ioutil.ReadAll(resp.Body)
	println(body)
	defer resp.Body.Close()

	println("Project deleted")
}

func project_edit(name string, value string, mainUrl string) {

	url := mainUrl + "/" + name
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
		fmt.Println("Projet renommé avec succès.")
	} else {
		fmt.Println("La modification du projet à échoué. Code de statut :", resp.StatusCode)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	println(body)
	defer resp.Body.Close()

	println("Secret renamed successfully")
}
