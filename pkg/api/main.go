package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/diego-alves/dns-server/pkg/hosts"
)

func GetEntries() ([]hosts.Entry, error) {
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

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var entries []hosts.Entry
	json.Unmarshal(data, &entries)
	for i := 0; i < len(entries); i++ {
		entries[i].Source = "api"
	}
	return entries, nil

}
