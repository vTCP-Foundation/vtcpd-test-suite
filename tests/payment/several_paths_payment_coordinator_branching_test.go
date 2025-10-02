package main

import (
	"context"
	"fmt"
	"testing"

	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
	"github.com/vTCP-Foundation/vtcpd-test-suite/tests/testconfig"
)

var (
	severalPathPaymentCoordinatorBranchingNextNodeIndex = 1
)

func getNextIPForSeveralPathPaymentCoordinatorBranchingTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForSeveralPathPaymentCoordinatorBranchingTest, severalPathPaymentCoordinatorBranchingNextNodeIndex)
	severalPathPaymentCoordinatorBranchingNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForSeveralPathPaymentCoordinatorBranchingTest(t *testing.T) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, 6)
	for i := range 6 {
		nodes[i] = vtcp.NewNode(t, getNextIPForSeveralPathPaymentCoordinatorBranchingTest(), fmt.Sprintf("node%d", i+1))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings) // Assuming clusterSettings is defined globally or passed
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes, false)

	nodes[1].OpenChannelAndCheck(t, nodes[0])
	nodes[5].OpenChannelAndCheck(t, nodes[2])
	nodes[4].OpenChannelAndCheck(t, nodes[3])
	nodes[2].OpenChannelAndCheck(t, nodes[1])
	nodes[5].OpenChannelAndCheck(t, nodes[4])
	nodes[5].OpenChannelAndCheck(t, nodes[1])
	nodes[3].OpenChannelAndCheck(t, nodes[1])

	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "1000")
	nodes[5].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "500")
	nodes[2].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "800")
	nodes[5].CreateAndSetSettlementLineAndCheck(t, nodes[2], testconfig.Equivalent, "200")
	nodes[3].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "700")
	nodes[4].CreateAndSetSettlementLineAndCheck(t, nodes[3], testconfig.Equivalent, "300")
	nodes[5].CreateAndSetSettlementLineAndCheck(t, nodes[4], testconfig.Equivalent, "900")

	nodes[0].CheckMaxFlow(t, nodes[5], testconfig.Equivalent, "1000")

	return nodes, cluster
}

func Test1CoordinatorBranchingNormalAmount(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentCoordinatorBranchingTest(t)

	nodes[0].CreateTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "1000", vtcp.StatusOK)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[5], testconfig.Equivalent, "0")
}

func Test2aCoordinatorBranchingSeveralRunNextNeighborResponseProcessingStageLostMessageMiddlePathPaymentPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentCoordinatorBranchingTest(t)

	nodes[2].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, "", "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "800", vtcp.StatusOK)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[5], testconfig.Equivalent, "200")
}

func Test2bCoordinatorBranchingSeveralRunNextNeighborResponseProcessingStageLostMessageMiddlePathPaymentDontPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentCoordinatorBranchingTest(t)

	nodes[2].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, "", "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[5], testconfig.Equivalent, "1000")
}

func Test3CoordinatorBranchingLostMsgRunNextNeighborResponseProcessingStageOnLongWay(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentCoordinatorBranchingTest(t)

	nodes[4].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, "", "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[5], testconfig.Equivalent, "1000")
}

func Test4aCoordinatorBranchingLostRunPreviousNeighborRequestProcessingStageMsgFromCommonNodeToCoordinatorPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentCoordinatorBranchingTest(t)

	nodes[1].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, nodes[5].GetIPAddressForRequests(), "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "500", vtcp.StatusOK)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[5], testconfig.Equivalent, "500")
}

func Test4bCoordinatorBranchingLostRunPreviousNeighborRequestProcessingStageMsgFromCommonNodeToCoordinatorDontPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentCoordinatorBranchingTest(t)

	nodes[1].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, nodes[5].GetIPAddressForRequests(), "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[5], testconfig.Equivalent, "1000")
}

func Test5aCoordinatorBranchingLostRunPreviousNeighborRequestProcessingStageMsgFromCommonNodeToCoordinatorPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentCoordinatorBranchingTest(t)

	nodes[1].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, nodes[2].GetIPAddressForRequests(), "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "800", vtcp.StatusOK)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[5], testconfig.Equivalent, "200")
}

func Test5bCoordinatorBranchingLostRunPreviousNeighborRequestProcessingStageMsgFromCommonNodeToCoordinatorDontPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentCoordinatorBranchingTest(t)

	nodes[1].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, nodes[2].GetIPAddressForRequests(), "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[5], testconfig.Equivalent, "1000")
}

func Test6CoordinatorBranchingLostRunPreviousNeighborRequestProcessingStageMsgFromCommonNodeToCoordinatorDontPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentCoordinatorBranchingTest(t)

	nodes[1].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, nodes[3].GetIPAddressForRequests(), "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[5], testconfig.Equivalent, "1000")
}
