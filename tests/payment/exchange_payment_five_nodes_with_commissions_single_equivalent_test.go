package main

import (
	"context"
	"fmt"
	"testing"

	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
	"github.com/vTCP-Foundation/vtcpd-test-suite/tests/testconfig"
)

var (
	// exchangePaymentFiveNodesWithCommissionsSingleEquivalentNextNodeIndex is used to assign unique IP addresses to nodes across different test functions in this file.
	exchangePaymentFiveNodesWithCommissionsSingleEquivalentNextNodeIndex = 1
)

func getNextIPForExchangePaymentFiveNodesWithCommissionsSingleEquivalentTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForExchangePaymentFiveNodesWithCommissionsSingleEquivalent, exchangePaymentFiveNodesWithCommissionsSingleEquivalentNextNodeIndex)
	exchangePaymentFiveNodesWithCommissionsSingleEquivalentNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForExchangePaymentFiveNodesWithCommissionsSingleEquivalentTest(t *testing.T) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, 5)
	for i := range 5 {
		nodes[i] = vtcp.NewNode(t, getNextIPForExchangePaymentFiveNodesWithCommissionsSingleEquivalentTest(), fmt.Sprintf("node%d", i+1))
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
	nodes[2].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "5000")
	nodes[3].CreateAndSetSettlementLineAndCheck(t, nodes[2], testconfig.Equivalent, "250")
	nodes[4].CreateAndSetSettlementLineAndCheck(t, nodes[3], testconfig.Equivalent, "500")

	return nodes, cluster
}

func Test1ExchangePaymentFiveNodesWithCommissionsSingleEquivalent(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentFiveNodesWithCommissionsSingleEquivalentTest(t)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 10},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[4], testconfig.Equivalent, "100", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 2)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.Equivalent, "0", "3000", "-110")
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.Equivalent, "0", "5000", "-100")
	nodes[2].CheckActiveSettlementLine(t, nodes[3], testconfig.Equivalent, "0", "250", "-100")
	nodes[3].CheckActiveSettlementLine(t, nodes[4], testconfig.Equivalent, "0", "500", "-100")

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.Equivalent}, "150")
}

func Test2ExchangePaymentFiveNodesWithCommissionsSingleEquivalent(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentFiveNodesWithCommissionsSingleEquivalentTest(t)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 10},
	})

	nodes[3].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 3},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[4], testconfig.Equivalent, "100", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 2)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.Equivalent, "0", "3000", "-113")
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.Equivalent, "0", "5000", "-103")
	nodes[2].CheckActiveSettlementLine(t, nodes[3], testconfig.Equivalent, "0", "250", "-103")
	nodes[3].CheckActiveSettlementLine(t, nodes[4], testconfig.Equivalent, "0", "500", "-100")

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.Equivalent}, "144")
}

func Test3ExchangePaymentFiveNodesWithCommissionsSingleEquivalent(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentFiveNodesWithCommissionsSingleEquivalentTest(t)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 10},
	})

	nodes[2].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 6},
	})

	nodes[3].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 3},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[4], testconfig.Equivalent, "100", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 2)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.Equivalent, "0", "3000", "-119")
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.Equivalent, "0", "5000", "-109")
	nodes[2].CheckActiveSettlementLine(t, nodes[3], testconfig.Equivalent, "0", "250", "-103")
	nodes[3].CheckActiveSettlementLine(t, nodes[4], testconfig.Equivalent, "0", "500", "-100")

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.Equivalent}, "144")
}
