package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/diego-alves/dns-server/pkg/docker"
	"github.com/diego-alves/dns-server/pkg/hosts"
)

func main() {
	entries := make(map[string][]string)
	var hostsFile hosts.HostsFile

	sig := make(chan os.Signal, 1)
	signal.Notify(sig,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

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
	go func() {
		s := <-sig
		fmt.Println(s)
		listener.Init(docker.KILL)
	}()

	listener.Init(docker.START)
	listener.Listen()

}
