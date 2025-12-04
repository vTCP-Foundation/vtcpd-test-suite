package main

import (
	"context"
	"fmt"
	"testing"

	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
	"github.com/vTCP-Foundation/vtcpd-test-suite/tests/testconfig"
)

var (
	paymentEstimationExchangeSimpleNextNodeIndex = 1
)

func getNextIPForPaymentEstimationExchangeSimple() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForPaymentEstimationExchangeSimple, paymentEstimationExchangeSimpleNextNodeIndex)
	paymentEstimationExchangeSimpleNextNodeIndex++
	return ip
}

// setupNodesForPaymentEstimationExchangeSimple creates a 4-node topology:
// A(1001) -> B(1001) -> X(exchange 1001->2002) -> C(2002)
// Settlement lines exist in appropriate equivalents to enable an exchange path.
func setupNodesForPaymentEstimationExchangeSimple(t *testing.T) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, 4)
	for i := range 4 {
		nodes[i] = vtcp.NewNode(t, getNextIPForPaymentEstimationExchangeSimple(), fmt.Sprintf("node%d", i+1))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes, false)

	// Open channels
	nodes[1].OpenChannelAndCheck(t, nodes[0]) // B-A
	nodes[2].OpenChannelAndCheck(t, nodes[1]) // X-B
	nodes[3].OpenChannelAndCheck(t, nodes[2]) // C-X

	// Setup settlement lines in sender equivalent (1001) along A->B->X
	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.ExchangeEquivalent, "2000")
	nodes[2].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.ExchangeEquivalent, "2000")

	// Setup settlement line in receiver equivalent (2002) along X->C
	nodes[3].CreateAndSetSettlementLineAndCheck(t, nodes[2], testconfig.Equivalent, "4000")

	return nodes, cluster
}

// TestEstimateReceiveAndPayment_SimpleExchange_NoCommissions verifies estimation for a simple exchange path with rate 2.0.
func TestEstimateReceiveAndPaymentSimpleExchangeWithCommissions(t *testing.T) {
	nodes, _ := setupNodesForPaymentEstimationExchangeSimple(t)

	// Configure exchange rate at node X: 1001 -> 2002 at 2.0
	// value=2, shift=0 represents 2.0
	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "2", 0, nil, nil, vtcp.StatusOK)
	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	// Trigger exchange max-flow to populate cached optimal paths
	_, _ = nodes[0].GetExchangeMaxFlow(t, nodes[3], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent})

	// Payment -> Receive: 200 (1001) -> expect 400 (2002)
	nodes[0].CheckEstimateReceiveForPaymentAmount(t, nodes[3], "200", testconfig.ExchangeEquivalent, testconfig.Equivalent, "380", vtcp.StatusOK)

	// Receive -> Payment: need to deliver 400 (2002) -> expect 200 (1001)
	nodes[0].CheckEstimatePaymentForReceiveAmount(t, nodes[3], "400", testconfig.ExchangeEquivalent, testconfig.Equivalent, "210", vtcp.StatusOK)

	nodes[0].CheckEstimateReceiveForPaymentAmount(t, nodes[3], "2100", testconfig.ExchangeEquivalent, testconfig.Equivalent, "", vtcp.StatusInsufficientFunds)
	nodes[0].CheckEstimateReceiveForPaymentAmount(t, nodes[3], "4000", testconfig.ExchangeEquivalent, testconfig.Equivalent, "", vtcp.StatusInsufficientFunds)
}
