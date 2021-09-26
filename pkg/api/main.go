package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func GetHosts() (map[string][]string, error) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:8080/hosts.json", nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	hosts := make(map[string][]string)

	json.Unmarshal(body, &hosts)

	return hosts, nil

}
