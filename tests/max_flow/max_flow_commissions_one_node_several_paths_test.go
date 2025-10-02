package main

import (
	"context"
	"fmt"
	"testing"

	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
	"github.com/vTCP-Foundation/vtcpd-test-suite/tests/testconfig"
)

var (
	// directCommissionsOneNodeSeveralPathsNextNodeIndex is used to assign unique IP addresses to nodes across different test functions in this file.
	directCommissionsOneNodeSeveralPathsNextNodeIndex = 1
)

func getNextIPForDirectCommissionsOneNodeSeveralPathsTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectCommissionsOneNodeSeveralPaths, directCommissionsOneNodeSeveralPathsNextNodeIndex)
	directCommissionsOneNodeSeveralPathsNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForDirectCommissionsOneNodeSeveralPathsTest(t *testing.T) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, 6)
	for i := range 6 {
		nodes[i] = vtcp.NewNode(t, getNextIPForDirectCommissionsOneNodeSeveralPathsTest(), fmt.Sprintf("node%d", i+1))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings) // Assuming clusterSettings is defined globally or passed
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes, false)

	nodes[1].OpenChannelAndCheck(t, nodes[0])
	nodes[5].OpenChannelAndCheck(t, nodes[2])
	nodes[4].OpenChannelAndCheck(t, nodes[3])
	nodes[2].OpenChannelAndCheck(t, nodes[1])
	nodes[5].OpenChannelAndCheck(t, nodes[4])
	nodes[5].OpenChannelAndCheck(t, nodes[1])
	nodes[3].OpenChannelAndCheck(t, nodes[1])

	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "1000")
	nodes[5].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "500")
	nodes[2].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "800")
	nodes[5].CreateAndSetSettlementLineAndCheck(t, nodes[2], testconfig.Equivalent, "200")
	nodes[3].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "700")
	nodes[4].CreateAndSetSettlementLineAndCheck(t, nodes[3], testconfig.Equivalent, "300")
	nodes[5].CreateAndSetSettlementLineAndCheck(t, nodes[4], testconfig.Equivalent, "900")

	nodes[0].CheckMaxFlow(t, nodes[5], testconfig.Equivalent, "1000")

	return nodes, cluster
}

func Test1DirectCommissionsOneNodeSeveralPaths(t *testing.T) {
	nodes, _ := setupNodesForDirectCommissionsOneNodeSeveralPathsTest(t)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 10},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.Equivalent}, "990")
}

func Test2DirectCommissionsOneNodeSeveralPaths(t *testing.T) {
	nodes, _ := setupNodesForDirectCommissionsOneNodeSeveralPathsTest(t)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 10},
	})
	nodes[2].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 20},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.Equivalent}, "970")
}
