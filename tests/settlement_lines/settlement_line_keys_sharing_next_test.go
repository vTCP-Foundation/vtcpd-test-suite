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
	keysSharingNextSettlementLineNextNodeIndex = 1
)

func getNextIPForKeysSharingNextSettlementLineTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForKeysSharingNextSettlementLineTest, keysSharingNextSettlementLineNextNodeIndex)
	keysSharingNextSettlementLineNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForKeysSharingNextSettlementLineTest(t *testing.T, count int) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, count)
	for i := range count {
		nodes[i] = vtcp.NewNode(t, getNextIPForKeysSharingNextSettlementLineTest(), fmt.Sprintf("node%c", 'A'+i))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings) // Assuming clusterSettings is defined globally or passed
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes)

	nodes[0].OpenChannelAndCheck(t, nodes[1])
	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "2000")
	nodes[0].SetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "1000")

	for x := 0; x < vtcp.DefaultKeysCount-vtcp.DefaultCriticalKeysCount-4; x++ {
		nodes[0].CreateTransactionCheckStatus(t, nodes[1], testconfig.Equivalent, "50", vtcp.StatusOK)
		time.Sleep(3 * time.Second)
	}

	return nodes, cluster
}

func TestSettlementLineKeysSharingByPaymentOnContractor(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "50", vtcp.StatusOK)
	time.Sleep(vtcp.DefaultKeysSharingSeconds * time.Second) // Allow time for processing

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineKeysPresence(t, nodeB, true, true)
	nodeB.CheckSettlementLineKeysPresence(t, nodeA, true, true)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByPaymentOnIntermediateNode(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 3)
	nodeA, nodeB, nodeC := nodes[0], nodes[1], nodes[2]

	nodeA.OpenChannelAndCheck(t, nodeC)
	nodeA.CreateAndSetSettlementLineAndCheck(t, nodeC, testconfig.Equivalent, "1000")

	nodeC.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "50", vtcp.StatusOK)
	time.Sleep(vtcp.DefaultKeysSharingSeconds * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeC.CheckSerializedTransaction(t, false, 0)

	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineState(t, nodeC, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeC.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)

	nodeA.CheckSettlementLineKeysPresence(t, nodeB, true, true)
	nodeB.CheckSettlementLineKeysPresence(t, nodeA, true, true)
	nodeC.CheckSettlementLineKeysPresence(t, nodeA, true, true)
	nodeA.CheckSettlementLineKeysPresence(t, nodeC, true, true)

	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
	nodeC.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByModifyingAsInitiator(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "500", vtcp.StatusOK)
	time.Sleep(vtcp.DefaultKeysSharingSeconds * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineKeysPresence(t, nodeB, true, true)
	nodeB.CheckSettlementLineKeysPresence(t, nodeA, true, true)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByModifyingAsContractor(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeB.SetSettlementLine(t, nodeA, testconfig.Equivalent, "500", vtcp.StatusOK)
	time.Sleep(vtcp.DefaultKeysSharingSeconds * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineKeysPresence(t, nodeB, true, true)
	nodeB.CheckSettlementLineKeysPresence(t, nodeA, true, true)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByClosingIncomingAsInitiator(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(vtcp.DefaultKeysSharingSeconds * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineKeysPresence(t, nodeB, true, true)
	nodeB.CheckSettlementLineKeysPresence(t, nodeA, true, true)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByClosingIncomingAsContractor(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeB.CloseMaxNegativeBalance(t, nodeA, testconfig.Equivalent)
	time.Sleep(vtcp.DefaultKeysSharingSeconds * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineKeysPresence(t, nodeB, true, true)
	nodeB.CheckSettlementLineKeysPresence(t, nodeA, true, true)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByModificationLostInitKeyMessage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyInitMessageType, "1", "0")
	if err != nil {
		t.Fatalf("failed to set testing flag: %v", err)
	}
	time.Sleep(time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(vtcp.DefaultKeysSharingWaitingResponseTime*time.Second + vtcp.DefaultKeysSharingSeconds*time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineKeysPresence(t, nodeB, true, true)
	nodeB.CheckSettlementLineKeysPresence(t, nodeA, true, true)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByClosingLostInitKeyMessage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyInitMessageType, "1", "0")
	if err != nil {
		t.Fatalf("failed to set testing flag: %v", err)
	}
	time.Sleep(time.Second)

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(vtcp.DefaultKeysSharingWaitingResponseTime*time.Second + vtcp.DefaultKeysSharingSeconds*time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineKeysPresence(t, nodeB, true, true)
	nodeB.CheckSettlementLineKeysPresence(t, nodeA, true, true)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByPaymentLostInitKeyMessage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyInitMessageType, "1", "0")
	if err != nil {
		t.Fatalf("failed to set testing flag: %v", err)
	}
	time.Sleep(time.Second)

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "50", vtcp.StatusOK)
	time.Sleep(vtcp.DefaultKeysSharingWaitingResponseTime*time.Second + vtcp.DefaultKeysSharingSeconds*time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineKeysPresence(t, nodeB, true, true)
	nodeB.CheckSettlementLineKeysPresence(t, nodeA, true, true)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByModificationLostInitKeyMessageWithTASleeping(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyInitMessageType, vtcp.DefaultMaxMessageSendingAttemptsStr, "0")
	if err != nil {
		t.Fatalf("failed to set testing flag: %v", err)
	}
	time.Sleep(time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(time.Duration(vtcp.DefaultMaxMessageSendingAttemptsInt)*vtcp.DefaultKeysSharingWaitingResponseTime*time.Second + vtcp.DefaultKeysSharingSeconds*time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineKeysPresence(t, nodeB, false, true)
	nodeB.CheckSettlementLineKeysPresence(t, nodeA, true, true)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByClosingLostInitKeyMessageWithTASleeping(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyInitMessageType, vtcp.DefaultMaxMessageSendingAttemptsStr, "0")
	if err != nil {
		t.Fatalf("failed to set testing flag: %v", err)
	}
	time.Sleep(time.Second)

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(time.Duration(vtcp.DefaultMaxMessageSendingAttemptsInt)*vtcp.DefaultKeysSharingWaitingResponseTime*time.Second + vtcp.DefaultKeysSharingSeconds*time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineKeysPresence(t, nodeB, false, true)
	nodeB.CheckSettlementLineKeysPresence(t, nodeA, true, true)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByPaymentLostInitKeyMessageWithTASleeping(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyInitMessageType, vtcp.DefaultMaxMessageSendingAttemptsStr, "0")
	if err != nil {
		t.Fatalf("failed to set testing flag: %v", err)
	}
	time.Sleep(time.Second)

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "50", vtcp.StatusOK)
	time.Sleep(time.Duration(vtcp.DefaultMaxMessageSendingAttemptsInt)*vtcp.DefaultKeysSharingWaitingResponseTime*time.Second + vtcp.DefaultKeysSharingSeconds*time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineKeysPresence(t, nodeB, false, true)
	nodeB.CheckSettlementLineKeysPresence(t, nodeA, true, true)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByModificationLostKeyMessage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyMessageType, "1", "0")
	if err != nil {
		t.Fatalf("failed to set testing flag: %v", err)
	}
	time.Sleep(time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(vtcp.DefaultKeysSharingWaitingResponseTime*time.Second + vtcp.DefaultKeysSharingSeconds*time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineKeysPresence(t, nodeB, true, true)
	nodeB.CheckSettlementLineKeysPresence(t, nodeA, true, true)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByClosingLostKeyMessage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyMessageType, "1", "0")
	if err != nil {
		t.Fatalf("failed to set testing flag: %v", err)
	}
	time.Sleep(time.Second)

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(vtcp.DefaultKeysSharingWaitingResponseTime*time.Second + vtcp.DefaultKeysSharingSeconds*time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineKeysPresence(t, nodeB, true, true)
	nodeB.CheckSettlementLineKeysPresence(t, nodeA, true, true)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByPaymentLostKeyMessage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyMessageType, "1", "0")
	if err != nil {
		t.Fatalf("failed to set testing flag: %v", err)
	}
	time.Sleep(time.Second)

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "50", vtcp.StatusOK)
	time.Sleep(vtcp.DefaultKeysSharingWaitingResponseTime*time.Second + vtcp.DefaultKeysSharingSeconds*time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineKeysPresence(t, nodeB, true, true)
	nodeB.CheckSettlementLineKeysPresence(t, nodeA, true, true)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByModificationLostKeyMessageWithTASleeping(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyMessageType, vtcp.DefaultMaxMessageSendingAttemptsStr, "0")
	if err != nil {
		t.Fatalf("failed to set testing flag: %v", err)
	}
	time.Sleep(time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(time.Duration(vtcp.DefaultMaxMessageSendingAttemptsInt)*vtcp.DefaultKeysSharingWaitingResponseTime*time.Second + vtcp.DefaultKeysSharingSeconds*time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineKeysPresence(t, nodeB, false, true)
	nodeB.CheckSettlementLineKeysPresence(t, nodeA, true, false)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByClosingLostKeyMessageWithTASleeping(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyMessageType, vtcp.DefaultMaxMessageSendingAttemptsStr, "0")
	if err != nil {
		t.Fatalf("failed to set testing flag: %v", err)
	}
	time.Sleep(time.Second)

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(time.Duration(vtcp.DefaultMaxMessageSendingAttemptsInt)*vtcp.DefaultKeysSharingWaitingResponseTime*time.Second + vtcp.DefaultKeysSharingSeconds*time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineKeysPresence(t, nodeB, false, true)
	nodeB.CheckSettlementLineKeysPresence(t, nodeA, true, true)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByModificationLostHashConfirmationMessage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyResponseMessageType, "1", "0")
	if err != nil {
		t.Fatalf("failed to set testing flag: %v", err)
	}
	time.Sleep(time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(vtcp.DefaultKeysSharingWaitingResponseTime*time.Second + vtcp.DefaultKeysSharingSeconds*time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineKeysPresence(t, nodeB, true, true)
	nodeB.CheckSettlementLineKeysPresence(t, nodeA, true, true)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByClosingLostHashConfirmationMessage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyResponseMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(vtcp.DefaultKeysSharingWaitingResponseTime*time.Second + vtcp.DefaultKeysSharingSeconds*time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineKeysPresence(t, nodeB, true, true)
	nodeB.CheckSettlementLineKeysPresence(t, nodeA, true, true)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByPaymentLostHashConfirmationMessage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyResponseMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "50", vtcp.StatusOK)
	time.Sleep(vtcp.DefaultKeysSharingWaitingResponseTime*time.Second + vtcp.DefaultKeysSharingSeconds*time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineKeysPresence(t, nodeB, true, true)
	nodeB.CheckSettlementLineKeysPresence(t, nodeA, true, true)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByClosingLostHashConfirmationMessageWithTASleeping(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyResponseMessageType, vtcp.DefaultMaxMessageSendingAttemptsStr, "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(time.Duration(vtcp.DefaultMaxMessageSendingAttemptsInt)*vtcp.DefaultKeysSharingWaitingResponseTime*time.Second + vtcp.DefaultKeysSharingSeconds*time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineKeysPresence(t, nodeB, false, true)
	nodeB.CheckSettlementLineKeysPresence(t, nodeA, true, true)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByPaymentLostHashConfirmationMessageWithTASleeping(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyResponseMessageType, vtcp.DefaultMaxMessageSendingAttemptsStr, "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "50", vtcp.StatusOK)
	time.Sleep(time.Duration(vtcp.DefaultMaxMessageSendingAttemptsInt)*vtcp.DefaultKeysSharingWaitingResponseTime*time.Second + vtcp.DefaultKeysSharingSeconds*time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineKeysPresence(t, nodeB, false, true)
	nodeB.CheckSettlementLineKeysPresence(t, nodeA, true, true)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByModificationIOExceptionOnInitiatorSendFirstKey(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionPreviousNeighborRequest, vtcp.SettlementLinePublicKeyMessageType, "2", "0") // IO Exception
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSettlementLine(t, nodeB, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
	nodeB.CheckSettlementLine(t, nodeA, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
}

func TestSettlementLineKeysSharingByModificationExceptionOnInitiatorSendFirstKey(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionPreviousNeighborRequest, vtcp.SettlementLinePublicKeyMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSettlementLine(t, nodeB, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
	nodeB.CheckSettlementLine(t, nodeA, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
}

func TestSettlementLineKeysSharingByClosingIOExceptionOnInitiatorSendFirstKey(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionPreviousNeighborRequest, vtcp.SettlementLinePublicKeyMessageType, "2", "0") // IO Exception
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSettlementLine(t, nodeB, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
	nodeB.CheckSettlementLine(t, nodeA, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
}

func TestSettlementLineKeysSharingByClosingExceptionOnInitiatorSendFirstKey(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionPreviousNeighborRequest, vtcp.SettlementLinePublicKeyMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSettlementLine(t, nodeB, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
	nodeB.CheckSettlementLine(t, nodeA, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
}

func TestSettlementLineKeysSharingByPaymentIOExceptionOnInitiatorSendFirstKey(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionPreviousNeighborRequest, vtcp.SettlementLinePublicKeyMessageType, "2", "0") // IO Exception
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "50", vtcp.StatusProtocolError)
	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSettlementLine(t, nodeB, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
	nodeB.CheckSettlementLine(t, nodeA, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
}

func TestSettlementLineKeysSharingByPaymentExceptionOnInitiatorSendFirstKey(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionPreviousNeighborRequest, vtcp.SettlementLinePublicKeyMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "50", vtcp.StatusProtocolError)
	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSettlementLine(t, nodeB, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
	nodeB.CheckSettlementLine(t, nodeA, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
}

func TestSettlementLineKeysSharingByModificationIOExceptionOnContractorReceiveKey(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeB.SetTestingSLFlag(vtcp.FlagThrowExceptionVote, vtcp.SettlementLinePublicKeyResponseMessageType, "2", "0") // IO Exception
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "50", vtcp.StatusProtocolError)
	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSettlementLine(t, nodeB, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
	nodeB.CheckSettlementLine(t, nodeA, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
}

func TestSettlementLineKeysSharingByModificationExceptionOnContractorReceiveKey(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeB.SetTestingSLFlag(vtcp.FlagThrowExceptionVote, vtcp.SettlementLinePublicKeyResponseMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "50", vtcp.StatusProtocolError)
	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSettlementLine(t, nodeB, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
	nodeB.CheckSettlementLine(t, nodeA, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
}

func TestSettlementLineKeysSharingByClosingIOExceptionOnContractorReceiveKey(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeB.SetTestingSLFlag(vtcp.FlagThrowExceptionVote, vtcp.SettlementLinePublicKeyResponseMessageType, "2", "0") // IO Exception
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSettlementLine(t, nodeB, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
	nodeB.CheckSettlementLine(t, nodeA, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
}

func TestSettlementLineKeysSharingByClosingExceptionOnContractorReceiveKey(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeB.SetTestingSLFlag(vtcp.FlagThrowExceptionVote, vtcp.SettlementLinePublicKeyResponseMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSettlementLine(t, nodeB, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
	nodeB.CheckSettlementLine(t, nodeA, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
}

func TestSettlementLineKeysSharingByModificationTerminateOnInitiatorSendNextKey(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeA.SetTestingSLFlag(vtcp.FlagTerminateProcessPreviousNeighborRequest, vtcp.SettlementLinePublicKeyMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(160 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByModificationTerminateAfterInitiatorSendNextKey(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeA.SetTestingSLFlag(vtcp.FlagTerminateProcessPreviousNeighborRequest, vtcp.SettlementLinePublicKeyMessageType, "2", "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	time.Sleep(160 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByClosingTerminateOnInitiatorSendNextKey(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeA.SetTestingSLFlag(vtcp.FlagTerminateProcessPreviousNeighborRequest, vtcp.SettlementLinePublicKeyMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(160 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByClosingTerminateAfterInitiatorSendNextKey(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeA.SetTestingSLFlag(vtcp.FlagTerminateProcessPreviousNeighborRequest, vtcp.SettlementLinePublicKeyMessageType, "2", "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(160 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}

func TestSettlementLineKeysSharingByPaymentTerminateOnInitiatorSendNextKey(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeA.SetTestingSLFlag(vtcp.FlagTerminateProcessPreviousNeighborRequest, vtcp.SettlementLinePublicKeyMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "50", vtcp.StatusOK)
	time.Sleep(160 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t)
	nodeB.CheckValidKeys(t)
}
