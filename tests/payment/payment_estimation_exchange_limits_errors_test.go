package main

import (
	"context"
	"fmt"
	"testing"

	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
	"github.com/vTCP-Foundation/vtcpd-test-suite/tests/testconfig"
)

var (
	paymentEstimationExchangeLimitsNextNodeIndex = 1
)

func getNextIPForPaymentEstimationExchangeLimits() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForPaymentEstimationExchangeLimits, paymentEstimationExchangeLimitsNextNodeIndex)
	paymentEstimationExchangeLimitsNextNodeIndex++
	return ip
}

// setupNodesForPaymentEstimationExchangeLimits creates a minimal exchange topology with min/max exchange constraints at X.
func setupNodesForPaymentEstimationExchangeLimits(t *testing.T) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, 3) // A=0, X=1, B=2
	for i := range 3 {
		nodes[i] = vtcp.NewNode(t, getNextIPForPaymentEstimationExchangeLimits(), fmt.Sprintf("node%d", i+1))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes, false)

	// Channels
	nodes[1].OpenChannelAndCheck(t, nodes[0]) // X-A
	nodes[2].OpenChannelAndCheck(t, nodes[1]) // B-X

	// Settlement lines
	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.ExchangeEquivalent, "1000") // A->X eq=1001
	nodes[2].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "2000")         // X->B eq=2002

	return nodes, cluster
}

// TestEstimate_Exchange_MinMaxLimits covers min/max exchange constraints and missing cache error.
func TestEstimateExchangeMinMaxLimits(t *testing.T) {
	nodes, _ := setupNodesForPaymentEstimationExchangeLimits(t)

	// Set rate 1.5 with min=100 (in sender eq) and max=500 (in sender eq)
	min := "100"
	max := "900"
	nodes[1].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "15", -1, &min, &max, vtcp.StatusOK) // 1.5

	// Populate cache
	_, _ = nodes[0].GetExchangeMaxFlow(t, nodes[2], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent})

	// Below min: receive=50 (2002) requires ~33 in 1001 < 100 -> expect 412
	nodes[0].CheckEstimatePaymentForReceiveAmount(t, nodes[2], "50", testconfig.ExchangeEquivalent, testconfig.Equivalent, "", vtcp.StatusInsufficientFunds)
	nodes[0].CheckEstimateReceiveForPaymentAmount(t, nodes[2], "50", testconfig.ExchangeEquivalent, testconfig.Equivalent, "", vtcp.StatusInsufficientFunds)

	// At cap: payment=600 (1001) capped to 500 -> output 500*1.5=750 (2002)
	nodes[0].CheckEstimateReceiveForPaymentAmount(t, nodes[2], "500", testconfig.ExchangeEquivalent, testconfig.Equivalent, "750", vtcp.StatusOK)
	nodes[0].CheckEstimatePaymentForReceiveAmount(t, nodes[2], "750", testconfig.ExchangeEquivalent, testconfig.Equivalent, "500", vtcp.StatusOK)

	// Missing cache for different pair: try swapped equivalents should be 462 (no cached paths)
	nodes[0].CheckEstimateReceiveForPaymentAmount(t, nodes[2], "100", testconfig.Equivalent, testconfig.ExchangeEquivalent, "", vtcp.StatusNoPaymentRoutes) // Use API mapping in CLI docs (462 -> 404), but engine returns 462; HTTP layer may map. Here we expect engine code 462.

	// On boundaries: expect OK
	nodes[0].CheckEstimateReceiveForPaymentAmount(t, nodes[2], "900", testconfig.ExchangeEquivalent, testconfig.Equivalent, "1350", vtcp.StatusOK)
	nodes[0].CheckEstimatePaymentForReceiveAmount(t, nodes[2], "1350", testconfig.ExchangeEquivalent, testconfig.Equivalent, "900", vtcp.StatusOK)

	// Above max: expect insufficient funds
	nodes[0].CheckEstimateReceiveForPaymentAmount(t, nodes[2], "950", testconfig.ExchangeEquivalent, testconfig.Equivalent, "", vtcp.StatusInsufficientFunds)
	nodes[0].CheckEstimatePaymentForReceiveAmount(t, nodes[2], "1400", testconfig.ExchangeEquivalent, testconfig.Equivalent, "", vtcp.StatusInsufficientFunds)
}
