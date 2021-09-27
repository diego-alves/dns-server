package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/diego-alves/dns-server/pkg/api"
	"github.com/diego-alves/dns-server/pkg/docker"
	"github.com/diego-alves/dns-server/pkg/hosts"
	"github.com/gorilla/mux"
)

func main() {
	log.Println("v3")
	var entries []hosts.Entry
	var hostsFile hosts.HostsFile

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		res, err := json.Marshal(entries)
		if err != nil {
			fmt.Println("error json")
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(res))
	})
	go http.ListenAndServe(":8081", r)
	log.Println("Http Server UP! port 8081")

	apiEntries, err := api.GetEntries()
	if err != nil {
		panic(err)
	}
	for _, apiEntry := range apiEntries {
		hostsFile.Add(apiEntry)
		entries = append(entries, apiEntry)
	}
	log.Println("Hosts from API ok!")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		s := <-sig
		fmt.Println("Exit", s)
		os.Exit(0)
	}()
	log.Println("Signal Term OK!")

	onStart := make(chan hosts.Entry, 10)
	go func() {
		for entry := range onStart {
			hostsFile.Add(entry)
			entries = append(entries, entry)
		}
	}()

	onKill := make(chan string, 10)
	go func() {
		for ip := range onKill {
			hostsFile.Remove(ip)
			for i := 0; i < len(entries); i++ {
				if entries[i].IpAddress == ip {
					entries = append(entries[:i], entries[i+1:]...)
				}
			}

		}
	}()
	log.Println("Start and Kill listeners ok")

	fmt.Println("init")
	listener := docker.NewDockerListeger(onStart, onKill)
	listener.Init(docker.START)
	listener.Listen()
	log.Println("Listen Docker events ok")

}
