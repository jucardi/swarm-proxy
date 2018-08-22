package docker

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types/swarm"
	"github.com/jucardi/go-beans/beans"
)

const DefaultClient = "docker-client"

var (
	// To validate the interface implementation at compile time instead of runtime.
	_ IDockerClient = (*dockerClient)(nil)

	instance *dockerClient
)

type dockerClient struct {
	cli *client.Client
}

func Client() IDockerClient {
	return beans.Resolve((*IDockerClient)(nil), DefaultClient).(IDockerClient)
}

// Registering the bean implementation.
func init() {
	beans.RegisterFunc((*IDockerClient)(nil), DefaultClient, func() interface{} {
		if instance != nil {
			return instance
		}

		instance = &dockerClient{}
		instance.init()
		return instance
	})
}

func (d *dockerClient) init() {
	clix, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	d.cli = clix
}

func (d *dockerClient) GetContainers() ([]types.Container, error) {
	return d.cli.ContainerList(context.Background(), types.ContainerListOptions{})
}

func (d *dockerClient) GetServices() ([]swarm.Service, error) {
	return d.cli.ServiceList(context.Background(), types.ServiceListOptions{})
}

func (d *dockerClient) GetNodes() ([]swarm.Node, error) {
	return d.cli.NodeList(context.Background(), types.NodeListOptions{})
}
