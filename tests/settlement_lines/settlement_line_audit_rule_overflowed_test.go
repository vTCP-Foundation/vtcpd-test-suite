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
	settlementLineAuditRuleOverflowedNextNodeIndex = 1
)

func getNextIPForSettlementLineAuditRuleOverflowedTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForSettlementLineAuditRuleOverflowedTest, settlementLineAuditRuleOverflowedNextNodeIndex)
	settlementLineAuditRuleOverflowedNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForSettlementLineAuditRuleOverflowedTest(t *testing.T, count int) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, count)
	for i := range count {
		nodes[i] = vtcp.NewNode(t, getNextIPForSettlementLineAuditRuleOverflowedTest(), fmt.Sprintf("node%c", 'A'+i))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings) // Assuming clusterSettings is defined globally or passed
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes)

	nodes[0].OpenChannelAndCheck(t, nodes[1])
	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "1000")
	nodes[0].SetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "2000")

	nodes[0].CreateTransactionCheckStatus(t, nodes[1], testconfig.Equivalent, "500", vtcp.StatusOK)
	nodes[1].CreateTransactionCheckStatus(t, nodes[0], testconfig.Equivalent, "1500", vtcp.StatusOK)
	return nodes, cluster
}

func TestSettlementLineAuditRuleOverflowedNormalPass(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(3 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckCurrentAudit(t, nodeB, testconfig.Equivalent, 3)
	nodeB.CheckCurrentAudit(t, nodeA, testconfig.Equivalent, 3)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "2000")
}

func TestSettlementLineAuditRuleOverflowedAuditLostSLMessage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineSetMessageType, "1", "")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep((vtcp.WaitingAuditResponseSec + 15) * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckCurrentAudit(t, nodeB, testconfig.Equivalent, 3)
	nodeB.CheckCurrentAudit(t, nodeA, testconfig.Equivalent, 3)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "2000")
}

func TestSettlementLineAuditRuleOverflowedAuditLostSLMessageWithTAResuming(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineSetMessageType, vtcp.DefaultMaxMessageSendingAttemptsStr, "")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep((vtcp.WaitingAuditResponseSec*vtcp.DefaultMaxMessageSendingAttemptsInt + 15) * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckCurrentAudit(t, nodeB, testconfig.Equivalent, 3)
	nodeB.CheckCurrentAudit(t, nodeA, testconfig.Equivalent, 3)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "2000")
}

func TestSettlementLineAuditRuleOverflowedAuditLostSLMessageWithTAResumingAndLostMessageAgain(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineSetMessageType, "4", "")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep((vtcp.WaitingAuditResponseSec*(vtcp.DefaultMaxMessageSendingAttemptsInt+1) + 15) * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckCurrentAudit(t, nodeB, testconfig.Equivalent, 3)
	nodeB.CheckCurrentAudit(t, nodeA, testconfig.Equivalent, 3)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "2000")
}

func TestSettlementLineAuditRuleOverflowedAuditLostSLMessageWithTAResumingSecondTime(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineSetMessageType, "6", "")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep((vtcp.WaitingAuditResponseSec*vtcp.DefaultMaxMessageSendingAttemptsInt*2 + 15) * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckCurrentAudit(t, nodeB, testconfig.Equivalent, 3)
	nodeB.CheckCurrentAudit(t, nodeA, testconfig.Equivalent, 3)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "2000")
}

func TestSettlementLineAuditRuleOverflowedAuditLostConfirmationMessage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineSetAuditMessageType, "1", "")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep((vtcp.WaitingAuditResponseSec + 15) * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckCurrentAudit(t, nodeB, testconfig.Equivalent, 3)
	nodeB.CheckCurrentAudit(t, nodeA, testconfig.Equivalent, 3)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "2000")
}

func TestSettlementLineAuditRuleOverflowedAuditLostConfirmationMessageWithTAResuming(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineSetAuditMessageType, vtcp.DefaultMaxMessageSendingAttemptsStr, "")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep((vtcp.WaitingAuditResponseSec*vtcp.DefaultMaxMessageSendingAttemptsInt + 15) * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckCurrentAudit(t, nodeB, testconfig.Equivalent, 3)
	nodeB.CheckCurrentAudit(t, nodeA, testconfig.Equivalent, 3)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "2000")
}

func TestSettlementLineAuditRuleOverflowedAuditLostConfirmationMessageWithTAResumingAndLostMessageAgain(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineSetAuditMessageType, "4", "")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep((vtcp.WaitingAuditResponseSec*(vtcp.DefaultMaxMessageSendingAttemptsInt+1) + 15) * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckCurrentAudit(t, nodeB, testconfig.Equivalent, 3)
	nodeB.CheckCurrentAudit(t, nodeA, testconfig.Equivalent, 3)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "2000")
}

func TestSettlementLineAuditRuleOverflowedAuditLostConfirmationMessageWithTAResumingSecondTime(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineSetAuditMessageType, "6", "")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep((vtcp.WaitingAuditResponseSec*vtcp.DefaultMaxMessageSendingAttemptsInt*2 + 15) * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckCurrentAudit(t, nodeB, testconfig.Equivalent, 3)
	nodeB.CheckCurrentAudit(t, nodeA, testconfig.Equivalent, 3)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "2000")
}

// Exception tests
func TestSettlementLineAuditRuleOverflowedAuditExceptionOnInitiatorModifyingStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionPreviousNeighborRequest, vtcp.SettlementLineSetMessageType, "1", "")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(60 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
}

func TestSettlementLineAuditRuleOverflowedAuditIOExceptionOnInitiatorModifyingStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionPreviousNeighborRequest, vtcp.SettlementLineSetMessageType, "2", "")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(60 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckCurrentAudit(t, nodeB, testconfig.Equivalent, 3)
	nodeB.CheckCurrentAudit(t, nodeA, testconfig.Equivalent, 3)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "2000")
}

func TestSettlementLineAuditRuleOverflowedAuditExceptionOnInitiatorResponseProcessingStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionCoordinatorRequest, vtcp.SettlementLineSetMessageType, "1", "")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(60 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckCurrentAudit(t, nodeB, testconfig.Equivalent, 3)
	nodeB.CheckCurrentAudit(t, nodeA, testconfig.Equivalent, 3)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "2000")
}

func TestSettlementLineAuditRuleOverflowedAuditIOExceptionOnInitiatorResponseProcessingStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionCoordinatorRequest, vtcp.SettlementLineSetMessageType, "2", "")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(60 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckCurrentAudit(t, nodeB, testconfig.Equivalent, 3)
	nodeB.CheckCurrentAudit(t, nodeA, testconfig.Equivalent, 3)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "2000")
}

func TestSettlementLineAuditRuleOverflowedAuditExceptionOnInitiatorResumingStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionNextNeighborResponse, vtcp.SettlementLineSetMessageType, "1", "")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(60 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckCurrentAudit(t, nodeB, testconfig.Equivalent, 3)
	nodeB.CheckCurrentAudit(t, nodeA, testconfig.Equivalent, 3)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "2000")
}

func TestSettlementLineAuditRuleOverflowedAuditIOExceptionOnInitiatorResumingStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionNextNeighborResponse, vtcp.SettlementLineSetMessageType, "2", "")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(60 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckCurrentAudit(t, nodeB, testconfig.Equivalent, 3)
	nodeB.CheckCurrentAudit(t, nodeA, testconfig.Equivalent, 3)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "2000")
}

func TestSettlementLineAuditRuleOverflowedAuditExceptionOnTargetStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeB.SetTestingSLFlag(vtcp.FlagThrowExceptionVote, vtcp.SettlementLineSetAuditMessageType, "1", "")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(60 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckCurrentAudit(t, nodeB, testconfig.Equivalent, 3)
	nodeB.CheckCurrentAudit(t, nodeA, testconfig.Equivalent, 3)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "2000")
}

func TestSettlementLineAuditRuleOverflowedAuditIOExceptionOnTargetStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeB.SetTestingSLFlag(vtcp.FlagThrowExceptionVote, vtcp.SettlementLineSetAuditMessageType, "2", "")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(60 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckCurrentAudit(t, nodeB, testconfig.Equivalent, 3)
	nodeB.CheckCurrentAudit(t, nodeA, testconfig.Equivalent, 3)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "2000")
}

// Termination tests
func TestSettlementLineAuditRuleOverflowedAuditTerminateOnInitiatorModifyingStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagTerminateProcessPreviousNeighborRequest, vtcp.SettlementLineSetMessageType, "1", "")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(140 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckCurrentAudit(t, nodeB, testconfig.Equivalent, 3)
	nodeB.CheckCurrentAudit(t, nodeA, testconfig.Equivalent, 3)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "2000")
}

func TestSettlementLineAuditRuleOverflowedAuditTerminateAfterInitiatorModifyingStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagTerminateProcessPreviousNeighborRequest, vtcp.SettlementLineSetMessageType, "2", "")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(140 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckCurrentAudit(t, nodeB, testconfig.Equivalent, 3)
	nodeB.CheckCurrentAudit(t, nodeA, testconfig.Equivalent, 3)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "2000")
}

func TestSettlementLineAuditRuleOverflowedAuditTerminateOnInitiatorResponseProcessingStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagTerminateProcessCoordinatorRequest, vtcp.SettlementLineSetMessageType, "1", "")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(160 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckCurrentAudit(t, nodeB, testconfig.Equivalent, 3)
	nodeB.CheckCurrentAudit(t, nodeA, testconfig.Equivalent, 3)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "2000")
}

func TestSettlementLineAuditRuleOverflowedAuditTerminateAfterInitiatorResponseProcessingStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagTerminateProcessCoordinatorRequest, vtcp.SettlementLineSetMessageType, "2", "")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(160 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckCurrentAudit(t, nodeB, testconfig.Equivalent, 3)
	nodeB.CheckCurrentAudit(t, nodeA, testconfig.Equivalent, 3)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "2000")
}

func TestSettlementLineAuditRuleOverflowedAuditTerminateOnInitiatorResumingStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagTerminateProcessNextNeighborResponse, vtcp.SettlementLineSetMessageType, "1", "")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(160 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckCurrentAudit(t, nodeB, testconfig.Equivalent, 3)
	nodeB.CheckCurrentAudit(t, nodeA, testconfig.Equivalent, 3)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "2000")
}

func TestSettlementLineAuditRuleOverflowedAuditTerminateAfterInitiatorResumingStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagTerminateProcessNextNeighborResponse, vtcp.SettlementLineSetMessageType, "2", "")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(160 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckCurrentAudit(t, nodeB, testconfig.Equivalent, 3)
	nodeB.CheckCurrentAudit(t, nodeA, testconfig.Equivalent, 3)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "2000")
}

func TestSettlementLineAuditRuleOverflowedAuditTerminateOnContractorStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeB.SetTestingSLFlag(vtcp.FlagTerminateProcessVote, vtcp.SettlementLineSetAuditMessageType, "1", "")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(60 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckCurrentAudit(t, nodeB, testconfig.Equivalent, 3)
	nodeB.CheckCurrentAudit(t, nodeA, testconfig.Equivalent, 3)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "2000")
}

func TestSettlementLineAuditRuleOverflowedAuditTerminateAfterContractorStage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditRuleOverflowedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeB.SetTestingSLFlag(vtcp.FlagTerminateProcessVote, vtcp.SettlementLineSetAuditMessageType, "2", "")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(60 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckCurrentAudit(t, nodeB, testconfig.Equivalent, 3)
	nodeB.CheckCurrentAudit(t, nodeA, testconfig.Equivalent, 3)

	// TODO: check active settlement line
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "2000")
}
