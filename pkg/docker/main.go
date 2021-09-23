package docker

import (
	"context"
	"fmt"
	"strings"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type DockerListener struct {
}

func (p *DockerListener) Init() error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	f := filters.NewArgs()
	f.Add("type", "container")
	opts := dockertypes.EventsOptions{
		Filters: f,
	}

	eventsch, errch := cli.Events(ctx, opts)
	for {
		select {
		case event := <-eventsch:
			if event.Action == "start" || event.Action == "kill" {
				json, err := cli.ContainerInspect(ctx, event.ID)
				if err != nil {
					fmt.Print("error")
				}

				fmt.Println(json.NetworkSettings.IPAddress, json.Config.Hostname, "#", event.Action)
				for _, el := range json.NetworkSettings.Networks {
					if len(el.Aliases) > 0 {
						fmt.Println(el.IPAddress, strings.Join(el.Aliases, " "), "#", event.Action)
					}
				}

			}
		case err := <-errch:
			fmt.Print(err.Error())
		case <-ctx.Done():
			return nil
		}

	}
}
