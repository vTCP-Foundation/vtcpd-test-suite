package main

import (
	"context"
	"fmt"
	"testing"

	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
	"github.com/vTCP-Foundation/vtcpd-test-suite/tests/testconfig"
)

var (
	// directExchangeSeneNodesNextNodeIndex is used to assign unique IP addresses to nodes across different test functions in this file.
	directExchangeSevenNodesNextNodeIndex = 1
)

func getNextIPForDirectExchangeSevenNodesTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectExchangeSevenNodes, directExchangeSevenNodesNextNodeIndex)
	directExchangeSevenNodesNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForDirectExchangeSevenNodesTest(t *testing.T) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, 7)
	for i := range 7 {
		nodes[i] = vtcp.NewNode(t, getNextIPForDirectExchangeSevenNodesTest(), fmt.Sprintf("node%d", i+1))
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
	nodes[5].OpenChannelAndCheck(t, nodes[4])
	nodes[6].OpenChannelAndCheck(t, nodes[5])

	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.ExchangeEquivalent, "3000")
	nodes[2].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.ExchangeEquivalent, "2500")
	nodes[3].CreateAndSetSettlementLineAndCheck(t, nodes[2], testconfig.ExchangeEquivalent, "2000")
	nodes[4].CreateAndSetSettlementLineAndCheck(t, nodes[3], testconfig.ExchangeEquivalent, "5000")
	nodes[5].CreateAndSetSettlementLineAndCheck(t, nodes[4], testconfig.ExchangeEquivalent, "1000")
	nodes[6].CreateAndSetSettlementLineAndCheck(t, nodes[5], testconfig.ExchangeEquivalent, "1500")

	//nodes[0].CheckMaxFlow(t, nodes[4], testconfig.ExchangeEquivalent, "1000")

	nodes[2].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "250")
	nodes[3].CreateAndSetSettlementLineAndCheck(t, nodes[2], testconfig.Equivalent, "200")
	nodes[4].CreateAndSetSettlementLineAndCheck(t, nodes[3], testconfig.Equivalent, "500")
	nodes[5].CreateAndSetSettlementLineAndCheck(t, nodes[4], testconfig.Equivalent, "100")
	nodes[6].CreateAndSetSettlementLineAndCheck(t, nodes[5], testconfig.Equivalent, "150")

	//nodes[1].CheckMaxFlow(t, nodes[4], testconfig.Equivalent, "1000")

	return nodes, cluster
}

// this test is not working because third node does not send exchange rate to coordinator
// will be fixed in the future
func Test1DirectExchange7Exchanger4thNode(t *testing.T) {
	nodes, _ := setupNodesForDirectExchangeSevenNodesTest(t)
	nodes[3].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "2", 0, nil, nil, vtcp.StatusOK)

	nodes[0].CheckExchangeMaxFlow(t, nodes[6], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "100")
}

func Test1DirectExchange7Exchanger5thNode(t *testing.T) {
	nodes, _ := setupNodesForDirectExchangeSevenNodesTest(t)

	nodes[4].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "2", 0, nil, nil, vtcp.StatusOK)

	nodes[0].CheckExchangeMaxFlow(t, nodes[6], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "100")
}

func Test1DirectExchange7Exchanger3thNode(t *testing.T) {
	nodes, _ := setupNodesForDirectExchangeSevenNodesTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "2", 0, nil, nil, vtcp.StatusOK)

	nodes[0].CheckExchangeMaxFlow(t, nodes[6], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "100")
}

func Test1DirectExchange7Exchanger6thNode(t *testing.T) {
	nodes, _ := setupNodesForDirectExchangeSevenNodesTest(t)

	nodes[5].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "2", 0, nil, nil, vtcp.StatusOK)

	nodes[0].CheckExchangeMaxFlow(t, nodes[6], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "100")
}

func Test1DirectExchange7Exchanger1thNode(t *testing.T) {
	nodes, _ := setupNodesForDirectExchangeSevenNodesTest(t)

	nodes[1].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "2", 0, nil, nil, vtcp.StatusOK)

	nodes[0].CheckExchangeMaxFlow(t, nodes[6], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "100")
}
