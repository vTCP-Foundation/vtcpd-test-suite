package main

import (
	"context"
	"fmt"
	"testing"

	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
	"github.com/vTCP-Foundation/vtcpd-test-suite/tests/testconfig"
)

var (
	severalPathPaymentIntermidiateBranchingNextNodeIndex = 1
)

func getNextIPForSeveralPathPaymentIntermidiateBranchingTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForSeveralPathPaymentIntermidiateBranchingTest, severalPathPaymentIntermidiateBranchingNextNodeIndex)
	severalPathPaymentIntermidiateBranchingNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t *testing.T) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, 8)
	for i := range 8 {
		nodes[i] = vtcp.NewNode(t, getNextIPForSeveralPathPaymentIntermidiateBranchingTest(), fmt.Sprintf("node%d", i+1))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes, false)

	// Setup topology according to Python version
	nodes[1].OpenChannelAndCheck(t, nodes[0])
	nodes[3].OpenChannelAndCheck(t, nodes[2])
	nodes[5].OpenChannelAndCheck(t, nodes[4])
	nodes[7].OpenChannelAndCheck(t, nodes[6])
	nodes[4].OpenChannelAndCheck(t, nodes[1])
	nodes[2].OpenChannelAndCheck(t, nodes[0])
	nodes[6].OpenChannelAndCheck(t, nodes[5])
	nodes[4].OpenChannelAndCheck(t, nodes[3])
	nodes[4].OpenChannelAndCheck(t, nodes[0])
	nodes[7].OpenChannelAndCheck(t, nodes[5])

	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "300")
	nodes[4].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "600")
	nodes[2].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "800")
	nodes[3].CreateAndSetSettlementLineAndCheck(t, nodes[2], testconfig.Equivalent, "200")
	nodes[4].CreateAndSetSettlementLineAndCheck(t, nodes[3], testconfig.Equivalent, "300")
	nodes[4].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "500")
	nodes[5].CreateAndSetSettlementLineAndCheck(t, nodes[4], testconfig.Equivalent, "1000")
	nodes[6].CreateAndSetSettlementLineAndCheck(t, nodes[5], testconfig.Equivalent, "500")
	nodes[7].CreateAndSetSettlementLineAndCheck(t, nodes[6], testconfig.Equivalent, "500")
	nodes[7].CreateAndSetSettlementLineAndCheck(t, nodes[5], testconfig.Equivalent, "500")

	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "1000")

	return nodes, cluster
}

func Test1IntermediateBranchingNormalAmount(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t)

	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "1000", vtcp.StatusOK)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "0")
}

func Test2aIntermediateBranchingRunNextNeighborResponseProcessingStageLostMsgCommonLeftNodeOnShortPathToCoordinatorPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t)

	nodes[4].SetTestingFlag(t, vtcp.FlagForbidSendMessageToCoordinatorReservation, nodes[5].GetIPAddressForRequests(), "500")
	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "500", vtcp.StatusOK)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "500")
}

func Test2bIntermediateBranchingRunNextNeighborResponseProcessingStageLostMsgCommonLeftNodeOnShortPathToCoordinatorNotPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t)

	nodes[4].SetTestingFlag(t, vtcp.FlagForbidSendMessageToCoordinatorReservation, nodes[5].GetIPAddressForRequests(), "500")
	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "1000")
}

func Test3aIntermediateBranchingRunNextNeighborResponseProcessingStageLostMsgCommonLeftNodeOnMiddlePathToCoordinatorPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t)

	nodes[4].SetTestingFlag(t, vtcp.FlagForbidSendMessageToCoordinatorReservation, nodes[5].GetIPAddressForRequests(), "300")
	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "700", vtcp.StatusOK)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "300")
}

func Test3bIntermediateBranchingRunNextNeighborResponseProcessingStageLostMsgCommonLeftNodeOnMiddlePathToCoordinatorNotPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t)

	nodes[4].SetTestingFlag(t, vtcp.FlagForbidSendMessageToCoordinatorReservation, nodes[5].GetIPAddressForRequests(), "300")
	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "1000")
}

func Test4IntermediateBranchingRunNextNeighborResponseProcessingStageLostMsgCommonLeftNodeOnLongPathToCoordinatorNotPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t)

	nodes[4].SetTestingFlag(t, vtcp.FlagForbidSendMessageToCoordinatorReservation, nodes[5].GetIPAddressForRequests(), "200")
	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "1000")
}

func Test5aIntermediateBranchingRunPreviousNeighborRequestProcessingStageLostMsgCommonLeftNodeOnShortPathToCoordinatorPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t)

	nodes[4].SetTestingFlag(t, vtcp.FlagForbidSendResponseToIntemediateOnReservation, nodes[0].GetIPAddressForRequests(), "500")
	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "500", vtcp.StatusOK)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "500")
}

func Test5bIntermediateBranchingRunPreviousNeighborRequestProcessingStageLostMsgCommonLeftNodeOnShortPathToCoordinatorNotPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t)

	nodes[4].SetTestingFlag(t, vtcp.FlagForbidSendResponseToIntemediateOnReservation, nodes[0].GetIPAddressForRequests(), "500")
	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "1000")
}

func Test6aIntermediateBranchingRunPreviousNeighborRequestProcessingStageLostMsgCommonLeftNodeOnMiddlePathToCoordinatorPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t)

	nodes[4].SetTestingFlag(t, vtcp.FlagForbidSendResponseToIntemediateOnReservation, nodes[1].GetIPAddressForRequests(), "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "700", vtcp.StatusOK)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "300")
}

func Test6bIntermediateBranchingRunPreviousNeighborRequestProcessingStageLostMsgCommonLeftNodeOnMiddlePathToCoordinatorNotPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t)

	nodes[4].SetTestingFlag(t, vtcp.FlagForbidSendResponseToIntemediateOnReservation, nodes[1].GetIPAddressForRequests(), "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "1000")
}

func Test7IntermediateBranchingRunPreviousNeighborRequestProcessingStageLostMsgCommonLeftNodeOnLongPathToCoordinatorNotPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t)

	nodes[4].SetTestingFlag(t, vtcp.FlagForbidSendResponseToIntemediateOnReservation, nodes[3].GetIPAddressForRequests(), "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "1000")
}

func Test8aIntermediateBranchingRunNextNeighborResponseProcessingStageLostMsgCommonRightNodeOnShortPathToCoordinatorPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t)

	nodes[5].SetTestingFlag(t, vtcp.FlagForbidSendMessageToCoordinatorReservation, nodes[7].GetIPAddressForRequests(), "500")
	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "500", vtcp.StatusOK)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "500")
}

func Test8bIntermediateBranchingRunNextNeighborResponseProcessingStageLostMsgCommonRightNodeOnShortPathToCoordinatorNotPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t)

	nodes[5].SetTestingFlag(t, vtcp.FlagForbidSendMessageToCoordinatorReservation, nodes[7].GetIPAddressForRequests(), "500")
	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "1000")
}

func Test9aIntermediateBranchingRunNextNeighborResponseProcessingStageLostMsgCommonRightNodeOnMiddlePathToCoordinatorPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t)

	nodes[5].SetTestingFlag(t, vtcp.FlagForbidSendMessageToCoordinatorReservation, nodes[6].GetIPAddressForRequests(), "300")
	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "700", vtcp.StatusOK)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "300")
}

func Test9bIntermediateBranchingRunNextNeighborResponseProcessingStageLostMsgCommonRightNodeOnMiddlePathToCoordinatorNotPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t)

	nodes[5].SetTestingFlag(t, vtcp.FlagForbidSendMessageToCoordinatorReservation, nodes[6].GetIPAddressForRequests(), "300")
	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "1000")
}

func Test10IntermediateBranchingRunNextNeighborResponseProcessingStageLostMsgCommonRightNodeOnLongPathToCoordinatorNotPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t)

	nodes[5].SetTestingFlag(t, vtcp.FlagForbidSendMessageToCoordinatorReservation, nodes[6].GetIPAddressForRequests(), "200")
	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "1000")
}

func Test11aIntermediateBranchingRunPreviousNeighborRequestProcessingStageLostMsgCommonRightNodeOnShortPathToCoordinatorPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t)

	nodes[5].SetTestingFlag(t, vtcp.FlagForbidSendResponseToIntemediateOnReservation, nodes[4].GetIPAddressForRequests(), "500")
	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "500", vtcp.StatusOK)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "500")
}

func Test11bIntermediateBranchingRunPreviousNeighborRequestProcessingStageLostMsgCommonRightNodeOnShortPathToCoordinatorNotPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t)

	nodes[5].SetTestingFlag(t, vtcp.FlagForbidSendResponseToIntemediateOnReservation, nodes[4].GetIPAddressForRequests(), "500")
	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "1000")
}

func Test12aIntermediateBranchingRunPreviousNeighborRequestProcessingStageLostMsgCommonRightNodeOnMiddlePathToCoordinatorPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t)

	nodes[5].SetTestingFlag(t, vtcp.FlagForbidSendResponseToIntemediateOnReservation, nodes[4].GetIPAddressForRequests(), "300")
	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "700", vtcp.StatusOK)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "300")
}

func Test12bIntermediateBranchingRunPreviousNeighborRequestProcessingStageLostMsgCommonRightNodeOnMiddlePathToCoordinatorNotPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t)

	nodes[5].SetTestingFlag(t, vtcp.FlagForbidSendResponseToIntemediateOnReservation, nodes[4].GetIPAddressForRequests(), "300")
	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "1000")
}

func Test13IntermediateBranchingRunPreviousNeighborRequestProcessingStageLostMsgCommonRightNodeOnLongPathToCoordinatorNotPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t)

	nodes[5].SetTestingFlag(t, vtcp.FlagForbidSendResponseToIntemediateOnReservation, nodes[4].GetIPAddressForRequests(), "200")
	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "1000")
}

func Test14aIntermediateBranchingAddNewPathsPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t)

	nodes[0].SetTestingFlag(t, vtcp.FlagForbidSendRequestToIntermediateReservation, nodes[4].GetIPAddressForRequests(), "")
	nodes[1].SetTestingFlag(t, vtcp.FlagForbidSendMessageToCoordinatorReservation, "", "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "100", vtcp.StatusOK)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "900")
}

func Test14bIntermediateBranchingAddNewPathsNotPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t)

	nodes[0].SetTestingFlag(t, vtcp.FlagForbidSendRequestToIntermediateReservation, nodes[4].GetIPAddressForRequests(), "")
	nodes[1].SetTestingFlag(t, vtcp.FlagForbidSendMessageToCoordinatorReservation, "", "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "300", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "1000")
}

func Test14cIntermediateBranchingAddNewPathsMsgLostNotPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathPaymentIntermidiateBranchingTest(t)

	nodes[0].SetTestingFlag(t, vtcp.FlagForbidSendMessageToCoordinatorReservation, nodes[4].GetIPAddressForRequests(), "")
	nodes[1].SetTestingFlag(t, vtcp.FlagForbidSendMessageToCoordinatorReservation, "", "")
	nodes[2].SetTestingFlag(t, vtcp.FlagForbidSendMessageToCoordinatorReservation, "", "")
	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "300", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckMaxFlow(t, nodes[7], testconfig.Equivalent, "1000")
}
