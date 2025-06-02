package main

import (
	"context"
	"fmt"
	"testing"

	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
	"github.com/vTCP-Foundation/vtcpd-test-suite/tests/testconfig"
)

var (
	severalPathPaymentReceiverBranchingNextNodeIndex = 1
)

func getNextIPForSeveralPathPaymentReceiverBranchingTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForSeveralPathPaymentReceiverBranchingTest, severalPathPaymentReceiverBranchingNextNodeIndex)
	severalPathPaymentReceiverBranchingNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForSeveralPathPaymentReceiverBranchingTest(t *testing.T) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, 6)
	for i := range 6 {
		nodes[i] = vtcp.NewNode(t, getNextIPForSeveralPathPaymentReceiverBranchingTest(), fmt.Sprintf("node%d", i+1))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes)

	nodes[1].OpenChannelAndCheck(t, nodes[0])
	nodes[4].OpenChannelAndCheck(t, nodes[3])
	nodes[5].OpenChannelAndCheck(t, nodes[1])
	nodes[2].OpenChannelAndCheck(t, nodes[0])
	nodes[1].OpenChannelAndCheck(t, nodes[2])
	nodes[3].OpenChannelAndCheck(t, nodes[0])
	nodes[1].OpenChannelAndCheck(t, nodes[4])

	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "500")
	nodes[5].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "1000")
	nodes[2].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "800")
	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[2], testconfig.Equivalent, "200")
	nodes[3].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "700")
	nodes[4].CreateAndSetSettlementLineAndCheck(t, nodes[3], testconfig.Equivalent, "300")
	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[4], testconfig.Equivalent, "900")

	nodes[0].CheckMaxFlow(t, nodes[5], testconfig.Equivalent, "1000")

	return nodes, cluster
}

func Test1ReceiverBranchingNormalAmount(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentReceiverBranchingTest(t)

	nodes[0].CreateTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "1000", vtcp.StatusOK)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[5], testconfig.Equivalent, "0")
}

func Test2aReceiverBranchingSeveralRunNextNeighborResponseProcessingStageLostMessageMiddlePathPaymentPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentReceiverBranchingTest(t)

	nodes[2].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, "", "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "800", vtcp.StatusOK)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[5], testconfig.Equivalent, "200")
}

func Test2bReceiverBranchingSeveralRunNextNeighborResponseProcessingStageLostMessageMiddlePathPaymentDontPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentReceiverBranchingTest(t)

	nodes[2].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, "", "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[5], testconfig.Equivalent, "1000")
}

func Test3ReceiverBranchingLostMsgRunNextNeighborResponseProcessingStageOnLongWay(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentReceiverBranchingTest(t)

	nodes[4].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, "", "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[5], testconfig.Equivalent, "1000")
}

func Test4aReceiverBranchingLostRunPreviousNeighborRequestProcessingStageMsgFromCommonNodeToCoordinatorPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentReceiverBranchingTest(t)

	nodes[1].SetTestingFlag(vtcp.FlagForbidSendResponseToIntemediateOnReservation, nodes[0].GetIPAddressForRequests(), "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "500", vtcp.StatusOK)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[5], testconfig.Equivalent, "500")
}

func Test4bReceiverBranchingLostRunPreviousNeighborRequestProcessingStageMsgFromCommonNodeToCoordinatorDontPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentReceiverBranchingTest(t)

	nodes[1].SetTestingFlag(vtcp.FlagForbidSendResponseToIntemediateOnReservation, nodes[0].GetIPAddressForRequests(), "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[5], testconfig.Equivalent, "1000")
}

func Test5aReceiverBranchingLostRunPreviousNeighborRequestProcessingStageMsgFromCommonNodeToCoordinatorPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentReceiverBranchingTest(t)

	nodes[1].SetTestingFlag(vtcp.FlagForbidSendResponseToIntemediateOnReservation, nodes[2].GetIPAddressForRequests(), "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "800", vtcp.StatusOK)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[5], testconfig.Equivalent, "200")
}

func Test5bReceiverBranchingLostRunPreviousNeighborRequestProcessingStageMsgFromCommonNodeToCoordinatorDontPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentReceiverBranchingTest(t)

	nodes[1].SetTestingFlag(vtcp.FlagForbidSendResponseToIntemediateOnReservation, nodes[2].GetIPAddressForRequests(), "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[5], testconfig.Equivalent, "1000")
}

func Test6ReceiverBranchingLostRunPreviousNeighborRequestProcessingStageMsgFromCommonNodeToPreviousNode(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentReceiverBranchingTest(t)

	nodes[1].SetTestingFlag(vtcp.FlagForbidSendResponseToIntemediateOnReservation, nodes[0].GetIPAddressForRequests(), "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[5], testconfig.Equivalent, "1000")
}

func Test7ReceiverBranchingLostRunPreviousNeighborRequestProcessingStageMsgFromReceiver(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentReceiverBranchingTest(t)

	nodes[5].SetTestingFlag(vtcp.FlagForbidSendResponseToIntemediateOnReservation, "", "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "500", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[5], testconfig.Equivalent, "1000")
}
