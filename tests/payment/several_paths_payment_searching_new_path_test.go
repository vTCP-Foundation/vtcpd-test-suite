package main

import (
	"context"
	"fmt"
	"testing"

	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
	"github.com/vTCP-Foundation/vtcpd-test-suite/tests/testconfig"
)

var (
	severalPathsPaymentSearchingNewPathNextNodeIndex = 1
)

func getNextIPForSeveralPathsPaymentSearchingNewPathTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForSeveralPathsPaymentSearchingNewPathTest, severalPathsPaymentSearchingNewPathNextNodeIndex)
	severalPathsPaymentSearchingNewPathNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForSeveralPathsPaymentSearchingNewPathTest(t *testing.T) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, 9)
	for i := range 9 {
		nodes[i] = vtcp.NewNode(t, getNextIPForSeveralPathsPaymentSearchingNewPathTest(), fmt.Sprintf("node%d", i+1))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes)

	// Setup topology according to Python version
	nodes[8].OpenChannelAndCheck(t, nodes[0])
	nodes[2].OpenChannelAndCheck(t, nodes[1])
	nodes[4].OpenChannelAndCheck(t, nodes[3])
	nodes[6].OpenChannelAndCheck(t, nodes[5])
	nodes[8].OpenChannelAndCheck(t, nodes[7])
	nodes[1].OpenChannelAndCheck(t, nodes[0])
	nodes[8].OpenChannelAndCheck(t, nodes[2])
	nodes[7].OpenChannelAndCheck(t, nodes[6])
	nodes[3].OpenChannelAndCheck(t, nodes[1])
	nodes[8].OpenChannelAndCheck(t, nodes[4])
	nodes[5].OpenChannelAndCheck(t, nodes[1])

	nodes[8].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "900")
	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "1000")
	nodes[2].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "1500")
	nodes[8].CreateAndSetSettlementLineAndCheck(t, nodes[2], testconfig.Equivalent, "2000")
	nodes[3].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "1000")
	nodes[4].CreateAndSetSettlementLineAndCheck(t, nodes[3], testconfig.Equivalent, "2000")
	nodes[8].CreateAndSetSettlementLineAndCheck(t, nodes[4], testconfig.Equivalent, "1000")
	nodes[5].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "2000")
	nodes[6].CreateAndSetSettlementLineAndCheck(t, nodes[5], testconfig.Equivalent, "1500")
	nodes[7].CreateAndSetSettlementLineAndCheck(t, nodes[6], testconfig.Equivalent, "1500")
	nodes[8].CreateAndSetSettlementLineAndCheck(t, nodes[7], testconfig.Equivalent, "2000")

	nodes[0].CheckMaxFlow(t, nodes[8], testconfig.Equivalent, "1900")

	return nodes, cluster
}

func Test1aSearchingNewPathNormalPassWithoutSearchingAdditionalPaths(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathsPaymentSearchingNewPathTest(t)

	nodes[8].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, "", "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[8], testconfig.Equivalent, "1000", vtcp.StatusOK)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[8], testconfig.Equivalent, "900")
}

func Test1bSearchingNewPathNormalPassWithSearchingAdditionalPath(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathsPaymentSearchingNewPathTest(t)

	nodes[8].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, "", "")
	nodes[1].SetTestingFlag(vtcp.FlagForbidSendRequestToIntermediateReservation, nodes[2].GetIPAddressForRequests(), "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[8], testconfig.Equivalent, "1000", vtcp.StatusOK)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[8], testconfig.Equivalent, "900")
}

func Test1cSearchingNewPathWithSearchingTwoAdditionalPaths(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathsPaymentSearchingNewPathTest(t)

	nodes[8].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, "", "")
	nodes[1].SetTestingFlag(vtcp.FlagForbidSendRequestToIntermediateReservation, nodes[2].GetIPAddressForRequests(), "")
	nodes[3].SetTestingFlag(vtcp.FlagForbidSendResponseToIntemediateOnReservation, "", "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[8], testconfig.Equivalent, "1000", vtcp.StatusOK)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[8], testconfig.Equivalent, "900")
}

func Test2aSearchingNewPathWithoutPassNoEnoughAmount(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathsPaymentSearchingNewPathTest(t)

	nodes[8].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, "", "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[8], testconfig.Equivalent, "1500", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[8], testconfig.Equivalent, "1900")
}

func Test2bSearchingNewPathWithoutPassLostMessage(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathsPaymentSearchingNewPathTest(t)

	nodes[8].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, "", "")
	nodes[1].SetTestingFlag(vtcp.FlagForbidSendRequestToIntermediateReservation, nodes[2].GetIPAddressForRequests(), "")
	nodes[3].SetTestingFlag(vtcp.FlagForbidSendResponseToIntemediateOnReservation, "", "")
	nodes[5].SetTestingFlag(vtcp.FlagForbidSendResponseToIntemediateOnReservation, "", "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[8], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[8], testconfig.Equivalent, "1900")
}
