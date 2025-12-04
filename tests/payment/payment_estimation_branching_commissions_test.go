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
	paymentEstimationBranchingNextNodeIndex = 1
)

func getNextIPForPaymentEstimationBranching() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForPaymentEstimationBranching, paymentEstimationBranchingNextNodeIndex)
	paymentEstimationBranchingNextNodeIndex++
	return ip
}

// setupNodesForPaymentEstimationBranching builds a branched topology with commissions:
// A -- B -- F
//
//	\- B - C - F
//	\- B - D - E - F
//
// All in equivalent 2002 (no exchange), commission at B in eq 2002.
func setupNodesForPaymentEstimationBranching(t *testing.T) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, 6) // A=0, B=1, C=2, D=3, E=4, F=5
	for i := range 6 {
		nodes[i] = vtcp.NewNode(t, getNextIPForPaymentEstimationBranching(), fmt.Sprintf("node%d", i+1))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes, false)

	// Channels
	nodes[1].OpenChannelAndCheck(t, nodes[0]) // B-A
	nodes[5].OpenChannelAndCheck(t, nodes[1]) // F-B
	nodes[2].OpenChannelAndCheck(t, nodes[1]) // C-B
	nodes[5].OpenChannelAndCheck(t, nodes[2]) // F-C
	nodes[3].OpenChannelAndCheck(t, nodes[1]) // D-B
	nodes[4].OpenChannelAndCheck(t, nodes[3]) // E-D
	nodes[5].OpenChannelAndCheck(t, nodes[4]) // F-E

	// Settlement lines, all in 2002
	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "1000") // A->B
	nodes[5].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "500")  // B->F
	nodes[2].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "800")  // B->C
	nodes[5].CreateAndSetSettlementLineAndCheck(t, nodes[2], testconfig.Equivalent, "200")  // C->F
	nodes[3].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "700")  // B->D
	nodes[4].CreateAndSetSettlementLineAndCheck(t, nodes[3], testconfig.Equivalent, "300")  // D->E
	nodes[5].CreateAndSetSettlementLineAndCheck(t, nodes[4], testconfig.Equivalent, "900")  // E->F

	// Commission at B in eq 2002: 10
	nodes[1].SetCommissions([]vtcp.CommissionPair{{Equivalent: testconfig.Equivalent, Amount: 10}})

	// Give some time for config to apply and node restart if occurs
	time.Sleep(2 * time.Second)

	return nodes, cluster
}

// TestEstimate_Payment_Receive_Branching_WithCommission validates "charge once" commission handling across multiple paths.
func TestEstimatePaymentReceiveBranchingWithCommission(t *testing.T) {
	nodes, _ := setupNodesForPaymentEstimationBranching(t)

	// Trigger exchange max-flow (sender==receiver equivalent) to populate cached optimal paths
	_, _ = nodes[0].GetExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.Equivalent})

	// Receive -> Payment: want to deliver 990, expect ~1000 due to single commission 10 once across paths
	nodes[0].CheckEstimatePaymentForReceiveAmount(t, nodes[5], "990", testconfig.Equivalent, testconfig.Equivalent, "1000", vtcp.StatusOK)

	// Payment -> Receive: pay 1000, expect ~990 after commission
	nodes[0].CheckEstimateReceiveForPaymentAmount(t, nodes[5], "1000", testconfig.Equivalent, testconfig.Equivalent, "990", vtcp.StatusOK)

	nodes[0].CheckEstimatePaymentForReceiveAmount(t, nodes[5], "500", testconfig.Equivalent, testconfig.Equivalent, "510", vtcp.StatusOK)
	nodes[0].CheckEstimateReceiveForPaymentAmount(t, nodes[5], "500", testconfig.Equivalent, testconfig.Equivalent, "490", vtcp.StatusOK)

	nodes[0].CheckEstimatePaymentForReceiveAmount(t, nodes[5], "1000", testconfig.Equivalent, testconfig.Equivalent, "0", vtcp.StatusInsufficientFunds)
	nodes[0].CheckEstimateReceiveForPaymentAmount(t, nodes[5], "1100", testconfig.Equivalent, testconfig.Equivalent, "0", vtcp.StatusInsufficientFunds)

	nodes[0].CheckEstimatePaymentForReceiveAmount(t, nodes[4], "990", testconfig.Equivalent, testconfig.Equivalent, "0", vtcp.StatusNoPaymentRoutes)
	nodes[0].CheckEstimateReceiveForPaymentAmount(t, nodes[4], "1000", testconfig.Equivalent, testconfig.Equivalent, "0", vtcp.StatusNoPaymentRoutes)
}
