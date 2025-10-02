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
	// directPaymentTwoNodesNextNodeIndex is used to assign unique IP addresses to nodes across different test functions in this file.
	directPaymentTwoNodesNextNodeIndex = 1
)

func Test1DirectPaymentNormalAmount(t *testing.T) {
	nodeA := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeA")
	directPaymentTwoNodesNextNodeIndex++
	nodeB := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeB")
	directPaymentTwoNodesNextNodeIndex++
	nodeC := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeC")
	directPaymentTwoNodesNextNodeIndex++
	nodeD := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeD")
	directPaymentTwoNodesNextNodeIndex++

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB, nodeC, nodeD}, false)

	nodeB.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlowBatch(t,
		[]vtcp.MaxFlowBatchCheck{
			{Node: nodeB, ExpectedMaxFlow: "1000"}},
		testconfig.Equivalent)

	nodeC.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlowBatch(t,
		[]vtcp.MaxFlowBatchCheck{
			{Node: nodeB, ExpectedMaxFlow: "1000"},
			{Node: nodeC, ExpectedMaxFlow: "1000"}},
		testconfig.Equivalent)

	nodeB.CreateChannelAndSettlementLineAndCheck(t, nodeD, testconfig.Equivalent, "500")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "600", vtcp.StatusOK)
	nodeA.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 2, 0, 1)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 2, 1, 0)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)

	time.Sleep(3 * time.Second)

	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "400")

	nodeA.CheckMaxFlowBatch(t,
		[]vtcp.MaxFlowBatchCheck{
			{Node: nodeB, ExpectedMaxFlow: "400"},
			{Node: nodeC, ExpectedMaxFlow: "1000"}},
		testconfig.Equivalent)

	nodeB.CreateTransactionCheckStatus(t, nodeA, testconfig.Equivalent, "600", vtcp.StatusOK)

	nodeA.CheckMaxFlowBatch(t,
		[]vtcp.MaxFlowBatchCheck{
			{Node: nodeB, ExpectedMaxFlow: "1000"},
			{Node: nodeC, ExpectedMaxFlow: "1000"}},
		testconfig.Equivalent)
}

// Test2DirectPaymentOvertrustAmount is an analogue of test_2_direct_payment_overtrust_amount
func Test2DirectPaymentOvertrustAmount(t *testing.T) {
	nodeA := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeA")
	directPaymentTwoNodesNextNodeIndex++
	nodeB := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeB")
	directPaymentTwoNodesNextNodeIndex++
	nodeC := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeC")
	directPaymentTwoNodesNextNodeIndex++
	nodeD := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeD")
	directPaymentTwoNodesNextNodeIndex++

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB, nodeC, nodeD}, false)

	// Common setup based on Python's prepare_topology
	nodeB.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000")

	// Test-specific logic for Test2DirectPaymentOvertrustAmount
	nodeB.SetSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000")
	nodeC.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlowBatch(t,
		[]vtcp.MaxFlowBatchCheck{
			{Node: nodeB, ExpectedMaxFlow: "1000"},
			{Node: nodeC, ExpectedMaxFlow: "1000"}},
		testconfig.Equivalent)
	nodeC.CreateChannelAndSettlementLineAndCheck(t, nodeD, testconfig.Equivalent, "500")
	nodeD.CheckMaxFlow(t, nodeC, testconfig.Equivalent, "500")

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "2500", vtcp.StatusInsufficientFunds)
	nodeA.CheckMaxFlowBatch(t,
		[]vtcp.MaxFlowBatchCheck{
			{Node: nodeB, ExpectedMaxFlow: "1000"},
			{Node: nodeC, ExpectedMaxFlow: "1000"}},
		testconfig.Equivalent)

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "2000", vtcp.StatusInsufficientFunds)
	nodeA.CheckMaxFlowBatch(t,
		[]vtcp.MaxFlowBatchCheck{
			{Node: nodeB, ExpectedMaxFlow: "1000"},
			{Node: nodeC, ExpectedMaxFlow: "1000"}},
		testconfig.Equivalent)

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1200", vtcp.StatusInsufficientFunds)
	nodeA.CheckMaxFlowBatch(t,
		[]vtcp.MaxFlowBatchCheck{
			{Node: nodeB, ExpectedMaxFlow: "1000"},
			{Node: nodeC, ExpectedMaxFlow: "1000"}},
		testconfig.Equivalent)
}

// Test3PaymentWithoutPath is an analogue of test_3_payment_without_path
func Test3aPaymentWithoutPath(t *testing.T) {
	nodeA := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeA")
	directPaymentTwoNodesNextNodeIndex++
	nodeB := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeB")
	directPaymentTwoNodesNextNodeIndex++
	nodeD := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeD")
	directPaymentTwoNodesNextNodeIndex++

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB, nodeD}, false)

	// Common setup based on Python's prepare_topology
	nodeB.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000")

	// Test-specific logic
	nodeA.CreateTransactionCheckStatus(t, nodeD, testconfig.Equivalent, "200", vtcp.StatusNoPaymentRoutes)
}

// Test3DirectPaymentWithTrustlineWithoutFlow is an analogue of test_3_direct_payment_with_trustline_without_flow
func Test3bDirectPaymentWithTrustlineWithoutFlow(t *testing.T) {
	nodeA := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeA")
	directPaymentTwoNodesNextNodeIndex++
	nodeB := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeB")
	directPaymentTwoNodesNextNodeIndex++
	nodeD := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeD")
	directPaymentTwoNodesNextNodeIndex++

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB, nodeD}, false)

	// Common setup based on Python's prepare_topology
	nodeB.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000")

	// Test-specific logic
	nodeA.CreateChannelAndSettlementLineAndCheck(t, nodeD, testconfig.Equivalent, "500")
	nodeA.CreateTransactionCheckStatus(t, nodeD, testconfig.Equivalent, "200", vtcp.StatusNoPaymentRoutes)
}

// Test4aLostMessageOnReservationStageToReceiver is an analogue of test_4a_lost_message_on_reservation_stage_to_receiver
func Test4aLostMessageOnReservationStageToReceiver(t *testing.T) {
	nodeA := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeA")
	directPaymentTwoNodesNextNodeIndex++
	nodeB := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeB")
	directPaymentTwoNodesNextNodeIndex++

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB}, false)

	// Common setup based on Python's prepare_topology
	nodeB.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000")

	// Test-specific logic
	// var flagForbidSendInitMessage uint32 = 1 // Placeholder value from self.flag_forbid_send_init_message
	err = nodeA.SetTestingFlag(vtcp.FlagForbidSendInitMessage, "", "")
	if err != nil {
		t.Fatalf("SetTestingFlag failed: %v", err)
	}
	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	// TODO : check synchronized settlement line
}

// Test4bLostMessageOnReservationStageToCoordinator is an analogue of test_4b_lost_message_on_reservation_stage_to_coordinator
func Test4bLostMessageOnReservationStageToCoordinator(t *testing.T) {
	nodeA := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeA")
	directPaymentTwoNodesNextNodeIndex++
	nodeB := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeB")
	directPaymentTwoNodesNextNodeIndex++

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB}, false)

	// Common setup based on Python's prepare_topology
	nodeB.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000")

	// Test-specific logic
	// var flagForbidSendResponseToIntermediateNodeOnReservation uint32 = 2 // Placeholder value
	err = nodeB.SetTestingFlag(vtcp.FlagForbidSendResponseToIntemediateOnReservation, "", "")
	if err != nil {
		t.Fatalf("SetTestingFlag failed: %v", err)
	}
	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	// TODO : check synchronized settlement line
}

// Test5aLostMessageWithPublicKeysToReceiver is an analogue of test_5a_lost_message_with_public_keys_to_receiver
func Test5aLostMessageWithPublicKeysToReceiver(t *testing.T) {
	nodeA := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeA")
	directPaymentTwoNodesNextNodeIndex++
	nodeB := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeB")
	directPaymentTwoNodesNextNodeIndex++

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB}, false)

	// Common setup
	nodeB.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000")

	err = nodeA.SetTestingFlag(vtcp.FlagForbidSendMessageVoteStage, "", "")
	if err != nil {
		t.Fatalf("SetTestingFlag failed: %v", err)
	}

	// it works because coordinator resend votes message
	_, err = nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed unexpectedly: %v", err)
	}
	time.Sleep(vtcp.WaitingParticipantsVotesSec * time.Second)

	nodeA.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 2, 0, 1)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 2, 1, 0)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}

// Test5bLostMessageWithSignatureToCoordinator is an analogue of test_5b_lost_message_with_signature_to_coordinator
func Test5bLostMessageWithSignatureToCoordinator(t *testing.T) {
	nodeA := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeA")
	directPaymentTwoNodesNextNodeIndex++
	nodeB := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeB")
	directPaymentTwoNodesNextNodeIndex++

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB}, false)

	// Common setup based on Python's prepare_topology
	nodeB.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000")

	// Test-specific logic
	// var flagForbidSendMessageOnVoteConsistency uint32 = 4 // Placeholder value
	err = nodeB.SetTestingFlag(vtcp.FlagForbidSendMessageVoteConsistency, "", "")
	if err != nil {
		t.Fatalf("SetTestingFlag failed: %v", err)
	}

	uuid, err := nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusProtocolError)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed unexpectedly: %v", err)
	}
	// Python original has extensive checks and sleeps here.
	time.Sleep(vtcp.WaitingParticipantsVotesSec * time.Second)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeA.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodeB.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodeA.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 0)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, true, 0)
	time.Sleep(vtcp.NodePaymentRecoveryTimePeriodSec * (vtcp.NodePaymentRecoveryAttempts - 1) * time.Second)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeA.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 0)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, true, 0)
	time.Sleep((vtcp.ObservingCntBlocksForClaiming*vtcp.ObservingCntSecondsPerBlock -
		vtcp.NodePaymentRecoveryTimePeriodSec*(vtcp.NodePaymentRecoveryAttempts-1)) * time.Second)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeA.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}

// Test5cLostMessageWithSignatureToReceiver is an analogue of test_5c_lost_message_with_signature_to_receiver
func Test5cLostMessageWithSignatureToReceiver(t *testing.T) {
	nodeA := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeA")
	directPaymentTwoNodesNextNodeIndex++
	nodeB := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeB")
	directPaymentTwoNodesNextNodeIndex++

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB}, false)

	// Common setup based on Python's prepare_topology
	nodeB.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000")

	// Test-specific logic
	// var flagForbidSendMessageOnVoteConsistency uint32 = 4 // Placeholder value
	err = nodeA.SetTestingFlag(vtcp.FlagForbidSendMessageVoteConsistency, "", "")
	if err != nil {
		t.Fatalf("SetTestingFlag failed: %v", err)
	}

	uuid, err := nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed unexpectedly: %v", err)
	}
	// Python original has extensive checks and sleeps here.
	time.Sleep(vtcp.WaitingParticipantsVotesSec * time.Second)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeA.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodeB.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodeA.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 0)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, true, 0)
}

// Test6aReceiverCrashReservationStage is an analogue of test_6a_receiver_crash_reservation_stage
func Test6aReceiverCrashReservationStage(t *testing.T) {
	nodeA := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeA")
	directPaymentTwoNodesNextNodeIndex++
	nodeB := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeB")
	directPaymentTwoNodesNextNodeIndex++

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB}, false)

	// Common setup based on Python's prepare_topology
	nodeB.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000")

	// Test-specific logic
	// var flagThrowExceptionOnPreviousNeighborRequest uint32 = 5 // Placeholder value
	err = nodeB.SetTestingFlag(vtcp.FlagThrowExceptionPreviousNeighborRequest, "", "")
	if err != nil {
		t.Fatalf("SetTestingFlag failed: %v", err)
	}
	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	time.Sleep(10 * time.Second)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
}

// Test6bReceiverCrashAfterSignBeforeSending is an analogue of test_6b_receiver_crash_after_sign_before_sending
func Test6bReceiverCrashAfterSignBeforeSending(t *testing.T) {
	nodeA := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeA")
	directPaymentTwoNodesNextNodeIndex++
	nodeB := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeB")
	directPaymentTwoNodesNextNodeIndex++

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB}, false)

	// Common setup based on Python's prepare_topology
	nodeB.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000")

	// Test-specific logic
	// var flagThrowExceptionOnVote uint32 = 6 // Placeholder value
	err = nodeB.SetTestingFlag(vtcp.FlagThrowExceptionVote, "", "")
	if err != nil {
		t.Fatalf("SetTestingFlag failed: %v", err)
	}

	uuid, err := nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusProtocolError)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed unexpectedly: %v", err)
	}

	time.Sleep(vtcp.WaitingParticipantsVotesSec * time.Second)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeA.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodeB.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodeA.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 0)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, true, 0)
	time.Sleep(vtcp.NodePaymentRecoveryTimePeriodSec * (vtcp.NodePaymentRecoveryAttempts - 1) * time.Second)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeA.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 0)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, true, 0)
	time.Sleep((vtcp.ObservingCntBlocksForClaiming*vtcp.ObservingCntSecondsPerBlock -
		vtcp.NodePaymentRecoveryTimePeriodSec*(vtcp.NodePaymentRecoveryAttempts-1)) * time.Second)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeA.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}

// Test6cReceiverCrashAfterVotesReceivingBeforeCommitting is an analogue of test_6c_receiver_crash_after_votes_receiving_before_committing
func Test6cReceiverCrashAfterVotesReceivingBeforeCommitting(t *testing.T) {
	nodeA := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeA")
	directPaymentTwoNodesNextNodeIndex++
	nodeB := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeB")
	directPaymentTwoNodesNextNodeIndex++

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB}, false)

	// Common setup based on Python's prepare_topology
	nodeB.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000")

	// Test-specific logic
	// var flagThrowExceptionOnVoteConsistency uint32 = 7 // Placeholder value
	err = nodeB.SetTestingFlag(vtcp.FlagThrowExceptionVoteConsistency, "", "")
	if err != nil {
		t.Fatalf("SetTestingFlag failed: %v", err)
	}
	uuid, err := nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK) // Python test expects normal creation
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed unexpectedly: %v", err)
	}

	time.Sleep(vtcp.WaitingParticipantsVotesSec * time.Second)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeA.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodeB.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodeA.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 2, 0, 1)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 2, 1, 0)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}

// Test7aCoordinatorCrashReservationStage is an analogue of test_7a_coordinator_crash_reservation_stage
func Test7aCoordinatorCrashReservationStage(t *testing.T) {
	nodeA := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeA")
	directPaymentTwoNodesNextNodeIndex++
	nodeB := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeB")
	directPaymentTwoNodesNextNodeIndex++

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB}, false)

	// Common setup based on Python's prepare_topology
	nodeB.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000")

	// Test-specific logic
	// var flagThrowExceptionOnPreviousNeighborRequest uint32 = 5 // Placeholder value
	err = nodeA.SetTestingFlag(vtcp.FlagThrowExceptionPreviousNeighborRequest, "", "")
	if err != nil {
		t.Fatalf("SetTestingFlag failed: %v", err)
	}
	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
}

// Test7bCoordinatorCrashAfterSendingMessageOnVoting is an analogue of test_7b_coordinator_crash_after_sending_message_on_voting
func Test7bCoordinatorCrashAfterSendingMessageOnVoting(t *testing.T) {
	nodeA := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeA")
	directPaymentTwoNodesNextNodeIndex++
	nodeB := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeB")
	directPaymentTwoNodesNextNodeIndex++

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB}, false)

	// Common setup based on Python's prepare_topology
	nodeB.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000")

	// Test-specific logic
	// var flagThrowExceptionOnVote uint32 = 6 // Placeholder value
	err = nodeA.SetTestingFlag(vtcp.FlagThrowExceptionVote, "", "")
	if err != nil {
		t.Fatalf("SetTestingFlag failed: %v", err)
	}

	uuid, err := nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed unexpectedly: %v", err)
	}

	time.Sleep(vtcp.WaitingParticipantsVotesSec * time.Second)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeA.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodeB.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodeA.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 0)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, true, 0)
	time.Sleep(vtcp.NodePaymentRecoveryTimePeriodSec * (vtcp.NodePaymentRecoveryAttempts - 1) * time.Second)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeA.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 0)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, true, 0)
	time.Sleep((vtcp.ObservingCntBlocksForClaiming*vtcp.ObservingCntSecondsPerBlock -
		vtcp.NodePaymentRecoveryTimePeriodSec*(vtcp.NodePaymentRecoveryAttempts-1)) * time.Second)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeA.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}

// Test7cCoordinatorCrashAfterReceivingMessageWithSignature is an analogue of test_7c_coordinator_crash_after_receiving_message_with_signature
func Test7cCoordinatorCrashAfterReceivingMessageWithSignature(t *testing.T) {
	nodeA := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeA")
	directPaymentTwoNodesNextNodeIndex++
	nodeB := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeB")
	directPaymentTwoNodesNextNodeIndex++

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB}, false)

	// Common setup based on Python's prepare_topology
	nodeB.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000")

	// Test-specific logic
	// var flagThrowExceptionOnVoteConsistency uint32 = 7 // Placeholder value
	err = nodeA.SetTestingFlag(vtcp.FlagThrowExceptionVoteConsistency, "", "")
	if err != nil {
		t.Fatalf("SetTestingFlag failed: %v", err)
	}

	uuid, err := nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed unexpectedly: %v", err)
	}

	time.Sleep(vtcp.WaitingParticipantsVotesSec * time.Second)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeA.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodeB.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodeA.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 0)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, true, 0)
	time.Sleep(vtcp.NodePaymentRecoveryTimePeriodSec * (vtcp.NodePaymentRecoveryAttempts - 1) * time.Second)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeA.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 0)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, true, 0)
	time.Sleep((vtcp.ObservingCntBlocksForClaiming*vtcp.ObservingCntSecondsPerBlock -
		vtcp.NodePaymentRecoveryTimePeriodSec*(vtcp.NodePaymentRecoveryAttempts-1)) * time.Second)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeA.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}

// Test7dCoordinatorCrashAfterApprovingBeforeSendingMessageWithSignature is an analogue of test_7d_coordinator_crash_after_approving_before_sending_message_with_signature
func Test7dCoordinatorCrashAfterApprovingBeforeSendingMessageWithSignature(t *testing.T) {
	nodeA := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeA")
	directPaymentTwoNodesNextNodeIndex++
	nodeB := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeB")
	directPaymentTwoNodesNextNodeIndex++

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB}, false)

	// Common setup based on Python's prepare_topology
	nodeB.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000")

	// Test-specific logic
	// var flagThrowExceptionCoordinatorAfterApprove uint32 = 8 // Placeholder for flag_throw_exception_on_coordinator_after_approve_before_send_message
	err = nodeA.SetTestingFlag(vtcp.FlagThrowExceptionCoordinatorAfterApprove, "", "")
	if err != nil {
		t.Fatalf("SetTestingFlag failed: %v", err)
	}
	uuid, err := nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK) // Python status_code=None
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed unexpectedly: %v", err)
	}

	time.Sleep(vtcp.WaitingParticipantsVotesSec * time.Second)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeA.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodeB.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodeA.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 2, 0, 1)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 2, 1, 0)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}

// Test8aReceiverProcessCrashReservationStage is an analogue of test_8a_receiver_process_crash_reservation_stage
func Test8aReceiverProcessCrashReservationStage(t *testing.T) {
	nodeA := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeA")
	directPaymentTwoNodesNextNodeIndex++
	nodeB := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeB")
	directPaymentTwoNodesNextNodeIndex++

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB}, false)

	// Common setup based on Python's prepare_topology
	nodeB.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000")

	// Test-specific logic
	// var flagTerminateProcessOnPreviousNeighborRequest uint32 = 9 // Placeholder value
	err = nodeB.SetTestingFlag(vtcp.FlagTerminateProcessPreviousNeighborRequest, "", "")
	if err != nil {
		t.Fatalf("SetTestingFlag failed: %v", err)
	}
	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
}

// Test8bReceiverProcessCrashAfterSignBeforeSending is an analogue of test_8b_receiver_process_crash_after_sign_before_sending
func Test8bReceiverProcessCrashAfterSignBeforeSending(t *testing.T) {
	nodeA := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeA")
	directPaymentTwoNodesNextNodeIndex++
	nodeB := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeB")
	directPaymentTwoNodesNextNodeIndex++

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB}, false)

	// Common setup based on Python's prepare_topology
	nodeB.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000")

	// Test-specific logic
	// var flagTerminateProcessOnVote uint32 = 10 // Placeholder value
	err = nodeB.SetTestingFlag(vtcp.FlagTerminateProcessVote, "", "")
	if err != nil {
		t.Fatalf("SetTestingFlag failed: %v", err)
	}
	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusProtocolError)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
}

// Test8cReceiverProcessCrashAfterVotesReceivingBeforeCommitting is an analogue of test_8c_receiver_process_crash_after_votes_receiving_before_committing
func Test8cReceiverProcessCrashAfterVotesReceivingBeforeCommitting(t *testing.T) {
	nodeA := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeA")
	directPaymentTwoNodesNextNodeIndex++
	nodeB := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeB")
	directPaymentTwoNodesNextNodeIndex++

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB}, false)

	// Common setup based on Python's prepare_topology
	nodeB.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000")

	// Test-specific logic
	// var flagTerminateProcessOnVoteConsistency uint32 = 11 // Placeholder value
	err = nodeB.SetTestingFlag(vtcp.FlagTerminateProcessVoteConsistency, "", "")
	if err != nil {
		t.Fatalf("SetTestingFlag failed: %v", err)
	}

	uuid, err := nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK) // Python test expects normal creation
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed unexpectedly: %v", err)
	}

	time.Sleep(vtcp.WaitingParticipantsVotesSec * time.Second)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeA.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodeB.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodeA.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 0)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, true, 0)
	time.Sleep(vtcp.NodePaymentRecoveryTimePeriodSec * (vtcp.NodePaymentRecoveryAttempts - 1) * time.Second)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeA.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 0)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, true, 0)
	time.Sleep((vtcp.ObservingCntBlocksForClaiming*vtcp.ObservingCntSecondsPerBlock -
		vtcp.NodePaymentRecoveryTimePeriodSec*(vtcp.NodePaymentRecoveryAttempts-1)) * time.Second)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeA.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}

// Test9aCoordinatorProcessCrashReservationStage is an analogue of test_9a_coordinator_process_crash_reservation_stage
func Test9aCoordinatorProcessCrashReservationStage(t *testing.T) {
	nodeA := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeA")
	directPaymentTwoNodesNextNodeIndex++
	nodeB := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeB")
	directPaymentTwoNodesNextNodeIndex++

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB}, false)

	// Common setup based on Python's prepare_topology
	nodeB.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000")

	// Test-specific logic
	// var flagTerminateProcessOnPreviousNeighborRequest uint32 = 9 // Placeholder value
	err = nodeA.SetTestingFlag(vtcp.FlagTerminateProcessPreviousNeighborRequest, "", "")
	if err != nil {
		t.Fatalf("SetTestingFlag failed: %v", err)
	}
	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
}

// Test9bCoordinatorProcessCrashAfterSendingMessageOnVoting is an analogue of test_9b_coordinator_process_crash_after_sending_message_on_voting
func Test9bCoordinatorProcessCrashAfterSendingMessageOnVoting(t *testing.T) {
	nodeA := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeA")
	directPaymentTwoNodesNextNodeIndex++
	nodeB := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeB")
	directPaymentTwoNodesNextNodeIndex++

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB}, false)

	// Common setup based on Python's prepare_topology
	nodeB.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000")

	// Test-specific logic
	// var flagTerminateProcessOnVote uint32 = 10 // Placeholder value
	err = nodeA.SetTestingFlag(vtcp.FlagTerminateProcessVote, "", "")
	if err != nil {
		t.Fatalf("SetTestingFlag failed: %v", err)
	}

	uuid, err := nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed unexpectedly: %v", err)
	}

	time.Sleep(vtcp.WaitingParticipantsVotesSec * time.Second)
	nodeA.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodeB.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodeA.CheckPaymentTransaction(t, "", 0, 0, 0, 1)
	nodeB.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}

// Test9cCoordinatorProcessCrashAfterReceivingMessageWithSignature is an analogue of test_9c_coordinator_process_crash_after_receiving_message_with_signature
func Test9cCoordinatorProcessCrashAfterReceivingMessageWithSignature(t *testing.T) {
	nodeA := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeA")
	directPaymentTwoNodesNextNodeIndex++
	nodeB := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeB")
	directPaymentTwoNodesNextNodeIndex++

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB}, false)

	// Common setup based on Python's prepare_topology
	nodeB.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000")

	// Test-specific logic
	// var flagTerminateProcessOnVoteConsistency uint32 = 11 // Placeholder value
	err = nodeA.SetTestingFlag(vtcp.FlagTerminateProcessVoteConsistency, "", "")
	if err != nil {
		t.Fatalf("SetTestingFlag failed: %v", err)
	}

	uuid, err := nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed unexpectedly: %v", err)
	}

	time.Sleep(vtcp.WaitingParticipantsVotesSec * time.Second)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeA.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodeB.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodeA.CheckPaymentTransaction(t, "", 0, 0, 0, 1)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 0)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, true, 0)
	time.Sleep(vtcp.NodePaymentRecoveryTimePeriodSec * (vtcp.NodePaymentRecoveryAttempts - 1) * time.Second)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeA.CheckPaymentTransaction(t, "", 0, 0, 0, 1)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 0)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, true, 0)
	time.Sleep((vtcp.ObservingCntBlocksForClaiming*vtcp.ObservingCntSecondsPerBlock -
		vtcp.NodePaymentRecoveryTimePeriodSec*(vtcp.NodePaymentRecoveryAttempts-1)) * time.Second)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeA.CheckPaymentTransaction(t, "", 0, 0, 0, 1)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}

// Test9dCoordinatorProcessCrashAfterApprovingBeforeSendingMessageWithSignature is an analogue of test_9d_coordinator_process_crash_after_approving_before_sending_message_with_signature
func Test9dCoordinatorProcessCrashAfterApprovingBeforeSendingMessageWithSignature(t *testing.T) {
	nodeA := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeA")
	directPaymentTwoNodesNextNodeIndex++
	nodeB := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentTwoNodes, directPaymentTwoNodesNextNodeIndex), "nodeB")
	directPaymentTwoNodesNextNodeIndex++

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB}, false)

	// Common setup based on Python's prepare_topology
	nodeB.CreateChannelAndSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")
	nodeA.CheckMaxFlow(t, nodeB, testconfig.Equivalent, "1000")

	// Test-specific logic
	// var flagTerminateProcessCoordinatorAfterApprove uint32 = 12 // Placeholder value
	err = nodeA.SetTestingFlag(vtcp.FlagTerminateProcessCoordinatorAfterApprove, "", "")
	if err != nil {
		t.Fatalf("SetTestingFlag failed: %v", err)
	}

	uuid, err := nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "1000", vtcp.StatusOK) // Python status_code=None
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed unexpectedly: %v", err)
	}

	time.Sleep(vtcp.WaitingParticipantsVotesSec * time.Second)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeA.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodeB.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodeA.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 2, 0, 1)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 2, 1, 0)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
}
