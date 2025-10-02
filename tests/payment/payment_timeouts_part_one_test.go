package main

import (
	"context"
	"fmt"
	"testing"

	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
	"github.com/vTCP-Foundation/vtcpd-test-suite/tests/testconfig"
)

var (
	paymentTimeoutsPartOneNextNodeIndex = 1
)

func getNextIPForPaymentTimeoutsPartOneTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForPaymentTimeoutsPartOneTest, paymentTimeoutsPartOneNextNodeIndex)
	paymentTimeoutsPartOneNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForPaymentTimeoutsPartOneTest(t *testing.T) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, 13)
	for i := range 13 {
		nodes[i] = vtcp.NewNode(t, getNextIPForPaymentTimeoutsPartOneTest(), fmt.Sprintf("node%d", i+1))
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
	nodes[1].OpenChannelAndCheck(t, nodes[7])
	nodes[9].OpenChannelAndCheck(t, nodes[8])
	nodes[11].OpenChannelAndCheck(t, nodes[10])
	nodes[1].OpenChannelAndCheck(t, nodes[2])
	nodes[3].OpenChannelAndCheck(t, nodes[0])
	nodes[5].OpenChannelAndCheck(t, nodes[4])
	nodes[7].OpenChannelAndCheck(t, nodes[6])
	nodes[10].OpenChannelAndCheck(t, nodes[9])
	nodes[12].OpenChannelAndCheck(t, nodes[11])
	nodes[8].OpenChannelAndCheck(t, nodes[0])
	nodes[1].OpenChannelAndCheck(t, nodes[12])

	nodes[2].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "200")
	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[2], testconfig.Equivalent, "100")
	nodes[3].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "800")
	nodes[4].CreateAndSetSettlementLineAndCheck(t, nodes[3], testconfig.Equivalent, "500")
	nodes[5].CreateAndSetSettlementLineAndCheck(t, nodes[4], testconfig.Equivalent, "600")
	nodes[6].CreateAndSetSettlementLineAndCheck(t, nodes[5], testconfig.Equivalent, "500")
	nodes[7].CreateAndSetSettlementLineAndCheck(t, nodes[6], testconfig.Equivalent, "1000")
	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[7], testconfig.Equivalent, "900")
	nodes[8].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "400")
	nodes[9].CreateAndSetSettlementLineAndCheck(t, nodes[8], testconfig.Equivalent, "500")
	nodes[10].CreateAndSetSettlementLineAndCheck(t, nodes[9], testconfig.Equivalent, "600")
	nodes[11].CreateAndSetSettlementLineAndCheck(t, nodes[10], testconfig.Equivalent, "500")
	nodes[12].CreateAndSetSettlementLineAndCheck(t, nodes[11], testconfig.Equivalent, "1000")
	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[12], testconfig.Equivalent, "900")

	nodes[0].CheckMaxFlow(t, nodes[1], testconfig.Equivalent, "1000")

	return nodes, cluster
}

func TestTimeoutsPartOne1aCheckNodeOnlyOnOnePathPaymentPass(t *testing.T) {
	nodes, _ := setupNodesForPaymentTimeoutsPartOneTest(t)

	nodes[3].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[4].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[5].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[6].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[7].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[8].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[9].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[10].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[11].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[12].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")

	nodes[0].CreateTransactionCheckStatus(t, nodes[1], testconfig.Equivalent, "1000", vtcp.StatusOK)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 60)
}

func TestTimeoutsPartOne1bCheckNodeLostFinalConfiguration(t *testing.T) {
	nodes, _ := setupNodesForPaymentTimeoutsPartOneTest(t)

	nodes[3].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[4].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[5].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[6].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[7].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[8].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[9].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[10].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[11].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[12].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[0].SetTestingFlag(vtcp.FlagForbidSendMessageFinalPathConfig, "", "")

	nodes[0].CreateTransactionCheckStatus(t, nodes[1], testconfig.Equivalent, "1000", vtcp.StatusOK)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 60)
}

func TestTimeoutsPartOne1cLostMessagesOnOtherPathWithoutPass(t *testing.T) {
	nodes, _ := setupNodesForPaymentTimeoutsPartOneTest(t)

	nodes[3].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[4].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[5].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[6].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[7].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[8].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[9].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[10].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[11].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[12].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[0].SetTestingFlag(vtcp.FlagForbidSendResponseToIntemediateOnReservation, nodes[12].IPAddress, "")

	nodes[0].CreateTransactionCheckStatus(t, nodes[1], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 60)
}

func TestTimeoutsPartOne2aCheckTTLReceiverWithPass(t *testing.T) {
	nodes, _ := setupNodesForPaymentTimeoutsPartOneTest(t)

	nodes[3].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[4].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[5].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[6].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[7].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[8].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[9].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[10].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[11].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[12].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[2].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, "", "")

	nodes[0].CreateTransactionCheckStatus(t, nodes[1], testconfig.Equivalent, "400", vtcp.StatusOK)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 60)
}

func TestTimeoutsPartOne2bCheckTTLReceiverWithoutPass(t *testing.T) {
	nodes, _ := setupNodesForPaymentTimeoutsPartOneTest(t)

	nodes[3].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[4].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[5].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[6].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[7].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[8].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[9].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[10].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[11].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[12].SetTestingFlag(vtcp.FlagSleepOnNextNeighborResponseProcessing, "", "")
	nodes[2].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, "", "")

	nodes[0].CreateTransactionCheckStatus(t, nodes[1], testconfig.Equivalent, "500", vtcp.StatusInsufficientFunds)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 60)
}
