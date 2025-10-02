package main

import (
	"context"
	"fmt"
	"testing"

	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
	"github.com/vTCP-Foundation/vtcpd-test-suite/tests/testconfig"
)

var (
	// directExchangeFiveNodesNextNodeIndex is used to assign unique IP addresses to nodes across different test functions in this file.
	directExchangeFiveNodesNextNodeIndex = 1
)

func getNextIPForDirectExchangeFiveNodesTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectExchangeFiveNodes, directExchangeFiveNodesNextNodeIndex)
	directExchangeFiveNodesNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForDirectExchangeFiveNodesTest(t *testing.T) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, 5)
	for i := range 5 {
		nodes[i] = vtcp.NewNode(t, getNextIPForDirectExchangeFiveNodesTest(), fmt.Sprintf("node%d", i+1))
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

	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.ExchangeEquivalent, "3000")
	nodes[2].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.ExchangeEquivalent, "2500")
	nodes[3].CreateAndSetSettlementLineAndCheck(t, nodes[2], testconfig.ExchangeEquivalent, "2000")
	nodes[4].CreateAndSetSettlementLineAndCheck(t, nodes[3], testconfig.ExchangeEquivalent, "5000")

	//nodes[0].CheckMaxFlow(t, nodes[4], testconfig.ExchangeEquivalent, "2000")

	nodes[2].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "250")
	nodes[3].CreateAndSetSettlementLineAndCheck(t, nodes[2], testconfig.Equivalent, "200")
	nodes[4].CreateAndSetSettlementLineAndCheck(t, nodes[3], testconfig.Equivalent, "500")

	//nodes[1].CheckMaxFlow(t, nodes[4], testconfig.Equivalent, "2000")

	return nodes, cluster
}

// Helper to create and run nodes for a test
func setupNodesForDirectExchangeFiveNodesTestWithCommissions(t *testing.T) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, 5)
	for i := range 5 {
		nodes[i] = vtcp.NewNode(t, getNextIPForDirectExchangeFiveNodesTest(), fmt.Sprintf("node%d", i+1))
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

	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.ExchangeEquivalent, "300")
	nodes[2].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.ExchangeEquivalent, "200")
	nodes[3].CreateAndSetSettlementLineAndCheck(t, nodes[2], testconfig.ExchangeEquivalent, "250")

	//nodes[0].CheckMaxFlow(t, nodes[4], testconfig.ExchangeEquivalent, "2000")

	nodes[3].CreateAndSetSettlementLineAndCheck(t, nodes[2], testconfig.Equivalent, "250")
	nodes[4].CreateAndSetSettlementLineAndCheck(t, nodes[3], testconfig.Equivalent, "500")

	//nodes[1].CheckMaxFlow(t, nodes[4], testconfig.Equivalent, "2000")

	return nodes, cluster
}

func Test1DirectExchange5Exchanger3thNode(t *testing.T) {
	nodes, _ := setupNodesForDirectExchangeFiveNodesTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -2, nil, nil, vtcp.StatusOK)

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "125")
}

func Test2DirectExchange5Exchanger3thAnd4thNodes(t *testing.T) {
	nodes, _ := setupNodesForDirectExchangeFiveNodesTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -2, nil, nil, vtcp.StatusOK)
	nodes[3].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -2, nil, nil, vtcp.StatusOK)

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "125")
}

func Test3DirectExchange5NodesWithCommissions(t *testing.T) {
	nodes, _ := setupNodesForDirectExchangeFiveNodesTest(t)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[3].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 3},
	})

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "197")
}

func Test4DirectExchange5NodesWithCommissions(t *testing.T) {
	nodes, _ := setupNodesForDirectExchangeFiveNodesTest(t)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 20},
	})

	nodes[3].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 3},
	})

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -2, nil, nil, vtcp.StatusOK)

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "122")
}

func Test5DirectExchange5Exchanger3thNodeWithCommission(t *testing.T) {
	nodes, _ := setupNodesForDirectExchangeFiveNodesTest(t)

	nodes[2].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 3},
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -2, nil, nil, vtcp.StatusOK)

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "125")
}

func Test6DirectExchange5Exchanger3thNodeWithCommission(t *testing.T) {
	nodes, _ := setupNodesForDirectExchangeFiveNodesTestWithCommissions(t)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[3].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 20},
	})

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "1", 0, nil, nil, vtcp.StatusOK)

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "180")
}

func Test7DirectExchange5Exchanger3thNodeWithCommission(t *testing.T) {
	nodes, _ := setupNodesForDirectExchangeFiveNodesTestWithCommissions(t)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[2].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 20},
	})

	nodes[3].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "1", 0, nil, nil, vtcp.StatusOK)

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "180")
}
