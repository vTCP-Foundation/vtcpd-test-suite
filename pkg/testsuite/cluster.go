package testsuite

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type ClusterSettings struct {
	NodeImageName string
	NetworkName   string
}

type Cluster struct {
	cli       *client.Client
	ctx       context.Context
	networkID string
	nodes     []*Node
	settings  *ClusterSettings
}

func NewCluster(ctx context.Context, t *testing.T, settings *ClusterSettings) (*Cluster, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %v", err)
	}

	cluster := &Cluster{
		cli:      cli,
		ctx:      ctx,
		settings: settings,
	}
	networkID, err := cluster.initNetwork()
	if err != nil {
		return nil, fmt.Errorf("failed to create network: %v", err)
	}

	t.Cleanup(func() {
		if err := cluster.dropNetwork(); err != nil {
			t.Logf("failed to remove network: %v", err)
		}
	})

	cluster.networkID = networkID
	return cluster, nil
}

func (c *Cluster) RunNode(ctx context.Context, t *testing.T, wg *sync.WaitGroup, node *Node) (err error) {
	// Create container
	resp, err := c.cli.ContainerCreate(c.ctx,
		&container.Config{
			Image: c.settings.NodeImageName,
			ExposedPorts: nat.PortSet{
				nat.Port(strconv.Itoa(int(node.Port))): struct{}{},
			},
			Env: node.Env,
		},
		&container.HostConfig{
			NetworkMode: container.NetworkMode(c.networkID),
			PortBindings: nat.PortMap{
				nat.Port(strconv.Itoa(int(node.Port))): []nat.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: "0", // Let Docker assign a random port
					},
				},
			},
		},
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				c.settings.NetworkName: {
					NetworkID: c.networkID,
				},
			},
		},
		nil,
		"",
	)
	if err != nil {
		return fmt.Errorf("failed to create container: %v", err)
	}

	// Start container
	if err := c.cli.ContainerStart(c.ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}

	// Get container info
	inspect, err := c.cli.ContainerInspect(c.ctx, resp.ID)
	if err != nil {
		return fmt.Errorf("failed to inspect container: %v", err)
	}

	ipAddress := inspect.NetworkSettings.Networks[c.settings.NetworkName].IPAddress
	node.IPAddress = ipAddress
	node.ContainerID = resp.ID

	// Automatically stop and remove container when test finishes.
	// Helps prevent boilerplate code in tests.
	t.Cleanup(func() {
		secondsToWait := 5
		if err := c.cli.ContainerStop(c.ctx, resp.ID, container.StopOptions{Timeout: &secondsToWait}); err != nil {
			t.Logf("failed to stop container: %v", err)
		}
		if err := c.cli.ContainerRemove(c.ctx, resp.ID, container.RemoveOptions{}); err != nil {
			t.Logf("failed to remove container: %v", err)
		}
	})

	return nil
}

func (c *Cluster) initNetwork() (string, error) {
	// Check if network already exists
	networks, err := c.cli.NetworkList(c.ctx, network.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to list networks: %v", err)
	}

	for _, nw := range networks {
		if nw.Name == c.settings.NetworkName {
			return nw.ID, nil
		}
	}

	// Create network if it doesn't exist
	resp, err := c.cli.NetworkCreate(c.ctx, c.settings.NetworkName, network.CreateOptions{
		Driver: "bridge",
	})
	if err != nil {
		return "", fmt.Errorf("failed to create network: %v", err)
	}
	return resp.ID, nil
}

func (c *Cluster) dropNetwork() error {
	return c.cli.NetworkRemove(c.ctx, c.networkID)
}
