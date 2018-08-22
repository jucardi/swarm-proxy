package docker

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
)

type IDockerClient interface {
	GetContainers() ([]types.Container, error)
	GetServices() ([]swarm.Service, error)
	GetNodes() ([]swarm.Node, error)
}
