package main

import (
	"context"
	"fmt"
	"testing"

	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
	"github.com/vTCP-Foundation/vtcpd-test-suite/tests/testconfig"
)

var (
	// directExchangeSeveralNodesNextNodeIndex is used to assign unique IP addresses to nodes across different test functions in this file.
	directExchangeSeveralNodesNextNodeIndex = 1
)

func getNextIPForDirectExchangeSeveralNodesTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectExchangeSeveralNodes, directExchangeSeveralNodesNextNodeIndex)
	directExchangeSeveralNodesNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForDirectExchangeSeveralNodesTest(t *testing.T) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, 8)
	for i := range 8 {
		nodes[i] = vtcp.NewNode(t, getNextIPForDirectExchangeSeveralNodesTest(), fmt.Sprintf("node%d", i+1))
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
	nodes[5].OpenChannelAndCheck(t, nodes[0])
	nodes[6].OpenChannelAndCheck(t, nodes[5])
	nodes[7].OpenChannelAndCheck(t, nodes[6])
	nodes[4].OpenChannelAndCheck(t, nodes[7])

	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.ExchangeEquivalent, "3000")
	nodes[2].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.ExchangeEquivalent, "2500")
	nodes[5].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.ExchangeEquivalent, "3000")
	nodes[6].CreateAndSetSettlementLineAndCheck(t, nodes[5], testconfig.ExchangeEquivalent, "2500")

	nodes[3].CreateAndSetSettlementLineAndCheck(t, nodes[2], testconfig.Equivalent, "200")
	nodes[4].CreateAndSetSettlementLineAndCheck(t, nodes[3], testconfig.Equivalent, "500")
	nodes[7].CreateAndSetSettlementLineAndCheck(t, nodes[6], testconfig.Equivalent, "200")
	nodes[4].CreateAndSetSettlementLineAndCheck(t, nodes[7], testconfig.Equivalent, "500")

	// nodes[1].CheckMaxFlow(t, nodes[4], testconfig.Equivalent, "0")
	// nodes[1].CheckMaxFlow(t, nodes[4], testconfig.ExchangeEquivalent, "0")

	return nodes, cluster
}

func Test1ExchangeSeveralNodes(t *testing.T) {
	nodes, _ := setupNodesForDirectExchangeSeveralNodesTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)
	nodes[6].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "400")
}

func Test2ExchangeSeveralNodesWithCommissions(t *testing.T) {
	nodes, _ := setupNodesForDirectExchangeSeveralNodesTest(t)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[7].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 3},
	})

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)
	nodes[6].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "397")
}
