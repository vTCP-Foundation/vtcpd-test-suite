package main

import (
	"context"
	"fmt"
	"testing"

	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
	"github.com/vTCP-Foundation/vtcpd-test-suite/tests/testconfig"
)

var (
	// directCommissionsFiveNodesNextNodeIndex is used to assign unique IP addresses to nodes across different test functions in this file.
	directCommissionsFiveNodesNextNodeIndex = 1
)

func getNextIPForDirectCommissionsFiveNodesTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectCommissionsFiveNodes, directCommissionsFiveNodesNextNodeIndex)
	directCommissionsFiveNodesNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForDirectCommissionsFiveNodesTest(t *testing.T) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, 5)
	for i := range 5 {
		nodes[i] = vtcp.NewNode(t, getNextIPForDirectCommissionsFiveNodesTest(), fmt.Sprintf("node%d", i+1))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes, false)

	// Setup topology according to Python version
	nodes[1].OpenChannelAndCheck(t, nodes[0])
	nodes[2].OpenChannelAndCheck(t, nodes[1])
	nodes[3].OpenChannelAndCheck(t, nodes[2])
	nodes[4].OpenChannelAndCheck(t, nodes[3])

	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "3000")
	nodes[2].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "2500")
	nodes[3].CreateAndSetSettlementLineAndCheck(t, nodes[2], testconfig.Equivalent, "2000")
	nodes[4].CreateAndSetSettlementLineAndCheck(t, nodes[3], testconfig.Equivalent, "5000")

	nodes[0].CheckMaxFlow(t, nodes[4], testconfig.Equivalent, "2000")

	return nodes, cluster
}

func Test1DirectCommissions5Commissions3thNode(t *testing.T) {
	nodes, _ := setupNodesForDirectCommissionsFiveNodesTest(t)

	nodes[2].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 10},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.Equivalent}, "2000")
}

func Test1DirectCommissions5Commissions4thNode(t *testing.T) {
	nodes, _ := setupNodesForDirectCommissionsFiveNodesTest(t)

	nodes[3].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 10},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.Equivalent}, "1990")
}

func Test1DirectCommissions5Commissions3thAnd4thNodes(t *testing.T) {
	nodes, _ := setupNodesForDirectCommissionsFiveNodesTest(t)

	nodes[2].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 10},
	})
	nodes[3].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 20},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.Equivalent}, "1980")
}
