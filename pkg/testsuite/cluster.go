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
	"github.com/google/uuid"
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

	// Create a unique network name for this cluster instance
	// Use first 8 chars of a UUID for a short, likely unique suffix.
	idSuffix := uuid.NewString()[:8]
	uniqueNetworkName := settings.NetworkName + "-" + idSuffix

	// Use a copy of the settings for this cluster instance with the unique network name.
	// This is important so that the original settings object isn't modified if it's shared,
	// and so that c.settings.NetworkName inside methods refers to the unique one.
	currentClusterSettings := *settings // Copy struct
	currentClusterSettings.NetworkName = uniqueNetworkName
	t.Logf("Using unique network name for this test run: %s", uniqueNetworkName)

	cluster := &Cluster{
		cli:      cli,
		ctx:      ctx,
		settings: &currentClusterSettings, // Store pointer to the struct with unique name
	}

	networkID, err := cluster.initNetwork(t)
	if err != nil {
		// The error from initNetwork should already include the specific network name.
		// Wrap the error to provide context from NewCluster.
		return nil, fmt.Errorf("failed to create cluster using network name '%s': %w", cluster.settings.NetworkName, err)
	}

	t.Cleanup(func() {
		// Capture networkName and networkID at the time of cleanup registration
		nameForCleanup := cluster.settings.NetworkName
		idForCleanup := cluster.networkID
		if err := cluster.dropNetwork(); err != nil {
			t.Logf("Failed to remove network '%s' (ID: %s): %v", nameForCleanup, idForCleanup, err)
		} else {
			t.Logf("Successfully removed network '%s' (ID: %s)", nameForCleanup, idForCleanup)
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

	time.Sleep(2 * time.Second)
}

func (c *Cluster) initNetwork(t *testing.T) (string, error) {
	// Try to inspect the network by name to see if it exists.
	networkResource, inspectErr := c.cli.NetworkInspect(c.ctx, c.settings.NetworkName, network.InspectOptions{})
	if inspectErr == nil {
		// Network exists. Attempt to remove it to ensure a clean state.
		if removeErr := c.cli.NetworkRemove(c.ctx, networkResource.ID); removeErr != nil {
			t.Logf("Warning: Pre-emptive removal of existing network %s (ID: %s) failed: %v. This might be due to attached containers or other issues.", c.settings.NetworkName, networkResource.ID, removeErr)
			// We'll still try to create it; if it fails, that error will be more definitive.
		}
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
