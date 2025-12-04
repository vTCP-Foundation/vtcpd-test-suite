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
	severalPathExchangeReceiverBranchingNextNodeIndex = 1
)

func getNextIPForSeveralPathExchangeReceiverBranchingTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForSeveralPathPaymentReceiverBranchingTest, severalPathExchangeReceiverBranchingNextNodeIndex)
	severalPathExchangeReceiverBranchingNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForSeveralPathExchangeReceiverBranchingTest(t *testing.T) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, 6)
	for i := range 6 {
		nodes[i] = vtcp.NewNode(t, getNextIPForSeveralPathExchangeReceiverBranchingTest(), fmt.Sprintf("node%d", i+1))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes, false)

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

func Test1ExchangeReceiverBranchingNormalAmount(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathExchangeReceiverBranchingTest(t)

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	time.Sleep(2 * time.Second)
	// need to restart node for clearing the cache
	nodes[0].SetHopsCount(6)
	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.Equivalent}, "0")
}

func Test2aExchangeReceiverBranchingSeveralRunNextNeighborResponseProcessingStageLostMessageMiddlePathPaymentPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathExchangeReceiverBranchingTest(t)

	nodes[2].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, "", "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "800", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.Equivalent}, "200")
}

func Test2bExchangeReceiverBranchingSeveralRunNextNeighborResponseProcessingStageLostMessageMiddlePathPaymentDontPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathExchangeReceiverBranchingTest(t)

	nodes[2].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, "", "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.Equivalent}, "1000")
}

func Test3ExchangeReceiverBranchingLostMsgRunNextNeighborResponseProcessingStageOnLongWay(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathExchangeReceiverBranchingTest(t)

	nodes[4].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, "", "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.Equivalent}, "1000")
}

func Test4aExchangeReceiverBranchingLostRunPreviousNeighborRequestProcessingStageMsgFromCommonNodeToCoordinatorPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathExchangeReceiverBranchingTest(t)

	nodes[1].SetTestingFlag(vtcp.FlagForbidSendResponseToIntemediateOnReservation, nodes[0].GetIPAddressForRequests(), "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "500", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.Equivalent}, "500")
}

func Test4bExchangeReceiverBranchingLostRunPreviousNeighborRequestProcessingStageMsgFromCommonNodeToCoordinatorDontPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathExchangeReceiverBranchingTest(t)

	nodes[1].SetTestingFlag(vtcp.FlagForbidSendResponseToIntemediateOnReservation, nodes[0].GetIPAddressForRequests(), "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.Equivalent}, "1000")
}

func Test5aExchangeReceiverBranchingLostRunPreviousNeighborRequestProcessingStageMsgFromCommonNodeToCoordinatorPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathExchangeReceiverBranchingTest(t)

	nodes[1].SetTestingFlag(vtcp.FlagForbidSendResponseToIntemediateOnReservation, nodes[2].GetIPAddressForRequests(), "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "800", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.Equivalent}, "200")
}

func Test5bExchangeReceiverBranchingLostRunPreviousNeighborRequestProcessingStageMsgFromCommonNodeToCoordinatorDontPassed(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathExchangeReceiverBranchingTest(t)

	nodes[1].SetTestingFlag(vtcp.FlagForbidSendResponseToIntemediateOnReservation, nodes[2].GetIPAddressForRequests(), "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.Equivalent}, "1000")
}

func Test6ExchangeReceiverBranchingLostRunPreviousNeighborRequestProcessingStageMsgFromCommonNodeToPreviousNode(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathExchangeReceiverBranchingTest(t)

	nodes[1].SetTestingFlag(vtcp.FlagForbidSendResponseToIntemediateOnReservation, nodes[0].GetIPAddressForRequests(), "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.Equivalent}, "1000")
}

func Test7ExchangeReceiverBranchingLostRunPreviousNeighborRequestProcessingStageMsgFromReceiver(t *testing.T) {
	nodes, _ := setupNodesForSeveralPathExchangeReceiverBranchingTest(t)

	nodes[5].SetTestingFlag(vtcp.FlagForbidSendResponseToIntemediateOnReservation, "", "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "500", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.Equivalent}, "1000")
}
