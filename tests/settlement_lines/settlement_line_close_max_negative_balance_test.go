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
	settlementLineCloseMaxNegativeBalanceNextNodeIndex = 1
)

func getNextIPForSettlementLineCloseMaxNegativeBalanceTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForSettlementLineCloseMaxNegativeBalanceTest, settlementLineCloseMaxNegativeBalanceNextNodeIndex)
	settlementLineCloseMaxNegativeBalanceNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t *testing.T, count int) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, count)
	for i := range count {
		nodes[i] = vtcp.NewNode(t, getNextIPForSettlementLineCloseMaxNegativeBalanceTest(), fmt.Sprintf("node%c", 'A'+i))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings) // Assuming clusterSettings is defined globally or passed
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes)

	nodes[0].OpenChannelAndCheck(t, nodes[1])
	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "1000")
	nodes[0].SetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "500")

	nodes[0].CheckMaxFlow(t, nodes[1], testconfig.Equivalent, "500")
	return nodes, cluster
}

func TestSettlementLineCloseMaxNegativeBalanceNormalPass(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(1 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

func TestSettlementLineCloseMaxNegativeBalanceLostSLMessage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineSetMessageType, "1", "")

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep((vtcp.WaitingAuditResponseSec + 15) * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

func TestSettlementLineCloseMaxNegativeBalanceLostSLMessageWithTAResuming(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineSetMessageType, vtcp.DefaultMaxMessageSendingAttemptsStr, "")

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(time.Duration(vtcp.WaitingAuditResponseSec*vtcp.DefaultMaxMessageSendingAttemptsInt+15) * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

func TestSettlementLineCloseMaxNegativeBalanceLostSLMessageWithTAResumingAndLostMessageAgain(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineSetMessageType, "4", "")

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(time.Duration(vtcp.WaitingAuditResponseSec*4+15) * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

func TestSettlementLineCloseMaxNegativeBalanceLostSLMessageWithTAResumingSecondTime(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineSetMessageType, "6", "")

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(time.Duration(vtcp.WaitingAuditResponseSec*6+15) * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

func TestSettlementLineCloseMaxNegativeBalanceLostSLConfirmationMessage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineSetAuditMessageType, "1", "")

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep((vtcp.WaitingAuditResponseSec + 15) * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

func TestSettlementLineCloseMaxNegativeBalanceLostSLConfirmationMessageWithTAResuming(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineSetAuditMessageType, vtcp.DefaultMaxMessageSendingAttemptsStr, "")

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(time.Duration(vtcp.WaitingAuditResponseSec*vtcp.DefaultMaxMessageSendingAttemptsInt+15) * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

func TestSettlementLineCloseMaxNegativeBalanceLostSLConfirmationMessageWithTAResumingAndLostMessageAgain(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineSetAuditMessageType, "4", "")

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(time.Duration(vtcp.WaitingAuditResponseSec*4+15) * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

func TestSettlementLineCloseMaxNegativeBalanceLostSLConfirmationMessageWithTAResumingSecondTime(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineSetAuditMessageType, "6", "")

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(time.Duration(vtcp.WaitingAuditResponseSec*6+15) * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

// Exception tests
func TestSettlementLineCloseMaxNegativeBalanceExceptionOnInitiatorModifyingStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionPreviousNeighborRequest, vtcp.SettlementLinePublicKeyInitMessageType, "1", "")

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(60 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

func TestSettlementLineCloseMaxNegativeBalanceIOExceptionOnInitiatorModifyingStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionPreviousNeighborRequest, vtcp.SettlementLinePublicKeyInitMessageType, "2", "")

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(60 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

func TestSettlementLineCloseMaxNegativeBalanceExceptionOnInitiatorResponseProcessingStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionCoordinatorRequest, vtcp.SettlementLinePublicKeyInitMessageType, "1", "")

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(60 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

func TestSettlementLineCloseMaxNegativeBalanceIOExceptionOnInitiatorResponseProcessingStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionCoordinatorRequest, vtcp.SettlementLinePublicKeyInitMessageType, "2", "")

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(60 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

func TestSettlementLineCloseMaxNegativeBalanceExceptionOnInitiatorResumingStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionNextNeighborResponse, vtcp.SettlementLinePublicKeyInitMessageType, "1", "")

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(60 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

func TestSettlementLineCloseMaxNegativeBalanceIOExceptionOnInitiatorResumingStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionNextNeighborResponse, vtcp.SettlementLinePublicKeyInitMessageType, "2", "")

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(60 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

func TestSettlementLineCloseMaxNegativeBalanceExceptionOnTargetStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeB.SetTestingSLFlag(vtcp.FlagThrowExceptionVote, vtcp.SettlementLineSetTargetTransactionType, "1", "")

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(60 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

func TestSettlementLineCloseMaxNegativeBalanceIOExceptionOnTargetStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeB.SetTestingSLFlag(vtcp.FlagThrowExceptionVote, vtcp.SettlementLineSetTargetTransactionType, "2", "")

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(60 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

// Termination tests
func TestSettlementLineCloseMaxNegativeBalanceTerminateOnInitiatorModifyingStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagTerminateProcessPreviousNeighborRequest, vtcp.SettlementLinePublicKeyInitMessageType, "1", "")

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(140 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

func TestSettlementLineCloseMaxNegativeBalanceTerminateAfterInitiatorModifyingStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagTerminateProcessPreviousNeighborRequest, vtcp.SettlementLinePublicKeyInitMessageType, "2", "")

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(140 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

func TestSettlementLineCloseMaxNegativeBalanceTerminateOnInitiatorResponseProcessingStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagTerminateProcessCoordinatorRequest, vtcp.SettlementLinePublicKeyInitMessageType, "1", "")

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(160 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

func TestSettlementLineCloseMaxNegativeBalanceTerminateAfterInitiatorResponseProcessingStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagTerminateProcessCoordinatorRequest, vtcp.SettlementLinePublicKeyInitMessageType, "2", "")

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(160 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

func TestSettlementLineCloseMaxNegativeBalanceTerminateOnInitiatorResumingStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagTerminateProcessNextNeighborResponse, vtcp.SettlementLinePublicKeyInitMessageType, "1", "")

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(160 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

func TestSettlementLineCloseMaxNegativeBalanceTerminateAfterInitiatorResumingStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagTerminateProcessNextNeighborResponse, vtcp.SettlementLinePublicKeyInitMessageType, "2", "")

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(160 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

func TestSettlementLineCloseMaxNegativeBalanceTerminateOnContractorStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeB.SetTestingSLFlag(vtcp.FlagTerminateProcessVote, vtcp.SettlementLineSetTargetTransactionType, "1", "")

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(60 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

func TestSettlementLineCloseMaxNegativeBalanceTerminateAfterContractorStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineCloseMaxNegativeBalanceTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeB.SetTestingSLFlag(vtcp.FlagTerminateProcessVote, vtcp.SettlementLineSetTargetTransactionType, "2", "")

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(60 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}
