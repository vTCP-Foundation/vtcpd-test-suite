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
	settlementLineAuditVsPaymentNextNodeIndex = 1
)

func getNextIPForSettlementLineAuditVsPaymentTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForSettlementLineAuditVsPaymentTest, settlementLineAuditVsPaymentNextNodeIndex)
	settlementLineAuditVsPaymentNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForSettlementLineAuditVsPaymentTest(t *testing.T, count int) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, count)
	for i := range count {
		nodes[i] = vtcp.NewNode(t, getNextIPForSettlementLineAuditVsPaymentTest(), fmt.Sprintf("node%c", 'A'+i))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings) // Assuming clusterSettings is defined globally or passed
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes, false)

	nodes[0].OpenChannelAndCheck(t, nodes[1])
	nodes[0].CreateSettlementLine(t, nodes[1], testconfig.Equivalent)
	time.Sleep(1 * time.Second)
	nodes[1].CheckSettlementLineState(t, nodes[0], testconfig.Equivalent, vtcp.SettlementLineStateActive)

	nodes[1].OpenChannelAndCheck(t, nodes[2])
	nodes[1].CreateSettlementLine(t, nodes[2], testconfig.Equivalent)
	time.Sleep(1 * time.Second)
	nodes[2].CheckSettlementLineState(t, nodes[1], testconfig.Equivalent, vtcp.SettlementLineStateActive)

	return nodes, cluster
}

func TestSettlementLineAuditVsPaymenttestForbidSetTrustLineWithReservations(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditVsPaymentTest(t, 3)
	nodeA, nodeB, nodeC := nodes[0], nodes[1], nodes[2]

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(2 * time.Second)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)

	nodeB.SetSettlementLine(t, nodeC, testconfig.Equivalent, "2000", vtcp.StatusOK)
	time.Sleep(2 * time.Second)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeC.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSettlementLineState(t, nodeC, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeC.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeC.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "1000")

	nodeC.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineParticipantVotesMessageType, "1", "")
	nodeC.CreateTransactionCheckStatus(t, nodeA, testconfig.Equivalent, "600", vtcp.StatusNoConsensusError)
	time.Sleep(2 * time.Second)
	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "3000", vtcp.StatusProtocolError)
}

func TestSettlementLineAuditVsPaymenttestForbidPaymentDuringAudit(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditVsPaymentTest(t, 3)
	nodeA, nodeB, nodeC := nodes[0], nodes[1], nodes[2]

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(2 * time.Second)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)

	nodeC.SetSettlementLine(t, nodeB, testconfig.Equivalent, "2000", vtcp.StatusOK)
	time.Sleep(2 * time.Second)
	nodeC.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeC.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeC, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeC.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "1000")

	nodeC.CreateTransactionCheckStatus(t, nodeA, testconfig.Equivalent, "200", vtcp.StatusNoConsensusError)

	nodeC.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineParticipantVotesMessageType, "1", "0")
	nodeC.CreateTransactionCheckStatus(t, nodeA, testconfig.Equivalent, "600", vtcp.StatusNoConsensusError)
	time.Sleep(2 * time.Second)
	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "3000", vtcp.StatusProtocolError)
	time.Sleep(2 * time.Second)
	nodeC.CreateTransactionCheckStatus(t, nodeA, testconfig.Equivalent, "600", vtcp.StatusInsufficientFunds)
	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "100", vtcp.StatusInsufficientFunds)
}

func TestSettlementLineAuditVsPaymenttestLostPaymentKeysMessage(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineAuditVsPaymentTest(t, 3)
	nodeA, nodeB, nodeC := nodes[0], nodes[1], nodes[2]

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(2 * time.Second)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)

	nodeC.SetSettlementLine(t, nodeB, testconfig.Equivalent, "2000", vtcp.StatusOK)
	time.Sleep(2 * time.Second)
	nodeC.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeC.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeC, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeC.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "1000")

	nodeC.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineParticipantVotesMessageType, "1", "0")
	nodeC.CreateTransactionCheckStatus(t, nodeA, testconfig.Equivalent, "600", vtcp.StatusOK)
	time.Sleep(2 * time.Second)
	nodeB.CreateTransactionCheckStatus(t, nodeA, testconfig.Equivalent, "100", vtcp.StatusOK)
	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "3000", vtcp.StatusOK)
	time.Sleep(2 * time.Second)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
}
