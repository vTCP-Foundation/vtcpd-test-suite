package tests

import (
	"context"
	"sync"
	"testing"

	// Assuming your go.mod module path is 'github.com/vTCP-Foundation/vtcpd-test-suite'
	// Adjust this path if your module name is different.
	"github.com/vTCP-Foundation/vtcpd-test-suite/internal/conf"
	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
)

var (
	// Get settings from the central configuration package
	testCfg         = conf.GetConfig()
	clusterSettings = &vtcp.ClusterSettings{
		NodeImageName: testCfg.NodeImageName,
		NetworkName:   testCfg.NetworkName,
	}
)

func TestTrustLineSet(t *testing.T) {
	nodeA := vtcp.NewNode(t)
	nodeB := vtcp.NewNode(t)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, clusterSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	wg := sync.WaitGroup{}
	{
		err = cluster.RunNode(ctx, t, &wg, nodeA)
		if err != nil {
			t.Fatalf("failed to run nodeA: %v", err)
		}

		err = cluster.RunNode(ctx, t, &wg, nodeB)
		if err != nil {
			t.Fatalf("failed to run nodeB: %v", err)
		}
	}
	wg.Wait()

	println(nodeA.IPAddress)
	println(nodeB.IPAddress)

	println(nodeA.ContainerID)
	println(nodeB.ContainerID)
}
