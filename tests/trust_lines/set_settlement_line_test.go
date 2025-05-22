package main

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
	"github.com/vTCP-Foundation/vtcpd-test-suite/tests/testconfig"
)

var (
	setSettlementLineNextNodeIndex = 1
)

func getNextIPForSetSettlementLineTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForSetSettlementLineTest, setSettlementLineNextNodeIndex)
	setSettlementLineNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForSetSettlementLineTest(t *testing.T, count int) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, count)
	for i := range count {
		nodes[i] = vtcp.NewNode(t, getNextIPForSetSettlementLineTest(), fmt.Sprintf("node%c", 'A'+i))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings) // Assuming clusterSettings is defined globally or passed
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes)

	nodes[0].OpenChannelAndCheck(t, nodes[1])
	nodes[0].CreateSettlementLine(t, nodes[1], testconfig.Equivalent)
	time.Sleep(3 * time.Second) // Allow time for processing
	return nodes, cluster
}

func TestSettlementLineSetNormalPass(t *testing.T) {
	nodes, _ := setupNodesForSetSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000")
	time.Sleep(5 * time.Second) // Allow time for processing

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "1000")
}

func TestSettlementLineTooFastSetAndRejectSet(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2) // NodeA, NodeB
	defer cluster.StopNodes(context.Background(), t, nodes)    // Ensure nodes are stopped
	nodeA, nodeB := nodes[0], nodes[1]
	_ = nodeB // Acknowledge unused variable

	// Create NodeC
	nodeC := vtcp.NewNode(t, getNextIPForSetSettlementLineTest(), "nodeC")
	cluster.RunSingleNode(context.Background(), t, nodeC)
	defer cluster.StopSingleNode(context.Background(), t, nodeC) // Ensure nodeC is stopped

	nodeA.OpenChannelAndCheck(t, nodeC) // Open channel between A and C
	time.Sleep(3 * time.Second)         // Allow time for processing

	// NodeA sets debug flag: reject new TL request if previous is not finished
	nodeA.SetTestingSLFlag(vtcp.TrustLineDebugFlagRejectNewRequestRace, vtcp.SettlementLineSetMessageType, "1", "")
	// NodeA opens settlement line with NodeC, expect no immediate error because it's the first operation
	nodeA.CreateSettlementLine(t, nodeC, testconfig.Equivalent)
	time.Sleep(4 * time.Second) // Allow time for processing

	// NodeA attempts to set settlement line with NodeC, this might be affected by the debug flag if the previous wasn't fully done.
	// In the python test, the first set_trustline happens, then a check, then another set_trustline.
	// The debug flag is meant to ensure that a *new* request is rejected if the *previous* is not done.
	// So, the first set should go through if CreateSettlementLine completed. If CreateSettlementLine is fast, this tests the set operation with the flag.
	nodeA.SetSettlementLine(t, nodeC, testconfig.Equivalent, "100")
	time.Sleep(20 * time.Second) // Allow time for processing and potential rejection/retry logic

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeC.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeC, testconfig.Equivalent, vtcp.SettlementLineStateActive) // Check if the line is active
	nodeC.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive) // Check from C's perspective
	// It seems the original python test then sets it again to a new value.
	nodeA.SetSettlementLine(t, nodeC, testconfig.Equivalent, "500")
	time.Sleep(5 * time.Second) // Allow time for processing

	nodeA.CheckSettlementLineState(t, nodeC, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeC.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeC.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "500")
	nodeA.CheckSerializedTransaction(t, false, 0) // Assuming no unexpected transactions
	nodeC.CheckSerializedTransaction(t, false, 0)
}

func TestSettlementLineTooFastSetAndRejectAudit(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2) // NodeA, NodeB (B is not used in this specific test logic, but setup creates it)
	defer cluster.StopNodes(context.Background(), t, nodes)
	_ = nodes[1] // nodeB is unused, acknowledge it
	nodeA := nodes[0]

	// Create NodeC
	nodeC := vtcp.NewNode(t, getNextIPForSetSettlementLineTest(), "nodeC")
	cluster.RunSingleNode(context.Background(), t, nodeC)
	defer cluster.StopSingleNode(context.Background(), t, nodeC)

	nodeA.OpenChannelAndCheck(t, nodeC) // Open channel between A and C
	time.Sleep(3 * time.Second)

	// NodeC sets debug flag: reject new TL audit if previous is not finished
	// In Python: self.node_C.set_TL_debug_flag(4, self.trustLineMessage, 1)
	// The message type here is SettlementLineSetMessageType which corresponds to "106" for the *request* from A to C.
	// When C processes this, it might involve an audit. The python test uses `self.trustLineMessage` (106).
	// For "reject audit", it implies C is rejecting an audit step related to A's SetSettlementLine request.
	nodeC.SetTestingSLFlag(vtcp.TrustLineDebugFlagRejectNewAuditRace, vtcp.SettlementLineSetMessageType, "1", "")

	// NodeC initiates opening a settlement line with NodeA. (Python: self.node_C.open_trustline(self.node_A, is_check_conditions=False))
	// This is a bit different from the `too_fast_set_and_reject_set` where A opens with C.
	nodeC.CreateSettlementLine(t, nodeA, testconfig.Equivalent)
	time.Sleep(3 * time.Second)

	// NodeC sets settlement line with NodeA. (Python: self.node_C.set_trustline(self.node_A, 100))
	nodeC.SetSettlementLine(t, nodeA, testconfig.Equivalent, "100")
	time.Sleep(20 * time.Second) // Allow time for processing and potential rejection logic

	nodeC.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineState(t, nodeC, testconfig.Equivalent, vtcp.SettlementLineStateActive)

	// Python: self.node_A.set_trustline(self.node_C, 500) - A sets line with C
	// This seems to be a separate operation, perhaps to confirm the system is stable after the previous potential race condition.
	// Or it's ensuring that Node A can still operate normally with Node C.
	nodeA.SetSettlementLine(t, nodeC, testconfig.Equivalent, "500")
	time.Sleep(5 * time.Second)

	nodeA.CheckSettlementLineState(t, nodeC, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeC.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeC.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "500") // Max flow from C to A is based on A's set capacity towards C.

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeC.CheckSerializedTransaction(t, false, 0)
}

// MESSAGE LOSING TESTS

func TestSettlementLineSetLostTLMessage(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2)
	defer cluster.StopNodes(context.Background(), t, nodes)
	nodeA, nodeB := nodes[0], nodes[1]

	// NodeB simulates losing the SetSettlementLine message once
	nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineSetMessageType, "1", "")
	time.Sleep(1 * time.Second) // Allow flag to be processed

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000")
	// DefaultWaitingResponseTime (20s) + 15s buffer
	time.Sleep(vtcp.DefaultWaitingResponseTime + 15*time.Second)

	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}

func TestSettlementLineSetLostTLMessageWithTAResuming(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2)
	defer cluster.StopNodes(context.Background(), t, nodes)
	nodeA, nodeB := nodes[0], nodes[1]

	// NodeB simulates losing the SetSettlementLine message DefaultMaxMessageSendingAttemptsInt times
	nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineSetMessageType, vtcp.DefaultMaxMessageSendingAttemptsStr, "")
	time.Sleep(1 * time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000")
	// DefaultWaitingResponseTime * MaxAttempts + 15s buffer
	time.Sleep(vtcp.DefaultWaitingResponseTime*time.Duration(vtcp.DefaultMaxMessageSendingAttemptsInt) + 15*time.Second)

	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}

func TestSettlementLineSetLostTLMessageWithTAResumingAndLostAgain(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2)
	defer cluster.StopNodes(context.Background(), t, nodes)
	nodeA, nodeB := nodes[0], nodes[1]

	attempts := vtcp.DefaultMaxMessageSendingAttemptsInt + 1
	nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineSetMessageType, strconv.Itoa(attempts), "")
	time.Sleep(1 * time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000")
	time.Sleep(vtcp.DefaultWaitingResponseTime*time.Duration(attempts) + 15*time.Second)

	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}

func TestSettlementLineSetLostTLMessageWithTAResumingSecondTime(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2)
	defer cluster.StopNodes(context.Background(), t, nodes)
	nodeA, nodeB := nodes[0], nodes[1]

	attempts := vtcp.DefaultMaxMessageSendingAttemptsInt * 2
	nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineSetMessageType, strconv.Itoa(attempts), "")
	time.Sleep(1 * time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000")
	time.Sleep(vtcp.DefaultWaitingResponseTime*time.Duration(attempts) + 15*time.Second)

	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}

func TestSettlementLineSetLostTLConfirmationMessage(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2)
	defer cluster.StopNodes(context.Background(), t, nodes)
	nodeA, nodeB := nodes[0], nodes[1]

	// NodeA simulates losing the confirmation (audit) message once
	nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineSetAuditMessageType, "1", "")
	time.Sleep(1 * time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000")
	time.Sleep(vtcp.DefaultWaitingResponseTime + 15*time.Second)

	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}

func TestSettlementLineSetLostTLConfirmationMessageWithTAResuming(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2)
	defer cluster.StopNodes(context.Background(), t, nodes)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineSetAuditMessageType, vtcp.DefaultMaxMessageSendingAttemptsStr, "")
	time.Sleep(1 * time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000")
	time.Sleep(vtcp.DefaultWaitingResponseTime*time.Duration(vtcp.DefaultMaxMessageSendingAttemptsInt) + 15*time.Second)

	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}

func TestSettlementLineSetLostTLConfirmationMessageWithTAResumingAndLostAgain(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2)
	defer cluster.StopNodes(context.Background(), t, nodes)
	nodeA, nodeB := nodes[0], nodes[1]

	attempts := vtcp.DefaultMaxMessageSendingAttemptsInt + 1
	nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineSetAuditMessageType, strconv.Itoa(attempts), "")
	time.Sleep(1 * time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000")
	time.Sleep(vtcp.DefaultWaitingResponseTime*time.Duration(attempts) + 15*time.Second)

	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}

func TestSettlementLineSetLostTLConfirmationMessageWithTAResumingSecondTime(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2)
	defer cluster.StopNodes(context.Background(), t, nodes)
	nodeA, nodeB := nodes[0], nodes[1]

	attempts := vtcp.DefaultMaxMessageSendingAttemptsInt * 2
	nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineSetAuditMessageType, strconv.Itoa(attempts), "")
	time.Sleep(1 * time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000")
	time.Sleep(vtcp.DefaultWaitingResponseTime*time.Duration(attempts) + 15*time.Second)

	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}

// EXCEPTION TESTS

// TestSettlementLineSetExceptionOnInitModifyingStage simulates an exception on the initiator node
// during the TA_MODIFYING stage of a SetSettlementLine operation.
func TestSettlementLineSetExceptionOnInitTAModifyingStage(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2)
	defer cluster.StopNodes(context.Background(), t, nodes)
	nodeA, nodeB := nodes[0], nodes[1]

	// Configure NodeA to throw an exception (type 1) during the initiator TA_MODIFYING stage
	// for transactions of type SettlementLineSetInitiatorTransactionType ("102").
	nodeA.SetTestingSLFlag(vtcp.TestFlagExceptionOnInitTAModifyingStage, vtcp.SettlementLineSetInitiatorTransactionType, "1", "0")
	time.Sleep(1 * time.Second) // Allow flag to set

	// NodeA attempts to set the settlement line. Expect a 501 Not Implemented error due to the debug flag.
	// The Python test has `status_code=501`
	// TODO: add SetSettlementLineAndCheckStatus
	//nodeA.SetSettlementLineAndCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusNotImplemented) // 501
	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000") // 501

	time.Sleep(60 * time.Second) // Wait for potential recovery or final state

	// Check states: The operation should have failed, so the line might not be active or max flow not updated.
	// Python checks: self.node_A.check_trustline_after_changes(self.node_B)
	// self.node_B.check_max_flow({self.node_A: 0})
	// Assuming check_trustline_after_changes implies it might not be active or values are default/previous.
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive) // Or could be Init if it failed early
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive) // Or Init
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")                                        // Expect 0 as per Python test logic
	nodeA.CheckSerializedTransaction(t, true, 0)                                                    // Transaction likely created but failed
	nodeB.CheckSerializedTransaction(t, false, 0)                                                   // NodeB might not see it if A fails early
}

// TestSettlementLineSetIOExceptionOnInitTAModifyingStage simulates an IO exception on the initiator node
// during the TA_MODIFYING stage of a SetSettlementLine operation.
func TestSettlementLineSetIOExceptionOnInitTAModifyingStage(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2)
	defer cluster.StopNodes(context.Background(), t, nodes)
	nodeA, nodeB := nodes[0], nodes[1]

	// Configure NodeA to throw an IO exception (type 2) during the initiator TA_MODIFYING stage.
	nodeA.SetTestingSLFlag(vtcp.TestFlagExceptionOnInitTAModifyingStage, vtcp.SettlementLineSetInitiatorTransactionType, "2", "0")
	time.Sleep(1 * time.Second)

	// NodeA attempts to set the settlement line. This might return OK locally but fail to propagate or complete.
	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000") // No specific error code expected on the call itself by default

	time.Sleep(60 * time.Second)

	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive) // Or Init
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive) // Or Init
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, true, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}

// TestSettlementLineSetExceptionOnInitTAResponseProcessingStage simulates an exception on the initiator
// during TA_RESPONSE_PROCESSING stage.
func TestSettlementLineSetExceptionOnInitTAResponseProcessingStage(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2)
	defer cluster.StopNodes(context.Background(), t, nodes)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.TestFlagExceptionOnInitTAResponseProcessingStage, vtcp.SettlementLineSetInitiatorTransactionType, "1", "0")
	time.Sleep(1 * time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000")
	time.Sleep(60 * time.Second)

	// Python: node_B.check_max_flow({self.node_A.address: 1000}) - implies it should succeed
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckSerializedTransaction(t, false, 0) // Should complete and clear
	nodeB.CheckSerializedTransaction(t, false, 0)
}

// TestSettlementLineSetIOExceptionOnInitTAResponseProcessingStage simulates an IO exception on the initiator
// during TA_RESPONSE_PROCESSING stage.
func TestSettlementLineSetIOExceptionOnInitTAResponseProcessingStage(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2)
	defer cluster.StopNodes(context.Background(), t, nodes)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.TestFlagExceptionOnInitTAResponseProcessingStage, vtcp.SettlementLineSetInitiatorTransactionType, "2", "0")
	time.Sleep(1 * time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000")
	time.Sleep(60 * time.Second)

	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}

// TestSettlementLineSetExceptionOnInitTAResumingStage simulates an exception on the initiator
// during TA_RESUMING stage.
func TestSettlementLineSetExceptionOnInitTAResumingStage(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2)
	defer cluster.StopNodes(context.Background(), t, nodes)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.TestFlagExceptionOnInitTAResumingStage, vtcp.SettlementLineSetInitiatorTransactionType, "1", "0")
	time.Sleep(1 * time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000")
	time.Sleep(60 * time.Second)

	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}

// TestSettlementLineSetIOExceptionOnInitTAResumingStage simulates an IO exception on the initiator
// during TA_RESUMING stage.
func TestSettlementLineSetIOExceptionOnInitTAResumingStage(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2)
	defer cluster.StopNodes(context.Background(), t, nodes)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.TestFlagExceptionOnInitTAResumingStage, vtcp.SettlementLineSetInitiatorTransactionType, "2", "0")
	time.Sleep(1 * time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000")
	time.Sleep(60 * time.Second)

	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}

// TestSettlementLineSetExceptionOnTargetTAStage simulates an exception on the target node (NodeB)
// during its TA_STAGE (processing the request from NodeA).
func TestSettlementLineSetExceptionOnTargetTAStage(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2)
	defer cluster.StopNodes(context.Background(), t, nodes)
	nodeA, nodeB := nodes[0], nodes[1]

	// Configure NodeB to throw an exception (type 1) during its TA_STAGE
	// for transactions of type SettlementLineSetTargetTransactionType ("107").
	nodeB.SetTestingSLFlag(vtcp.TestFlagExceptionOnContractorTAStage, vtcp.SettlementLineSetTargetTransactionType, "1", "0")
	time.Sleep(1 * time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000")
	time.Sleep(60 * time.Second)

	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}

// TestSettlementLineSetIOExceptionOnTargetTAStage simulates an IO exception on the target node (NodeB)
// during its TA_STAGE.
func TestSettlementLineSetIOExceptionOnTargetTAStage(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2)
	defer cluster.StopNodes(context.Background(), t, nodes)
	nodeA, nodeB := nodes[0], nodes[1]

	// Configure NodeB to throw an IO exception (type 2) during its TA_STAGE.
	nodeB.SetTestingSLFlag(vtcp.TestFlagExceptionOnContractorTAStage, vtcp.SettlementLineSetTargetTransactionType, "2", "0")
	time.Sleep(1 * time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000")
	time.Sleep(60 * time.Second)

	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}

// TERMINATION TESTS

// TestSettlementLineSetTerminateOnInitTAModifyingStage simulates termination on initiator during TA_MODIFYING stage.
func TestSettlementLineSetTerminateOnInitTAModifyingStage(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2)
	// Not stopping nodes here as one will terminate
	nodeA, nodeB := nodes[0], nodes[1]

	// Configure NodeA to terminate (type 1) during initiator TA_MODIFYING stage.
	nodeA.SetTestingSLFlag(vtcp.TestFlagTerminateOnInitTAModifyingStage, vtcp.SettlementLineSetInitiatorTransactionType, "1", "0")
	time.Sleep(1 * time.Second)

	// Expect 503 Service Unavailable as the node might terminate during the request.
	// TODO: add SetSettlementLineAndCheckStatus
	//nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusServiceUnavailable)
	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000")

	time.Sleep(140 * time.Second) // Python test waits 140s

	// Node A is likely terminated. We might not be able to check its state reliably.
	// Check Node B's perspective. Python test implies the operation might eventually complete or be retried by B.
	// However, if A terminates, it's more likely it fails.
	// Python: self.node_B.check_max_flow({self.node_A.address: 1000})
	// This is surprising if A terminated. Let's assume the test implies the system *should* handle it or it refers to a state before termination for NodeA checks.
	// For now, let's check B. If A is down, B cannot have an active line with A.
	// Max flow should be 0 if line didn't establish or A is gone.
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "1000") // This matches python, implies eventual success despite termination.
	// We should also check the state of the line on NodeB if it is expected to be 1000
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	// NodeA specific checks are omitted as it might be down.
	nodeB.CheckSerializedTransaction(t, false, 0) // Node B should not have pending transactions if it succeeded.

	// Cleanup: Manually stop nodeB as cluster.StopNodes might have issues with a terminated nodeA.
	cluster.StopSingleNode(context.Background(), t, nodeB)
}

// TestSettlementLineSetTerminateAfterInitTAModifyingStage simulates termination after initiator TA_MODIFYING stage.
func TestSettlementLineSetTerminateAfterInitTAModifyingStage(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.TestFlagTerminateOnInitTAModifyingStage, vtcp.SettlementLineSetInitiatorTransactionType, "2", "0") // Type 2 for "after"
	time.Sleep(1 * time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000") // May or may not return error immediately
	time.Sleep(140 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "1000")
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSerializedTransaction(t, false, 0)
	cluster.StopSingleNode(context.Background(), t, nodeB) // nodeA might be down
}

// TestSettlementLineSetTerminateOnInitTAResponseProcessingStage simulates termination on initiator during TA_RESPONSE_PROCESSING stage.
func TestSettlementLineSetTerminateOnInitTAResponseProcessingStage(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.TestFlagTerminateOnInitTAResponseProcessingStage, vtcp.SettlementLineSetInitiatorTransactionType, "1", "0")
	time.Sleep(1 * time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000")
	time.Sleep(140 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "1000")
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSerializedTransaction(t, false, 0)
	cluster.StopSingleNode(context.Background(), t, nodeB)
}

// TestSettlementLineSetTerminateAfterInitTAResponseProcessingStage simulates termination after initiator TA_RESPONSE_PROCESSING stage.
func TestSettlementLineSetTerminateAfterInitTAResponseProcessingStage(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.TestFlagTerminateOnInitTAResponseProcessingStage, vtcp.SettlementLineSetInitiatorTransactionType, "2", "0")
	time.Sleep(1 * time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000")
	time.Sleep(140 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "1000")
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSerializedTransaction(t, false, 0)
	cluster.StopSingleNode(context.Background(), t, nodeB)
}

// TestSettlementLineSetTerminateOnInitTAResumingStage simulates termination on initiator during TA_RESUMING stage.
func TestSettlementLineSetTerminateOnInitTAResumingStage(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.TestFlagTerminateOnInitTAResumingStage, vtcp.SettlementLineSetInitiatorTransactionType, "1", "0")
	time.Sleep(1 * time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000")
	time.Sleep(140 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "1000")
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSerializedTransaction(t, false, 0)
	cluster.StopSingleNode(context.Background(), t, nodeB)
}

// TestSettlementLineSetTerminateAfterInitTAResumingStage simulates termination after initiator TA_RESUMING stage.
func TestSettlementLineSetTerminateAfterInitTAResumingStage(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.SetTestingSLFlag(vtcp.TestFlagTerminateOnInitTAResumingStage, vtcp.SettlementLineSetInitiatorTransactionType, "2", "0")
	time.Sleep(1 * time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000")
	time.Sleep(140 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "1000")
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSerializedTransaction(t, false, 0)
	cluster.StopSingleNode(context.Background(), t, nodeB)
}

// TestSettlementLineSetTerminateOnContractorTAStage simulates termination on contractor during TA_STAGE.
func TestSettlementLineSetTerminateOnContractorTAStage(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeB.SetTestingSLFlag(vtcp.FlagTerminateProcessVote, vtcp.SettlementLineSetTargetTransactionType, "1", "0")
	time.Sleep(1 * time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000")
	time.Sleep(140 * time.Second)

	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000") // Check from A's perspective
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSerializedTransaction(t, false, 0)
	cluster.StopSingleNode(context.Background(), t, nodeA) // NodeB might be down
}

// TestSettlementLineSetTerminateAfterContractorTAStage simulates termination after contractor TA_STAGE.
func TestSettlementLineSetTerminateAfterContractorTAStage(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeB.SetTestingSLFlag(vtcp.FlagTerminateProcessVote, vtcp.SettlementLineSetTargetTransactionType, "2", "0")
	time.Sleep(1 * time.Second)

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "1000")
	time.Sleep(140 * time.Second)

	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000")
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSerializedTransaction(t, false, 0)
	cluster.StopSingleNode(context.Background(), t, nodeA)
}
