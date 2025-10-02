package main

import (
	"context"
	"fmt"
	"testing"

	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
	"github.com/vTCP-Foundation/vtcpd-test-suite/tests/testconfig"
)

var (
	paymentTimeoutsPartTwoNextNodeIndex = 1
)

func getNextIPForPaymentTimeoutsPartTwoTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForPaymentTimeoutsPartTwoTest, paymentTimeoutsPartTwoNextNodeIndex)
	paymentTimeoutsPartTwoNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForPaymentTimeoutsPartTwoTest(t *testing.T) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, 11)
	for i := range 13 {
		nodes[i] = vtcp.NewNode(t, getNextIPForPaymentTimeoutsPartTwoTest(), fmt.Sprintf("node%d", i+1))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings) // Assuming clusterSettings is defined globally or passed
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes, false)

	nodes[2].OpenChannelAndCheck(t, nodes[0])
	nodes[4].OpenChannelAndCheck(t, nodes[3])
	nodes[6].OpenChannelAndCheck(t, nodes[5])
	nodes[8].OpenChannelAndCheck(t, nodes[7])
	nodes[10].OpenChannelAndCheck(t, nodes[9])
	nodes[1].OpenChannelAndCheck(t, nodes[2])
	nodes[3].OpenChannelAndCheck(t, nodes[0])
	nodes[5].OpenChannelAndCheck(t, nodes[4])
	nodes[9].OpenChannelAndCheck(t, nodes[8])
	nodes[1].OpenChannelAndCheck(t, nodes[6])
	nodes[7].OpenChannelAndCheck(t, nodes[0])
	nodes[2].OpenChannelAndCheck(t, nodes[10])

	nodes[2].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "200")
	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[2], testconfig.Equivalent, "1000")
	nodes[3].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "800")
	nodes[4].CreateAndSetSettlementLineAndCheck(t, nodes[3], testconfig.Equivalent, "500")
	nodes[5].CreateAndSetSettlementLineAndCheck(t, nodes[4], testconfig.Equivalent, "600")
	nodes[6].CreateAndSetSettlementLineAndCheck(t, nodes[5], testconfig.Equivalent, "500")
	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[6], testconfig.Equivalent, "900")
	nodes[7].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "400")
	nodes[8].CreateAndSetSettlementLineAndCheck(t, nodes[7], testconfig.Equivalent, "500")
	nodes[9].CreateAndSetSettlementLineAndCheck(t, nodes[8], testconfig.Equivalent, "500")
	nodes[10].CreateAndSetSettlementLineAndCheck(t, nodes[9], testconfig.Equivalent, "600")
	nodes[2].CreateAndSetSettlementLineAndCheck(t, nodes[10], testconfig.Equivalent, "1000")

	nodes[0].CheckMaxFlow(t, nodes[1], testconfig.Equivalent, "1100")

	return nodes, cluster
}

func TestTimeoutsPartTwo1aCheckNodeOnlyOnOnePathPaymentPass(t *testing.T) {
	nodes, _ := setupNodesForPaymentTimeoutsPartTwoTest(t)

	nodes[3].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[4].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[5].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[6].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[7].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[8].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[9].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[10].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")

	nodes[0].CreateTransactionCheckStatus(t, nodes[1], testconfig.Equivalent, "1000", vtcp.StatusOK)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 60)
}

func TestTimeoutsPartTwo1bCheckNodeLostFinalConfiguration(t *testing.T) {
	nodes, _ := setupNodesForPaymentTimeoutsPartTwoTest(t)

	nodes[3].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[4].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[5].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[6].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[7].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[8].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[9].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[10].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")

	nodes[0].SetTestingFlag(vtcp.FlagForbidSendMessageFinalPathConfig, "", "")

	nodes[0].CreateTransactionCheckStatus(t, nodes[1], testconfig.Equivalent, "1000", vtcp.StatusOK)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 60)
}

func TestTimeoutsPartTwo1cLostMessagesOnOtherPathWithoutPass(t *testing.T) {
	nodes, _ := setupNodesForPaymentTimeoutsPartTwoTest(t)

	nodes[3].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[4].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[5].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[6].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[7].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[8].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[9].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[10].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")

	nodes[1].SetTestingFlag(vtcp.FlagForbidSendResponseToIntemediateOnReservation, nodes[6].IPAddress, "")

	nodes[0].CreateTransactionCheckStatus(t, nodes[1], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 60)
}
