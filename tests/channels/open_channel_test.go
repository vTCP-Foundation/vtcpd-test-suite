package main

import (
	"context"
	"fmt"
	"testing"

	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
	"github.com/vTCP-Foundation/vtcpd-test-suite/tests/testconfig"
)

var (
	openChannelNextNodeIndex = 1
)

func getNextIPForOpenChannelTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForOpenChannelTest, openChannelNextNodeIndex)
	openChannelNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForOpenChannelTest(t *testing.T, count int) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, count)
	for i := range count {
		nodes[i] = vtcp.NewNode(t, getNextIPForOpenChannelTest(), fmt.Sprintf("node%c", 'A'+i))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings) // Assuming clusterSettings is defined globally or passed
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes, false)
	return nodes, cluster
}

func TestOpenChannelNormalPass(t *testing.T) {
	nodes, _ := setupNodesForOpenChannelTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	// TODO: Check DB
}

func TestOpenChannelExistingChannel(t *testing.T) {
	nodes, _ := setupNodesForOpenChannelTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)
	channelInfo := nodeA.GetChannelInfo(t, "0")

	nodeA.InitChannelCheckStatusCode(t, nodeB, "", "", vtcp.StatusAlreadyExists)
	nodeA.InitChannelCheckStatusCode(t, nodeB, channelInfo.ChannelCryptoKey, channelInfo.ChannelID, vtcp.StatusAlreadyExists)

	nodeB.InitChannelCheckStatusCode(t, nodeA, "", "", vtcp.StatusAlreadyExists)
	nodeB.InitChannelCheckStatusCode(t, nodeA, channelInfo.ChannelCryptoKey, channelInfo.ChannelID, vtcp.StatusAlreadyExists)
}
