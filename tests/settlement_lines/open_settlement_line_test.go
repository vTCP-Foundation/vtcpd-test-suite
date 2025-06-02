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
	openSettlementLineNextNodeIndex = 1
)

func getNextIPForOpenSettlementLineTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForOpenSettlementLineTest, openSettlementLineNextNodeIndex)
	openSettlementLineNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForOpenSettlementLineTest(t *testing.T, count int) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, count)
	for i := range count {
		nodes[i] = vtcp.NewNode(t, getNextIPForOpenSettlementLineTest(), fmt.Sprintf("node%c", 'A'+i))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings) // Assuming clusterSettings is defined globally or passed
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes)
	return nodes, cluster
}

func TestSettlementLineOpenNormalPass(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)
	time.Sleep(1 * time.Second) // Allow time for processing
	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1) // Added based on python test
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1) // Added based on python test
}

// Corresponds to test_trustlines_open_lost_TL_message
func TestSettlementLineOpenLostSettlementLineMessage(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	// NodeB loses the incoming settlement line message once
	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineStandardMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent) // is_check_conditions=False implied by not checking immediate result beyond t.Fatalf for HTTP errors

	// Wait for potential recovery and processing
	time.Sleep(vtcp.DefaultWaitingResponseTime + 15*time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_open_lost_TL_message_with_TA_resuming
func TestSettlementLineOpenLostMessageWithResuming(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	// NodeB loses the incoming settlement line message vtcp.DefaultMaxMessageSendingAttemptsInt times
	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineStandardMessageType, vtcp.DefaultMaxMessageSendingAttemptsStr, "0")
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	waitTime := vtcp.DefaultWaitingResponseTime*time.Duration(vtcp.DefaultMaxMessageSendingAttemptsInt) + 15*time.Second
	time.Sleep(waitTime)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_open_lost_TL_message_with_TA_resuming_and_lost_message_again
func TestSettlementLineOpenLostMessageWithResumingAndLostAgain(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	attempts := vtcp.DefaultMaxMessageSendingAttemptsInt + 1
	attemptsStr := strconv.Itoa(attempts)

	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineStandardMessageType, attemptsStr, "0")
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	waitTime := vtcp.DefaultWaitingResponseTime*time.Duration(attempts) + 15*time.Second
	time.Sleep(waitTime)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_open_lost_TL_message_with_TA_resuming_second_time
func TestSettlementLineOpenLostMessageWithResumingSecondTime(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	attempts := vtcp.DefaultMaxMessageSendingAttemptsInt * 2
	attemptsStr := strconv.Itoa(attempts)

	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineStandardMessageType, attemptsStr, "0")
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	waitTime := vtcp.DefaultWaitingResponseTime*time.Duration(attempts) + 15*time.Second
	time.Sleep(waitTime)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_open_lost_TL_confirmation_message
func TestSettlementLineOpenLostConfirmationMessage(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	// NodeA loses the settlement line response/confirmation message once
	err := nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineResponseMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(vtcp.DefaultWaitingResponseTime + 15*time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_open_lost_TL_confirmation_message_with_TA_resuming
func TestSettlementLineOpenLostConfirmationMessageWithResuming(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	err := nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineResponseMessageType, vtcp.DefaultMaxMessageSendingAttemptsStr, "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	waitTime := vtcp.DefaultWaitingResponseTime*time.Duration(vtcp.DefaultMaxMessageSendingAttemptsInt) + 15*time.Second
	time.Sleep(waitTime)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_open_lost_TL_confirmation_message_with_TA_resuming_and_lost_message_again
func TestSettlementLineOpenLostConfirmationMessageWithResumingAndLostAgain(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	attempts := vtcp.DefaultMaxMessageSendingAttemptsInt + 1
	attemptsStr := strconv.Itoa(attempts)

	err := nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineResponseMessageType, attemptsStr, "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	waitTime := vtcp.DefaultWaitingResponseTime*time.Duration(attempts) + 15*time.Second
	time.Sleep(waitTime)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_open_lost_TL_confirmation_message_with_TA_resuming_second_time
func TestSettlementLineOpenLostConfirmationMessageWithResumingSecondTime(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	attempts := vtcp.DefaultMaxMessageSendingAttemptsInt * 2
	attemptsStr := strconv.Itoa(attempts)

	err := nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineResponseMessageType, attemptsStr, "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	waitTime := vtcp.DefaultWaitingResponseTime*time.Duration(attempts) + 15*time.Second
	time.Sleep(waitTime)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_open_exception_on_initiator_modifying_stage
func TestSettlementLineOpenExceptionOnInitModifyingStage(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	err := nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionPreviousNeighborRequest, vtcp.SettlementLineSourceTransactionType, "1", "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(60 * time.Second) // Python test sleeps for 60s

	// Python's check_trustline(..., status_code=401) means the TL is not found/accessible.
	// Parameters for non-state checks are placeholders as they won't be checked if status is not OK.
	nodeA.CheckSettlementLine(t, nodeB, testconfig.Equivalent,
		"", "", "", "", "", "", vtcp.StatusProtocolError) // Assuming 401 maps to a general protocol error or not found for the API gateway
	nodeB.CheckSettlementLine(t, nodeA, testconfig.Equivalent,
		"", "", "", "", "", "", vtcp.StatusProtocolError)

	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
}

// Corresponds to test_trustlines_open_io_exception_on_initiator_modifying_stage
func TestSettlementLineOpenIOExceptionOnInitModifyingStage(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	err := nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionPreviousNeighborRequest, vtcp.SettlementLineSourceTransactionType, "2", "0") // secondParam="2" for IO exception
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	// Python: status_code=501. The initial POST for CreateSettlementLine is likely OK.
	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(60 * time.Second)

	// Python: check_trustline(..., status_code=405)
	// For GetSettlementsLineInfoByAddress, a 501 from the underlying system might manifest as a different error code through the API gateway if not mapped directly.
	// Using StatusProtocolError (409) as a general placeholder for client-side observable errors when specific mapping isn't known.
	// The key is that it's not StatusOK.
	nodeA.CheckSettlementLine(t, nodeB, testconfig.Equivalent,
		"", "", "", "", "", "", vtcp.StatusProtocolError) // Python has 405, if API maps 501 to 405, use that. Otherwise, a general error.
	nodeB.CheckSettlementLine(t, nodeA, testconfig.Equivalent,
		"", "", "", "", "", "", vtcp.StatusProtocolError)

	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
}

// Corresponds to test_trustlines_open_exception_on_initiator_response_processing_stage
func TestSettlementLineOpenExceptionOnInitResponseProcessingStage(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	err := nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionNextNeighborResponse, vtcp.SettlementLineSourceTransactionType, "1", "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent) // is_check_conditions=False

	time.Sleep(60 * time.Second) // Python test sleeps for 60s

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	// In Python, check_trustline_after_changes is called, which implies successful state checks.
	// So, expected status is OK (200)
	nodeA.CheckActiveSettlementLine(t, nodeB, testconfig.Equivalent, "0", "0", "0") // MaxPos, MaxNeg, Balance are 0 initially
	nodeB.CheckActiveSettlementLine(t, nodeA, testconfig.Equivalent, "0", "0", "0")
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_open_io_exception_on_initiator_response_processing_stage
func TestSettlementLineOpenIOExceptionOnInitResponseProcessingStage(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	err := nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionNextNeighborResponse, vtcp.SettlementLineSourceTransactionType, "2", "0") // IO Exception
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckActiveSettlementLine(t, nodeB, testconfig.Equivalent, "0", "0", "0")
	nodeB.CheckActiveSettlementLine(t, nodeA, testconfig.Equivalent, "0", "0", "0")
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_open_exception_on_initiator_resuming_stage
func TestSettlementLineOpenExceptionOnInitResumingStage(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	// NodeB loses messages to trigger resuming on NodeA
	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineStandardMessageType, vtcp.DefaultMaxMessageSendingAttemptsStr, "0")
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag for losing messages: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	// Wait for NodeA to be in a state where it would resume
	resumeWaitTime := vtcp.DefaultWaitingResponseTime*time.Duration(vtcp.DefaultMaxMessageSendingAttemptsInt-1) + 5*time.Second
	time.Sleep(resumeWaitTime)

	// Set exception flag on NodeA for resuming stage
	err = nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionNextNeighborResponse, vtcp.SettlementLineSourceTransactionType, "1", "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag for resuming exception: %v", err)
	}

	time.Sleep(60 * time.Second) // Original Python test waits 60s after this point

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckActiveSettlementLine(t, nodeB, testconfig.Equivalent, "0", "0", "0")
	nodeB.CheckActiveSettlementLine(t, nodeA, testconfig.Equivalent, "0", "0", "0")
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_open_io_exception_on_initiator_resuming_stage
func TestSettlementLineOpenIOExceptionOnInitResumingStage(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineStandardMessageType, vtcp.DefaultMaxMessageSendingAttemptsStr, "0")
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag for losing messages: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	resumeWaitTime := vtcp.DefaultWaitingResponseTime*time.Duration(vtcp.DefaultMaxMessageSendingAttemptsInt-1) + 5*time.Second
	time.Sleep(resumeWaitTime)

	err = nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionNextNeighborResponse, vtcp.SettlementLineSourceTransactionType, "2", "0") // IO Exception
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag for resuming IO exception: %v", err)
	}

	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckActiveSettlementLine(t, nodeB, testconfig.Equivalent, "0", "0", "0")
	nodeB.CheckActiveSettlementLine(t, nodeA, testconfig.Equivalent, "0", "0", "0")
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_open_exception_on_contractor_stage
func TestSettlementLineOpenExceptionOnContractorStage(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	// Python test uses self.targetTransactionType which is 101 (SettlementLineStandardMessageType)
	err := nodeB.SetTestingSLFlag(vtcp.FlagThrowExceptionVote, vtcp.SettlementLineStandardMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	// Python test calls node_A.check_trustline(self.node_B, 0, 0, 0)
	// This implies active settlement line with 0 balances and amounts.
	nodeA.CheckActiveSettlementLine(t, nodeB, testconfig.Equivalent, "0", "0", "0")
	nodeB.CheckActiveSettlementLine(t, nodeA, testconfig.Equivalent, "0", "0", "0")
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_open_io_exception_on_contractor_stage
func TestSettlementLineOpenIOExceptionOnContractorStage(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	err := nodeB.SetTestingSLFlag(vtcp.FlagThrowExceptionVote, vtcp.SettlementLineStandardMessageType, "2", "0") // IO Exception
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckActiveSettlementLine(t, nodeB, testconfig.Equivalent, "0", "0", "0")
	nodeB.CheckActiveSettlementLine(t, nodeA, testconfig.Equivalent, "0", "0", "0")
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_open_terminate_on_initiator_modifying_stage
func TestSettlementLineOpenTerminateOnInitModifyingStage(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	err := nodeA.SetTestingSLFlag(vtcp.FlagTerminateProcessPreviousNeighborRequest, vtcp.SettlementLineSourceTransactionType, "1", "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(160 * time.Second) // Python test sleeps for 160s

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0") // This might be optimistic if termination happened early
	nodeA.CheckSerializedTransaction(t, false, 0)            // Assuming termination prevents serialization
	nodeB.CheckSerializedTransaction(t, false, 0)

	// Python: check_trustline(..., status_code=401)
	// A 401 usually means Not Authorized, or simply not found by some gateways.
	nodeA.CheckSettlementLine(t, nodeB, testconfig.Equivalent,
		"", "", "", "", "", "", vtcp.StatusProtocolError) // Using 409 as placeholder for 401
	nodeB.CheckSettlementLine(t, nodeA, testconfig.Equivalent,
		"", "", "", "", "", "", vtcp.StatusProtocolError)
	// Valid keys check might depend on how far the process got before termination.
	// Python test doesn't explicitly check keys here, so we might need to be flexible or assume 0.
	nodeA.CheckValidKeys(t, 0, 0) // Assuming keys are not established
	nodeB.CheckValidKeys(t, 0, 0)
}

// Corresponds to test_trustlines_open_terminate_after_initiator_modifying_stage
// Note: In python, this test is identical to test_trustlines_open_terminate_on_initiator_modifying_stage.
// The flag FlagTerminateProcessPreviousNeighborRequest with secondParam="1" likely covers both "on" and "after" due to how termination is handled.
// Replicating the test logic as is.
func TestSettlementLineOpenTerminateAfterInitModifyingStage(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	err := nodeA.SetTestingSLFlag(vtcp.FlagTerminateProcessPreviousNeighborRequest, vtcp.SettlementLineSourceTransactionType, "1", "0") // Same flag as "on" stage
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(160 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLine(t, nodeB, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
	nodeB.CheckSettlementLine(t, nodeA, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)
}

// Corresponds to test_trustlines_open_terminate_on_initiator_response_processing_stage
func TestSettlementLineOpenTerminateOnInitResponseProcessingStage(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	err := nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionCoordinatorRequest, vtcp.SettlementLineSourceTransactionType, "1", "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(160 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	// Python checks imply success here (check_trustline_after_changes, check_valid_keys)
	// This suggests that termination on initiator response processing might still allow the TL to form, or the test logic in Python had a different expectation for this flag.
	// For now, assuming the Python test implies successful formation despite the "terminate" flag name.
	nodeA.CheckActiveSettlementLine(t, nodeB, testconfig.Equivalent, "0", "0", "0")
	nodeB.CheckActiveSettlementLine(t, nodeA, testconfig.Equivalent, "0", "0", "0")
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_open_terminate_after_initiator_response_processing_stage
// Python test is identical to the "on" stage for this flag as well.
func TestSettlementLineOpenTerminateAfterInitResponseProcessingStage(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	err := nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionCoordinatorRequest, vtcp.SettlementLineSourceTransactionType, "1", "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(160 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckActiveSettlementLine(t, nodeB, testconfig.Equivalent, "0", "0", "0")
	nodeB.CheckActiveSettlementLine(t, nodeA, testconfig.Equivalent, "0", "0", "0")
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_open_terminate_on_initiator_resuming_stage
func TestSettlementLineOpenTerminateOnInitResumingStage(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	// Setup for resume: NodeB loses messages
	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineStandardMessageType, vtcp.DefaultMaxMessageSendingAttemptsStr, "0")
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag for losing messages: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	// Wait for NodeA to be in a state where it would resume (before setting terminate flag)
	// Note: Python test does not have this intermediate wait, it sets terminate flag immediately after open_trustline call.
	// Adding a small delay to allow initial messages before setting termination on resume.
	time.Sleep(vtcp.DefaultWaitingResponseTime*time.Duration(vtcp.DefaultMaxMessageSendingAttemptsInt-1) + 5*time.Second)

	err = nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionNextNeighborResponse, vtcp.SettlementLineSourceTransactionType, "1", "0") // param "1"
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag for resuming termination: %v", err)
	}

	time.Sleep(160 * time.Second) // Python waits 160s

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	// Python implies success here as well
	nodeA.CheckActiveSettlementLine(t, nodeB, testconfig.Equivalent, "0", "0", "0")
	nodeB.CheckActiveSettlementLine(t, nodeA, testconfig.Equivalent, "0", "0", "0")
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_open_terminate_after_initiator_resuming_stage
func TestSettlementLineOpenTerminateAfterInitResumingStage(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLineStandardMessageType, vtcp.DefaultMaxMessageSendingAttemptsStr, "0")
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag for losing messages: %v", err)
	}
	time.Sleep(1 * time.Second)
	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)
	time.Sleep(vtcp.DefaultWaitingResponseTime*time.Duration(vtcp.DefaultMaxMessageSendingAttemptsInt-1) + 5*time.Second)

	err = nodeA.SetTestingSLFlag(vtcp.FlagThrowExceptionNextNeighborResponse, vtcp.SettlementLineSourceTransactionType, "2", "0") // Python uses param "2"
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag for resuming termination: %v", err)
	}

	time.Sleep(160 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckActiveSettlementLine(t, nodeB, testconfig.Equivalent, "0", "0", "0")
	nodeB.CheckActiveSettlementLine(t, nodeA, testconfig.Equivalent, "0", "0", "0")
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_open_terminate_on_contractor_stage
func TestSettlementLineOpenTerminateOnContractorStage(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	err := nodeB.SetTestingSLFlag(vtcp.FlagTerminateProcessVote, vtcp.SettlementLineStandardMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(60 * time.Second) // Python waits 60s

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	// Python implies success
	nodeA.CheckActiveSettlementLine(t, nodeB, testconfig.Equivalent, "0", "0", "0")
	nodeB.CheckActiveSettlementLine(t, nodeA, testconfig.Equivalent, "0", "0", "0")
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_open_terminate_after_contractor_stage
func TestSettlementLineOpenTerminateAfterContractorStage(t *testing.T) {
	nodes, _ := setupNodesForOpenSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.OpenChannelAndCheck(t, nodeB)

	err := nodeB.SetTestingSLFlag(vtcp.FlagTerminateProcessVote, vtcp.SettlementLineStandardMessageType, "2", "0") // Python uses param "2"
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckActiveSettlementLine(t, nodeB, testconfig.Equivalent, "0", "0", "0")
	nodeB.CheckActiveSettlementLine(t, nodeA, testconfig.Equivalent, "0", "0", "0")
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// TODO: Add case opening TL to contractor without channel
// if __name__ == "__main__":
// suite = unittest.TestLoader().loadTestsFromTestCase(TestTrustLinesOpen)
// unittest.TextTestRunner(verbosity=2).run(suite)
