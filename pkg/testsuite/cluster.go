package testsuite

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

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

	networkID, err := cluster.initNetwork(t)
	if err != nil {
		// The error from initNetwork should already include the specific network name.
		// Wrap the error to provide context from NewCluster.
		return nil, fmt.Errorf("failed to create cluster using network name '%s': %w", cluster.settings.NetworkName, err)
	}

	t.Cleanup(func() {
		// // Capture networkName and networkID at the time of cleanup registration
		// nameForCleanup := cluster.settings.NetworkName
		// idForCleanup := cluster.networkID
		// if err := cluster.dropNetwork(); err != nil {
		// 	t.Logf("Failed to remove network '%s' (ID: %s): %v", nameForCleanup, idForCleanup, err)
		// } else {
		// 	t.Logf("Successfully removed network '%s' (ID: %s)", nameForCleanup, idForCleanup)
		// }
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
				nat.Port(strconv.Itoa(int(node.NodePort))):    struct{}{},
				nat.Port(strconv.Itoa(int(node.CLIPort))):     struct{}{},
				nat.Port(strconv.Itoa(int(node.CLIPortTest))): struct{}{},
			},
			Env: node.Env,
		},
		&container.HostConfig{
			NetworkMode: container.NetworkMode(c.networkID),
			PortBindings: nat.PortMap{
				nat.Port(strconv.Itoa(int(node.NodePort))): []nat.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: "0", // Let Docker assign a random port
					},
				},
				nat.Port(strconv.Itoa(int(node.CLIPort))): []nat.PortBinding{
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
					IPAMConfig: &network.EndpointIPAMConfig{
						IPv4Address: node.IPAddress,
					},
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

func (c *Cluster) RunNodes(ctx context.Context, t *testing.T, nodes []*Node) {
	wg := sync.WaitGroup{}
	{
		for _, node := range nodes {
			err := c.RunNode(ctx, t, &wg, node)
			if err != nil {
				t.Fatalf("failed to run %s: %v", node.Alias, err)
			}
		}
	}
	wg.Wait()

	for _, node := range nodes {
		println(fmt.Sprintf("Node %s is running : [%s : %s]", node.Alias, node.IPAddress, node.ContainerID))
	}

	time.Sleep(5 * time.Second)
}

func (c *Cluster) RunSingleNode(ctx context.Context, t *testing.T, node *Node) {
	// Use a dummy WaitGroup as RunNode expects one, though for a single node it's not strictly necessary for synchronization here.
	var wg sync.WaitGroup
	wg.Add(1) // Add to the counter before starting the goroutine/operation
	go func() {
		defer wg.Done()                     // Decrement counter when goroutine finishes
		err := c.RunNode(ctx, t, &wg, node) // Pass wg, though RunNode doesn't use it to wait directly
		if err != nil {
			t.Fatalf("failed to run %s: %v", node.Alias, err)
		}
	}()
	wg.Wait() // Wait for the single node to be processed by RunNode

	t.Logf("Node %s is running : [%s : %s]", node.Alias, node.IPAddress, node.ContainerID)
	time.Sleep(2 * time.Second) // Give some time for the node to fully initialize
}

func (c *Cluster) StopSingleNode(ctx context.Context, t *testing.T, node *Node) {
	if node.ContainerID == "" {
		t.Logf("Node %s has no container ID, skipping stop.", node.Alias)
		return
	}
	secondsToWait := 5
	if err := c.cli.ContainerStop(c.ctx, node.ContainerID, container.StopOptions{Timeout: &secondsToWait}); err != nil {
		t.Logf("failed to stop container for node %s (ID: %s): %v", node.Alias, node.ContainerID, err)
	}
	// Note: ContainerRemove is handled by t.Cleanup in RunNode
	t.Logf("Stopped container for node %s (ID: %s)", node.Alias, node.ContainerID)
}

func (c *Cluster) StopNodes(ctx context.Context, t *testing.T, nodes []*Node) {
	for _, node := range nodes {
		c.StopSingleNode(ctx, t, node)
	}
}

func (c *Cluster) initNetwork(t *testing.T) (string, error) {
	// Try to inspect the network by name to see if it exists.
	networkResource, inspectErr := c.cli.NetworkInspect(c.ctx, c.settings.NetworkName, network.InspectOptions{})
	if inspectErr == nil {
		t.Logf("Network %s already exists.", c.settings.NetworkName)
		return networkResource.ID, nil
	} else if !client.IsErrNotFound(inspectErr) {
		// Inspect failed for a reason other than "not found", which is an issue.
		return "", fmt.Errorf("failed to inspect network %s prior to creation: %v", c.settings.NetworkName, inspectErr)
	}
	// If inspectErr was client.IsErrNotFound(inspectErr), network doesn't exist, which is good. We proceed to create.

	// Now, attempt to create the network.
	resp, createErr := c.cli.NetworkCreate(c.ctx, c.settings.NetworkName, network.CreateOptions{
		Driver: "bridge",
		IPAM: &network.IPAM{
			Driver: "default",
			Config: []network.IPAMConfig{
				{
					Subnet: "172.18.0.0/16",
				},
			},
		},
	})
	if createErr != nil {
		// Provide context if we had issues during inspection/pre-emptive removal attempt.
		if inspectErr == nil {
			return "", fmt.Errorf("failed to create network %s (it existed, removal was attempted, but creation still failed): %v", c.settings.NetworkName, createErr)
		}
		return "", fmt.Errorf("failed to create network %s: %v", c.settings.NetworkName, createErr)
	}
	return resp.ID, nil
}

func (c *Cluster) dropNetwork() error {
	return c.cli.NetworkRemove(c.ctx, c.networkID)
}
