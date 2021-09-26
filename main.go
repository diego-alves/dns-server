package main

import (
	"fmt"

	"github.com/diego-alves/dns-server/pkg/api"
	"github.com/diego-alves/dns-server/pkg/docker"
	"github.com/diego-alves/dns-server/pkg/hosts"
)

func main() {
	entries := make(map[string][]string)
	var hostsFile hosts.HostsFile

	hosts, err := api.GetHosts()
	if err != nil {
		panic(err)
	}
	fmt.Println(hosts)
	for ip, hosts := range hosts {
		hostsFile.Add(append([]string{ip}, hosts...))
	}

	// sig := make(chan os.Signal, 1)
	// signal.Notify(sig,
	// 	syscall.SIGHUP,
	// 	syscall.SIGINT,
	// 	syscall.SIGTERM,
	// 	syscall.SIGQUIT)

	onStart := make(chan []string, 10)
	go func() {
		for entry := range onStart {
			hostsFile.Add(entry)
			entries[entry[0]] = entry[1:]
		}
	}()

	onKill := make(chan string, 10)
	go func() {
		for ip := range onKill {
			hostsFile.Remove(ip)
			delete(entries, ip)
		}
	}()

	fmt.Println("init")
	listener := docker.NewDockerListeger(onStart, onKill)
	// go func() {
	// 	s := <-sig
	// 	fmt.Println(s)
	// 	listener.Init(docker.KILL)
	// }()

	listener.Init(docker.START)
	listener.Listen()

}
