package docker

import (
	"context"
	"fmt"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type Action string

type DockerListener struct {
	ctx     context.Context
	cli     *client.Client
	onStart chan<- []string
	onKill  chan<- string
}

const (
	START Action = "start"
	KILL  Action = "stop"
)

func NewDockerListeger(onStart chan<- []string, onKill chan<- string) *DockerListener {
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

	entries[json.NetworkSettings.IPAddress] = []string{json.Config.Hostname}

	for _, network := range json.NetworkSettings.Networks {
		if len(network.Aliases) > 0 {
			if val, ok := entries[network.IPAddress]; ok {
				entries[network.IPAddress] = append(val, network.Aliases...)
			} else {
				entries[network.IPAddress] = network.Aliases
			}
		}
	}

	return entries, nil
}

func (l *DockerListener) notify(entries map[string][]string, action Action) {
	for ip, hosts := range entries {
		fmt.Println("notify", ip, action)
		switch action {
		case START:
			l.onStart <- append([]string{ip}, hosts...)
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
