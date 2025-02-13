package tests

import (
	"context"
	"sync"
	"testing"

	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
)

var (
	clusterSettings = &vtcp.ClusterSettings{
		NodeImageName: "vtcpd-test:manjaro",
		NetworkName:   "vtcpd-test-network",
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
