package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
	"github.com/vTCP-Foundation/vtcpd-test-suite/tests/testconfig"
)

var (
	// exchangePaymentSimpleThreeNodesNextNodeIndex is used to assign unique IP addresses to nodes across different test functions in this file.
	exchangePaymentSimpleThreeNodesNextNodeIndex = 1
)

func getNextIPForExchangePaymentSimpleThreeNodesTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForExchangePaymentSimpleThreeNodes, exchangePaymentSimpleThreeNodesNextNodeIndex)
	exchangePaymentSimpleThreeNodesNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForExchangePaymentSimpleThreeNodesTest(t *testing.T) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, 3)
	for i := range 3 {
		nodes[i] = vtcp.NewNode(t, getNextIPForExchangePaymentSimpleThreeNodesTest(), fmt.Sprintf("node%d", i+1))
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

	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.ExchangeEquivalent, "3000")

	nodes[2].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "250")

	return nodes, cluster
}

func Test1ExchangePaymentSimpleThreeNodes(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentSimpleThreeNodesTest(t)

	nodes[1].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -2, nil, nil, vtcp.StatusOK)

	nodes[0].CheckExchangeMaxFlow(t, nodes[2], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "150")

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[2], testconfig.Equivalent, "100", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	nodes[0].CheckExchangeMaxFlow(t, nodes[2], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "50")
}

func Test2ExchangePaymentSimpleThreeNodes(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentSimpleThreeNodesTest(t)

	nodes[1].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", 0, nil, nil, vtcp.StatusOK)

	nodes[0].CheckExchangeMaxFlow(t, nodes[2], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "250")

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[2], testconfig.Equivalent, "100", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	nodes[0].CheckExchangeMaxFlow(t, nodes[2], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "150")
}

func Test3ExchangePaymentSimpleThreeNodesChangeExchangeRate(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentSimpleThreeNodesTest(t)

	nodes[1].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -2, nil, nil, vtcp.StatusOK)

	nodes[0].CheckExchangeMaxFlow(t, nodes[2], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "150")

	nodes[1].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "6", -2, nil, nil, vtcp.StatusOK)

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[2], testconfig.Equivalent, "100", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-1667")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.Equivalent, "0", "250", "-100")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[2], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "79")
}
