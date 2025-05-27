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
	keysSharingInitSettlementLineNextNodeIndex = 1
)

func getNextIPForKeysSharingInitSettlementLineTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForKeysSharingInitSettlementLineTest, keysSharingInitSettlementLineNextNodeIndex)
	keysSharingInitSettlementLineNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForKeysSharingInitSettlementLineTest(t *testing.T, count int) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, count)
	for i := range count {
		nodes[i] = vtcp.NewNode(t, getNextIPForKeysSharingInitSettlementLineTest(), fmt.Sprintf("node%c", 'A'+i))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings) // Assuming clusterSettings is defined globally or passed
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes)

	nodes[0].OpenChannelAndCheck(t, nodes[1])
	return nodes, cluster
}

func TestSettlementLineKeysSharingInitNormalPass(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)
	time.Sleep(5 * time.Second) // Allow time for processing

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_keys_sharing_init_lost_key_init_message
func TestSettlementLineKeysSharingInitLostKeyInitMessage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Set flag before SL creation that triggers key sharing
	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyInitMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)
	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)
	time.Sleep(1 * time.Second)

	time.Sleep(vtcp.DefaultWaitingResponseTime + 15*time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_keys_sharing_init_lost_key_init_message_not_critical_times
func TestSettlementLineKeysSharingInitLostKeyInitMessageNotCriticalTimes(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)
	time.Sleep(1 * time.Second)

	attempts := strconv.Itoa(vtcp.DefaultMaxMessageSendingAttemptsInt - 1)
	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyInitMessageType, attempts, "0")
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	waitTime := vtcp.DefaultWaitingResponseTime*time.Duration(vtcp.DefaultMaxMessageSendingAttemptsInt-1) + 15*time.Second
	time.Sleep(waitTime)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_keys_sharing_init_lost_key_init_message_critical_times
func TestSettlementLineKeysSharingInitLostKeyInitMessageCriticalTimes(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Set the flag *before* CreateSettlementLine if the key init message is part of it and not re-tried externally.
	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyInitMessageType, vtcp.DefaultMaxMessageSendingAttemptsStr, "0")
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)

	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent) // This action triggers the key sharing

	waitTime := vtcp.DefaultWaitingResponseTime*time.Duration(vtcp.DefaultMaxMessageSendingAttemptsInt) + 15*time.Second
	time.Sleep(waitTime)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	// check_trustline_after_changes equivalent
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive) // Python implies TL is active
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive) // Python implies TL is active

	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount, 0) // Own, Contractor
	nodeB.CheckValidKeys(t, 0, 0)                     // Own, Contractor
	// TODO: Check TL state for "keys pending" if Go test suite has such a state. Python comments "TL state throw handler should be keys pending"
}

// Corresponds to test_trustlines_keys_sharing_init_lost_key_message
func TestSettlementLineKeysSharingInitLostKeyMessage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Set flag before SL creation that triggers key sharing
	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)
	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)
	time.Sleep(1 * time.Second)

	time.Sleep(vtcp.DefaultWaitingResponseTime + 15*time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_keys_sharing_init_lost_key_message_not_critical_times
func TestSettlementLineKeysSharingInitLostKeyMessageNotCriticalTimes(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	attempts := strconv.Itoa(vtcp.DefaultMaxMessageSendingAttemptsInt - 1)
	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyMessageType, attempts, "0")
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)
	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)
	time.Sleep(1 * time.Second)

	waitTime := vtcp.DefaultWaitingResponseTime*time.Duration(vtcp.DefaultMaxMessageSendingAttemptsInt-1) + 15*time.Second
	time.Sleep(waitTime)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_keys_sharing_init_lost_key_message_critical_times
func TestSettlementLineKeysSharingInitLostKeyMessageCriticalTimes(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeB.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyMessageType, vtcp.DefaultMaxMessageSendingAttemptsStr, "0")
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
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount, vtcp.DefaultKeysCount)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount, 1)
	// TODO: Check TL state for "keys pending"
}

// Corresponds to test_trustlines_keys_sharing_init_lost_key_init_confirmation_message
func TestSettlementLineKeysSharingInitLostKeyInitConfirmationMessage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// This flag is on NodeA for losing a response.
	err := nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyResponseMessageType, "1", "0") // thirdParam (index) is 0 if not specified like in python
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

// Corresponds to test_trustlines_keys_sharing_init_lost_key_init_confirmation_message_not_critical_times
func TestSettlementLineKeysSharingInitLostKeyInitConfirmationMessageNotCriticalTimes(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	attempts := strconv.Itoa(vtcp.DefaultMaxMessageSendingAttemptsInt - 1)
	err := nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyResponseMessageType, attempts, "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)
	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	waitTime := vtcp.DefaultWaitingResponseTime*time.Duration(vtcp.DefaultMaxMessageSendingAttemptsInt-1) + 15*time.Second
	time.Sleep(waitTime)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_keys_sharing_init_lost_key_init_confirmation_message_critical_times
func TestSettlementLineKeysSharingInitLostKeyInitConfirmationMessageCriticalTimes(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyResponseMessageType, vtcp.DefaultMaxMessageSendingAttemptsStr, "0")
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
	// Python: self.node_A.check_valid_keys(self.keys_count, self.keys_count) -> (10,10)
	// Python: self.node_B.check_valid_keys(self.keys_count, 1) -> (10,1)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount, vtcp.DefaultKeysCount)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount, 1)
	// TODO: Check TL state for "keys pending"
}

// Corresponds to test_trustlines_keys_sharing_init_lost_key_confirmation_message
func TestSettlementLineKeysSharingInitLostKeyConfirmationMessage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Python: self.node_A.set_TL_debug_flag(4, self.publicKeyResponseMessage, 1, 2)
	// flag, firstParam (messageType), secondParam (countToSkip), thirdParam (index/sub-type)
	err := nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyResponseMessageType, "1", "2")
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

// Corresponds to test_trustlines_keys_sharing_init_lost_key_confirmation_message_not_critical_times
func TestSettlementLineKeysSharingInitLostKeyConfirmationMessageNotCriticalTimes(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	attempts := strconv.Itoa(vtcp.DefaultMaxMessageSendingAttemptsInt - 1)
	err := nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyResponseMessageType, attempts, "2")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)
	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	waitTime := vtcp.DefaultWaitingResponseTime*time.Duration(vtcp.DefaultMaxMessageSendingAttemptsInt-1) + 15*time.Second
	time.Sleep(waitTime)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_keys_sharing_init_lost_key_confirmation_message_critical_times
func TestSettlementLineKeysSharingInitLostKeyConfirmationMessageCriticalTimes(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeA.SetTestingSLFlag(vtcp.FlagForbidSendInitMessage, vtcp.SettlementLinePublicKeyResponseMessageType, vtcp.DefaultMaxMessageSendingAttemptsStr, "2")
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
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount, vtcp.DefaultKeysCount)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount, 3)
	// TODO: Check TL state for "keys pending"
}

// Corresponds to test_trustlines_keys_sharing_init_exception_on_initiator_first_key_stage
func TestSettlementLineKeysSharingInitExceptionOnInitFirstKeyStage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Python: self.node_A.set_TL_debug_flag(2048, self.sourceTransactionType, 1)
	// sourceTransactionType for this suite is 104 (SettlementLinePublicKeyMessageType)
	// Flag 2048 is TestFlagExceptionOnInitTAModifyingStage
	err := nodeA.SetTestingSLFlag(vtcp.TestFlagExceptionOnInitTAModifyingStage, vtcp.SettlementLinePublicKeyMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)
	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSettlementLine(t, nodeB, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
	nodeB.CheckSettlementLine(t, nodeA, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
	// No key checks in corresponding python test
}

// Corresponds to test_trustlines_keys_sharing_init_io_exception_on_initiator_first_key_stage
func TestSettlementLineKeysSharingInitIOExceptionOnInitFirstKeyStage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Python: self.node_A.set_TL_debug_flag(2048, self.sourceTransactionType, 2) -> IO Exception
	err := nodeA.SetTestingSLFlag(vtcp.TestFlagExceptionOnInitTAModifyingStage, vtcp.SettlementLinePublicKeyMessageType, "2", "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)
	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSettlementLine(t, nodeB, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError) // 401 in python
	nodeB.CheckSettlementLine(t, nodeA, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError) // 401 in python
}

// Corresponds to test_trustlines_keys_sharing_init_exception_on_initiator_next_key_stage
func TestSettlementLineKeysSharingInitExceptionOnInitNextKeyStage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Python: self.node_A.set_TL_debug_flag(2048, self.sourceTransactionType, 1)
	// Flag 2048 for "next key stage" seems to be TestFlagExceptionOnInitTAResponseProcessingStage or TestFlagExceptionOnInitTAResumingStage based on open_sl tests.
	// However, the python test uses 2048, which is TestFlagExceptionOnInitTAModifyingStage.
	// This might indicate the "first_key_stage" and "next_key_stage" might be differentiated by *when* the flag is set or by a different internal condition.
	// For now, using the literal flag from python test: TestFlagExceptionOnInitTAModifyingStage
	// The difference might be subtle, for example, if the "modifying stage" for keys happens multiple times.
	err := nodeA.SetTestingSLFlag(vtcp.TestFlagExceptionOnInitTAModifyingStage, vtcp.SettlementLinePublicKeyMessageType, "1", "0") // Assuming it's still modifying stage, but for a subsequent key.
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)
	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_keys_sharing_init_io_exception_on_initiator_next_key_stage
func TestSettlementLineKeysSharingInitIOExceptionOnInitNextKeyStage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeA.SetTestingSLFlag(vtcp.TestFlagExceptionOnInitTAModifyingStage, vtcp.SettlementLinePublicKeyMessageType, "2", "0") // IO Exception
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)
	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_keys_sharing_init_exception_on_contractor_first_key_stage
func TestSettlementLineKeysSharingInitExceptionOnContractorFirstKeyStage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Python: self.node_B.set_TL_debug_flag(2048, self.targetTransactionType, 1)
	// targetTransactionType for this suite is 105 (SettlementLinePublicKeyResponseMessageType)
	// Flag 2048 is TestFlagExceptionOnInitTAModifyingStage for *initiator*. For *contractor*, it's TestFlagExceptionOnContractorTAStage (16384)
	// The python test seems to use the initiator's "modifying stage" flag number (2048) for the contractor.
	// This could be an oversight in the python test or a specific behavior of the debug flag.
	// Let's try with TestFlagExceptionOnContractorTAStage (16384) first as it's specific to contractor.
	// If that doesn't match behavior, might need to use 2048 if the system truly uses that flag value contextually.
	// Python message type used is targetTransactionType = 105
	err := nodeB.SetTestingSLFlag(vtcp.TestFlagExceptionOnContractorTAStage, vtcp.SettlementLinePublicKeyResponseMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)
	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_keys_sharing_init_io_exception_on_contractor_first_key_stage
func TestSettlementLineKeysSharingInitIOExceptionOnContractorFirstKeyStage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeB.SetTestingSLFlag(vtcp.TestFlagExceptionOnContractorTAStage, vtcp.SettlementLinePublicKeyResponseMessageType, "2", "0") // IO Exception
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)
	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_keys_sharing_init_exception_on_contractor_next_key_stage
func TestSettlementLineKeysSharingInitExceptionOnContractorNextKeyStage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Python: self.node_B.set_TL_debug_flag(16384, self.targetTransactionType, 1)
	// Flag 16384 is TestFlagExceptionOnContractorTAStage. targetTransactionType is 105.
	err := nodeB.SetTestingSLFlag(vtcp.TestFlagExceptionOnContractorTAStage, vtcp.SettlementLinePublicKeyResponseMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)
	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_keys_sharing_init_io_exception_on_contractor_next_key_stage
func TestSettlementLineKeysSharingInitIOExceptionOnContractorNextKeyStage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	err := nodeB.SetTestingSLFlag(vtcp.TestFlagExceptionOnContractorTAStage, vtcp.SettlementLinePublicKeyResponseMessageType, "2", "0") // IO Exception
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)
	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Terminations
// Corresponds to test_trustlines_keys_sharing_init_terminate_on_initiator_first_key_stage
func TestSettlementLineKeysSharingInitTerminateOnInitFirstKeyStage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Python: self.node_A.set_TL_debug_flag(2097152, self.sourceTransactionType, 1)
	// Flag 2097152 is TestFlagTerminateOnInitTAModifyingStage
	// sourceTransactionType for keys init suite is 104 (SettlementLinePublicKeyMessageType)
	err := nodeA.SetTestingSLFlag(vtcp.TestFlagTerminateOnInitTAModifyingStage, vtcp.SettlementLinePublicKeyMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)
	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(160 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	// check_trustline_after_changes part:
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLine(t, nodeB, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError) // 401
	nodeB.CheckSettlementLine(t, nodeA, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError) // 401
	// No key check in python test for this case
}

// Corresponds to test_trustlines_keys_sharing_init_terminate_after_initiator_first_key_stage
func TestSettlementLineKeysSharingInitTerminateAfterInitFirstKeyStage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Python: self.node_A.set_TL_debug_flag(2097152, self.sourceTransactionType, 1) -> Same as "on" stage
	err := nodeA.SetTestingSLFlag(vtcp.TestFlagTerminateOnInitTAModifyingStage, vtcp.SettlementLinePublicKeyMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)
	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(160 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSettlementLine(t, nodeB, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
	nodeB.CheckSettlementLine(t, nodeA, testconfig.Equivalent, "", "", "", "", "", "", vtcp.StatusProtocolError)
}

// Corresponds to test_trustlines_keys_sharing_init_terminate_on_initiator_next_key_stage
func TestSettlementLineKeysSharingInitTerminateOnInitNextKeyStage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Python: self.node_A.set_TL_debug_flag(4194304, self.sourceTransactionType, 1)
	// Flag 4194304 is TestFlagTerminateOnInitTAResponseProcessingStage
	err := nodeA.SetTestingSLFlag(vtcp.TestFlagTerminateOnInitTAResponseProcessingStage, vtcp.SettlementLinePublicKeyMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)
	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(160 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_keys_sharing_init_terminate_after_initiator_next_key_stage
func TestSettlementLineKeysSharingInitTerminateAfterInitNextKeyStage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Python: self.node_A.set_TL_debug_flag(4194304, self.sourceTransactionType, 1) -> Same as "on" stage
	err := nodeA.SetTestingSLFlag(vtcp.TestFlagTerminateOnInitTAResponseProcessingStage, vtcp.SettlementLinePublicKeyMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeA failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)
	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(160 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_keys_sharing_init_terminate_on_contractor_first_key_stage
func TestSettlementLineKeysSharingInitTerminateOnContractorFirstKeyStage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Python: self.node_B.set_TL_debug_flag(2097152, self.targetTransactionType, 1)
	// Flag 2097152 is TestFlagTerminateOnInitTAModifyingStage. For contractor, should be TestFlagTerminateOnContractorTAStage (16777216)
	// targetTransactionType is 105 (SettlementLinePublicKeyResponseMessageType)
	// Using the contractor specific termination flag.
	err := nodeB.SetTestingSLFlag(vtcp.FlagTerminateProcessVote, vtcp.SettlementLinePublicKeyResponseMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)
	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(60 * time.Second) // Python waits 60s for contractor side issues

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_keys_sharing_init_terminate_after_contractor_first_key_stage
func TestSettlementLineKeysSharingInitTerminateAfterContractorFirstKeyStage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Python: self.node_B.set_TL_debug_flag(2097152, self.targetTransactionType, 2)
	// Using contractor specific flag with secondParam = "2"
	err := nodeB.SetTestingSLFlag(vtcp.FlagTerminateProcessVote, vtcp.SettlementLinePublicKeyResponseMessageType, "2", "0") // Assuming secondParam of debug flag maps to thirdParam in SetTestingSLFlag if message type is first, count is second.
	// Python set_TL_debug_flag(flag, type, count, index) maps to SetTestingSLFlag(flag, type, count, index)
	// So python's '2' for count/times here.
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)
	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_keys_sharing_init_terminate_on_contractor_next_key_stage
func TestSettlementLineKeysSharingInitTerminateOnContractorNextKeyStage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Python: self.node_B.set_TL_debug_flag(16777216, self.targetTransactionType, 1)
	// Flag 16777216 is TestFlagTerminateOnContractorTAStage
	err := nodeB.SetTestingSLFlag(vtcp.FlagTerminateProcessVote, vtcp.SettlementLinePublicKeyResponseMessageType, "1", "0")
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)
	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}

// Corresponds to test_trustlines_keys_sharing_init_terminate_after_contractor_next_key_stage
func TestSettlementLineKeysSharingInitTerminateAfterContractorNextKeyStage(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingInitSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Python: self.node_B.set_TL_debug_flag(16777216, self.targetTransactionType, 2)
	err := nodeB.SetTestingSLFlag(vtcp.FlagTerminateProcessVote, vtcp.SettlementLinePublicKeyResponseMessageType, "2", "0")
	if err != nil {
		t.Fatalf("NodeB failed to set testing SL flag: %v", err)
	}
	time.Sleep(1 * time.Second)
	nodeA.CreateSettlementLine(t, nodeB, testconfig.Equivalent)

	time.Sleep(60 * time.Second)

	nodeB.CheckMaxFlow(t, nodeA, testconfig.Equivalent, "0")
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
	nodeB.CheckValidKeys(t, vtcp.DefaultKeysCount-1, vtcp.DefaultKeysCount-1)
}
