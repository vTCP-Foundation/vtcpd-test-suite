package testsuite

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
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
	SudoPassword  string
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

	t.Cleanup(func() {})

	cluster.networkID = networkID
	return cluster, nil
}

func (c *Cluster) RunNode(ctx context.Context, t *testing.T, wg *sync.WaitGroup, node *Node) (err error) {
	// Get VTCPD_DATABASE_CONFIG from environment and add it to node.Env if it exists
	envVars := node.Env
	if dbConfig := os.Getenv("VTCPD_DATABASE_CONFIG"); dbConfig != "" {
		envVars = append(envVars, fmt.Sprintf("VTCPD_DATABASE_CONFIG=%s", dbConfig))
	}

	// Create container
	resp, err := c.cli.ContainerCreate(c.ctx,
		&container.Config{
			Image: c.settings.NodeImageName,
			ExposedPorts: nat.PortSet{
				nat.Port(strconv.Itoa(int(node.NodePort))):    struct{}{},
				nat.Port(strconv.Itoa(int(node.CLIPort))):     struct{}{},
				nat.Port(strconv.Itoa(int(node.CLIPortTest))): struct{}{},
			},
			Env: envVars,
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

	// Wait for all nodes to be ready with health checks
	// Use longer timeout for PostgreSQL which takes more time to initialize
	timeout := 60 * time.Second

	for _, node := range nodes {
		if err := node.WaitForReady(t, timeout); err != nil {
			t.Fatalf("Node %s failed to become ready: %v", node.Alias, err)
		}
	}
}

func (c *Cluster) RunSingleNode(ctx context.Context, t *testing.T, node *Node) {
	// No need for goroutine when running a single node
	var wg sync.WaitGroup // Dummy WaitGroup as RunNode expects one
	err := c.RunNode(ctx, t, &wg, node)
	if err != nil {
		t.Fatalf("failed to run %s: %v", node.Alias, err)
	}

	t.Logf("Node %s is running : [%s : %s]", node.Alias, node.IPAddress, node.ContainerID)

	// Wait for node to be ready with health check
	timeout := 60 * time.Second
	if err := node.WaitForReady(t, timeout); err != nil {
		t.Fatalf("Node %s failed to become ready: %v", node.Alias, err)
	}
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
		return "", fmt.Errorf("failed to create network %s: %v", c.settings.NetworkName, createErr)
	}
	return resp.ID, nil
}

// NetworkConditions defines network simulation parameters
type NetworkConditions struct {
	// Bandwidth limit (e.g., "1mbit", "100kbit", "10mbit", "1gbit"). Empty means no limit.
	Bandwidth string
	// Delay in milliseconds (e.g., 100 for 100ms delay). 0 means no delay.
	DelayMs int
	// Jitter in milliseconds - random variation in delay (e.g., 10 for ±10ms). 0 means no jitter.
	JitterMs int
	// Packet loss percentage (e.g., 10.0 for 10%). 0 means no loss.
	LossPercent float64
	// Packet duplication percentage (e.g., 1.0 for 1%). 0 means no duplication.
	DuplicatePercent float64
	// Packet corruption percentage (e.g., 0.1 for 0.1%). 0 means no corruption.
	CorruptPercent float64
	// Packet reordering percentage (e.g., 25 for 25%). 0 means no reordering.
	ReorderPercent float64
	// Gap for reordering - how many packets to delay for reordering (default: 5).
	ReorderGap int
}

// executeSudoCommand executes a command with sudo, using password if configured
func (c *Cluster) executeSudoCommand(args []string) error {
	var cmd *exec.Cmd
	var stdin bytes.Buffer

	if c.settings.SudoPassword != "" {
		// Use sudo -S to read password from stdin
		sudoArgs := append([]string{"-S"}, args...)
		cmd = exec.Command("sudo", sudoArgs...)
		stdin.WriteString(c.settings.SudoPassword + "\n")
		cmd.Stdin = &stdin
	} else {
		// Use sudo without password (will prompt interactively if needed)
		cmd = exec.Command("sudo", args...)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sudo command failed: %v. Command: 'sudo %s'. Stderr: %s",
			err, strings.Join(args, " "), stderr.String())
	}

	return nil
}

// ConfigureNetworkConditions configures comprehensive network conditions for a given node's container.
// It uses 'tc' and 'netem'/'tbf' on the host system and requires sudo privileges.
func (c *Cluster) ConfigureNetworkConditions(node *Node, conditions *NetworkConditions, containerInterfaceName string) error {
	if node.ContainerID == "" {
		return fmt.Errorf("node %s has no container ID, cannot configure network conditions", node.Alias)
	}
	if containerInterfaceName == "" {
		containerInterfaceName = "eth0" // Default to eth0
	}

	// 1. Get ifindex of the interface inside the container
	dockerCmd := exec.Command("docker", "exec", node.ContainerID, "cat", fmt.Sprintf("/sys/class/net/%s/ifindex", containerInterfaceName))
	cmdOutput, err := dockerCmd.Output()
	if err != nil {
		errMsg := fmt.Sprintf("failed to get ifindex for %s in container %s (%s)", containerInterfaceName, node.Alias, node.ContainerID)
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("%s: %v, stderr: %s", errMsg, err, string(exitErr.Stderr))
		}
		return fmt.Errorf("%s: %v", errMsg, err)
	}
	containerIfindexStr := strings.TrimSpace(string(cmdOutput))

	// 2. Find host veth interface linked to the container's interface index
	ipCmd := exec.Command("ip", "-o", "link")
	cmdOutput, err = ipCmd.Output()
	if err != nil {
		errMsg := "failed to list host interfaces using 'ip -o link'"
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("%s: %v, stderr: %s", errMsg, err, string(exitErr.Stderr))
		}
		return fmt.Errorf("%s: %v", errMsg, err)
	}

	hostVethInterface := ""
	re := regexp.MustCompile(`^\d+:\s+([^@\s]+)@if` + regexp.QuoteMeta(containerIfindexStr) + `\b`)
	lines := strings.Split(string(cmdOutput), "\n")
	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) > 1 {
			hostVethInterface = strings.TrimSpace(matches[1])
			break
		}
	}

	if hostVethInterface == "" {
		return fmt.Errorf("could not find host veth interface for container %s (alias %s) with internal ifindex %s on interface %s",
			node.ContainerID, node.Alias, containerIfindexStr, containerInterfaceName)
	}

	// 3. Clear any existing qdisc first
	clearArgs := []string{"tc", "qdisc", "del", "dev", hostVethInterface, "root"}
	// Ignore errors as there might not be any existing qdisc - this is normal
	c.executeSudoCommand(clearArgs)

	// 4. Configure bandwidth limitation if specified
	if conditions.Bandwidth != "" {
		// Use TBF (Token Bucket Filter) for bandwidth limiting
		// Default burst and latency values that work well for most cases
		burst := "32kbit"
		latency := "400ms"

		tbfArgs := []string{
			"tc", "qdisc", "add", "dev", hostVethInterface, "root", "handle", "1:",
			"tbf", "rate", conditions.Bandwidth, "burst", burst, "latency", latency,
		}

		if err := c.executeSudoCommand(tbfArgs); err != nil {
			return fmt.Errorf("failed to configure bandwidth limit for node %s: %v", node.Alias, err)
		}

		// If we have bandwidth limiting, netem should be added as a child qdisc
		parentHandle := "1:1"
		netemHandle := "2:"

		// Add netem as child if we have any netem parameters
		if c.hasNetemParams(conditions) {
			netemArgs := []string{"tc", "qdisc", "add", "dev", hostVethInterface, "parent", parentHandle, "handle", netemHandle, "netem"}
			netemArgs = append(netemArgs, c.buildNetemParams(conditions)...)

			if err := c.executeSudoCommand(netemArgs); err != nil {
				return fmt.Errorf("failed to configure netem conditions for node %s: %v", node.Alias, err)
			}
		}
	} else if c.hasNetemParams(conditions) {
		// No bandwidth limiting, just use netem directly on root
		netemArgs := []string{"tc", "qdisc", "add", "dev", hostVethInterface, "root", "netem"}
		netemArgs = append(netemArgs, c.buildNetemParams(conditions)...)

		println(fmt.Sprintf("Executing netem command: %s", strings.Join(netemArgs, " ")))
		if err := c.executeSudoCommand(netemArgs); err != nil {
			return fmt.Errorf("failed to configure netem conditions for node %s: %v", node.Alias, err)
		}
	}

	time.Sleep(2 * time.Second) // Allow time for changes to take effect
	return nil
}

// hasNetemParams checks if any netem parameters are specified
func (c *Cluster) hasNetemParams(conditions *NetworkConditions) bool {
	return conditions.DelayMs > 0 || conditions.JitterMs > 0 || conditions.LossPercent > 0 ||
		conditions.DuplicatePercent > 0 || conditions.CorruptPercent > 0 || conditions.ReorderPercent > 0
}

// buildNetemParams builds the netem parameter list from NetworkConditions
func (c *Cluster) buildNetemParams(conditions *NetworkConditions) []string {
	var params []string

	// Add delay and jitter
	if conditions.DelayMs > 0 {
		if conditions.JitterMs > 0 {
			params = append(params, "delay", fmt.Sprintf("%dms", conditions.DelayMs), fmt.Sprintf("%dms", conditions.JitterMs))
		} else {
			params = append(params, "delay", fmt.Sprintf("%dms", conditions.DelayMs))
		}
	}

	// Add packet loss
	if conditions.LossPercent > 0 {
		params = append(params, "loss", fmt.Sprintf("%.2f%%", conditions.LossPercent))
	}

	// Add packet duplication
	if conditions.DuplicatePercent > 0 {
		params = append(params, "duplicate", fmt.Sprintf("%.2f%%", conditions.DuplicatePercent))
	}

	// Add packet corruption
	if conditions.CorruptPercent > 0 {
		params = append(params, "corrupt", fmt.Sprintf("%.2f%%", conditions.CorruptPercent))
	}

	// Add packet reordering
	if conditions.ReorderPercent > 0 {
		gap := conditions.ReorderGap
		if gap <= 0 {
			gap = 5 // Default gap
		}
		params = append(params, "reorder", fmt.Sprintf("%.0f%%", conditions.ReorderPercent), "gap", fmt.Sprintf("%d", gap))
	}

	return params
}

// RemoveNetworkConditions removes all network condition configurations for a node
func (c *Cluster) RemoveNetworkConditions(node *Node, containerInterfaceName string) error {
	if node.ContainerID == "" {
		return fmt.Errorf("node %s has no container ID, cannot remove network conditions", node.Alias)
	}
	if containerInterfaceName == "" {
		containerInterfaceName = "eth0"
	}

	// Get the host veth interface (reusing the same logic as in ConfigureNetworkConditions)
	dockerCmd := exec.Command("docker", "exec", node.ContainerID, "cat", fmt.Sprintf("/sys/class/net/%s/ifindex", containerInterfaceName))
	cmdOutput, err := dockerCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get ifindex for %s in container %s: %v", containerInterfaceName, node.Alias, err)
	}
	containerIfindexStr := strings.TrimSpace(string(cmdOutput))

	ipCmd := exec.Command("ip", "-o", "link")
	cmdOutput, err = ipCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list host interfaces: %v", err)
	}

	hostVethInterface := ""
	re := regexp.MustCompile(`^\d+:\s+([^@\s]+)@if` + regexp.QuoteMeta(containerIfindexStr) + `\b`)
	lines := strings.Split(string(cmdOutput), "\n")
	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) > 1 {
			hostVethInterface = strings.TrimSpace(matches[1])
			break
		}
	}

	if hostVethInterface == "" {
		return fmt.Errorf("could not find host veth interface for container %s", node.ContainerID)
	}

	// Remove all qdisc rules
	clearArgs := []string{"tc", "qdisc", "del", "dev", hostVethInterface, "root"}
	if err := c.executeSudoCommand(clearArgs); err != nil {
		// It's okay if this fails - there might not be any rules configured
		return nil
	}

	return nil
}
