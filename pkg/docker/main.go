package docker

import (
	"context"
	"fmt"

	"github.com/diego-alves/dns-server/pkg/hosts"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type Action string

type DockerListener struct {
	ctx     context.Context
	cli     *client.Client
	onStart chan<- hosts.Entry
	onKill  chan<- string
}

const (
	START Action = "start"
	KILL  Action = "kill"
)

func NewDockerListeger(onStart chan<- hosts.Entry, onKill chan<- string) *DockerListener {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	l := new(DockerListener)
	l.ctx = context.Background()
	l.onKill = onKill
	l.onStart = onStart
	l.cli = cli
	return l
}

func (l *DockerListener) entries(id string) (map[string][]string, error) {
	entries := make(map[string][]string)
	json, err := l.cli.ContainerInspect(l.ctx, id)
	if err != nil {
		return nil, err
	}

	hostname := json.Config.Hostname
	if json.NetworkSettings.IPAddress != "" {
		entries[json.NetworkSettings.IPAddress] = []string{hostname}
	}

	for _, network := range json.NetworkSettings.Networks {
		if len(network.Aliases) > 0 {
			hosts := append([]string{hostname}, network.Aliases...)
			if val, ok := entries[network.IPAddress]; ok {
				hosts = append(hosts, val...)
			}
			entries[network.IPAddress] = hosts
		}
	}

	return entries, nil
}

func (l *DockerListener) notify(entries map[string][]string, action Action) {
	for ip, hostnames := range entries {
		fmt.Println(action, ip, hostnames)
		switch action {
		case START:
			entry := hosts.NewEntry(ip, hostnames, "docker")
			l.onStart <- *entry
		case KILL:
			l.onKill <- ip
		default:
			fmt.Println("Error")
		}
	}
}

func (l *DockerListener) Init(action Action) error {
	opts := dockertypes.ContainerListOptions{}
	containers, err := l.cli.ContainerList(l.ctx, opts)
	if err != nil {
		return err
	}
	for _, container := range containers {
		entries, err := l.entries(container.ID)
		if err != nil {
			return err
		}
		l.notify(entries, action)
	}
	return nil
}

func (l *DockerListener) Listen() error {
	f := filters.NewArgs()
	f.Add("type", "container")
	opts := dockertypes.EventsOptions{
		Filters: f,
	}
	eventsch, errch := l.cli.Events(l.ctx, opts)
	for {
		select {
		case event := <-eventsch:
			action := Action(event.Action)
			if action == START || action == KILL {
				entries, err := l.entries(event.ID)
				if err != nil {
					fmt.Println(err)
				}
				l.notify(entries, action)
			}
		case err := <-errch:
			fmt.Print(err.Error())
		case <-l.ctx.Done():
			return nil
		}

	}
}
