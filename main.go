package main

import (
	"github.com/diego-alves/dns-server/pkg/docker"
)

func main() {
	var listener docker.DockerListener
	listener.Init()
}
