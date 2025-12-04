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
	// directPaymentSeneNodesNextNodeIndex is used to assign unique IP addresses to nodes across different test functions in this file.
	directExchangePaymentSeneNodesNextNodeIndex = 1
)

func getNextIPForDirectPaymentSevenNodesTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentSevenNodes, directExchangePaymentSeneNodesNextNodeIndex)
	directExchangePaymentSeneNodesNextNodeIndex++
	return ip
}

func setupNodesForDirectPaymentSevenNodesTest(t *testing.T) []*vtcp.Node {
	startIndex := 1
	node1 := vtcp.NewNode(t, getNextIPForDirectPaymentSevenNodesTest(), fmt.Sprintf("node%d", startIndex))
	node2 := vtcp.NewNode(t, getNextIPForDirectPaymentSevenNodesTest(), fmt.Sprintf("node%d", startIndex+1))
	node3 := vtcp.NewNode(t, getNextIPForDirectPaymentSevenNodesTest(), fmt.Sprintf("node%d", startIndex+2))
	node4 := vtcp.NewNode(t, getNextIPForDirectPaymentSevenNodesTest(), fmt.Sprintf("node%d", startIndex+3))
	node5 := vtcp.NewNode(t, getNextIPForDirectPaymentSevenNodesTest(), fmt.Sprintf("node%d", startIndex+4))
	node6 := vtcp.NewNode(t, getNextIPForDirectPaymentSevenNodesTest(), fmt.Sprintf("node%d", startIndex+5))
	node7 := vtcp.NewNode(t, getNextIPForDirectPaymentSevenNodesTest(), fmt.Sprintf("node%d", startIndex+6))

	nodes := []*vtcp.Node{node1, node2, node3, node4, node5, node6, node7}
	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes, false)

	node2.CreateChannelAndSettlementLineAndCheck(t, node1, testconfig.Equivalent, "3000")
	node3.CreateChannelAndSettlementLineAndCheck(t, node2, testconfig.Equivalent, "2500")
	node4.CreateChannelAndSettlementLineAndCheck(t, node3, testconfig.Equivalent, "2000")
	node5.CreateChannelAndSettlementLineAndCheck(t, node4, testconfig.Equivalent, "5000")
	node6.CreateChannelAndSettlementLineAndCheck(t, node5, testconfig.Equivalent, "1000")
	node7.CreateChannelAndSettlementLineAndCheck(t, node6, testconfig.Equivalent, "1500")

	return nodes
}

func Test1DirectExchange7NormalAmount(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test2DirectExchange7NodesAmountTooBig(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1500", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4aLostAskNeighborToReserveAmountMsgFromCoordinatorToFirstIntermediateNodeExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_request_to_intermediate_node_on_reservation (Corresponds to flag 1 in two_nodes tests)
	if err := nodes[0].SetTestingFlag(vtcp.FlagForbidSendRequestToIntermediateReservation, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node1: %v", err)
	}
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4bLostAskNeighborToApproveFurtherNodeReservationMsgFromCoordinatorToFirstIntermediateNodeExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_message_to_coordinator_on_reservation (New flag)
	if err := nodes[0].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node1: %v", err)
	}
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4cLostAskRemoteNodeToApproveReservationMsgFromCoordinatorToLastIntermediateNodeExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_message_to_coordinator_on_reservation (New flag)
	if err := nodes[0].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, nodes[2].GetIPAddressForRequests(), ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node1: %v", err)
	}
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4dLostProcessNeighborAmountReservationResponseMsgFromFirstIntermediateNodeToCoordinatorExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_response_to_intermediate_node_on_reservation (Corresponds to flag 2)
	if err := nodes[1].SetTestingFlag(vtcp.FlagForbidSendResponseToIntemediateOnReservation, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node2: %v", err)
	}
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4eLostMsgFromFirstIntermediateNodeToNextIntermediateNodeExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_request_to_intermediate_node_on_reservation (Corresponds to flag 1)
	if err := nodes[1].SetTestingFlag(vtcp.FlagForbidSendRequestToIntermediateReservation, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node2: %v", err)
	}
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4fLostMsgFromNextIntermediateNodeToPreviousExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_response_to_intermediate_node_on_reservation (Corresponds to flag 2)
	if err := nodes[2].SetTestingFlag(vtcp.FlagForbidSendResponseToIntemediateOnReservation, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node3: %v", err)
	}
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4gLostMsgFromLastIntermediateNodeReceiverExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_request_to_intermediate_node_on_reservation (Corresponds to flag 1)
	if err := nodes[5].SetTestingFlag(vtcp.FlagForbidSendRequestToIntermediateReservation, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node6: %v", err)
	}
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4hLostMsgReceiverToPreviousExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_response_to_intermediate_node_on_reservation (Corresponds to flag 2)
	if err := nodes[6].SetTestingFlag(vtcp.FlagForbidSendResponseToIntemediateOnReservation, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node7: %v", err)
	}
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4jLostProcessNeighborFurtherReservationResponseMsgFromFirstIntermediateNodeToCoordinatorExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_message_to_coordinator_on_reservation (New flag)
	if err := nodes[0].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, "", ""); err != nil { // Assuming node1 is the coordinator contextually
		t.Fatalf("SetTestingFlag failed for node1: %v", err)
	}
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4kLostProcessRemoteNodeResponseMsgFromLastIntermediateNodeToCoordinatorExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_message_to_coordinator_on_reservation (New flag)
	if err := nodes[5].SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node6: %v", err)
	}
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test5LostMessageWithPathFinalConfigurationExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_message_with_final_path_configuration (New flag)
	if err := nodes[0].SetTestingFlag(vtcp.FlagForbidSendMessageFinalPathConfig, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node1: %v", err)
	}
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK) // Default status in Pyt	hon
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test6aLostMessageWithFinalConfigurationToIntermediateNodeExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_message_on_final_amount_clarification (New flag)
	if err := nodes[0].SetTestingFlag(vtcp.FlagForbidSendMessageFinalAmountClarification, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node1: %v", err)
	}
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusProtocolError)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test6bLostMessageWithFinalConfigurationToCoordinatorExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_message_on_final_amount_clarification (New flag)
	if err := nodes[3].SetTestingFlag(vtcp.FlagForbidSendMessageFinalAmountClarification, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node4: %v", err)
	}
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusProtocolError)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test7aLostMsgWithPublicKeysToFirstIntermediateNodeExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	err := nodes[0].SetTestingFlag(vtcp.FlagForbidSendMessageVoteStage, "", "")
	if err != nil {
		t.Fatalf("SetTestingFlag failed: %v", err)
	}

	uuid, err := nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusProtocolError)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	nodes[0].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[1].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[2].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[3].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[4].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[5].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[6].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)

	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, true, 0)
	nodes[2].CheckSerializedTransaction(t, true, 0)
	nodes[3].CheckSerializedTransaction(t, true, 0)
	nodes[4].CheckSerializedTransaction(t, true, 0)
	nodes[5].CheckSerializedTransaction(t, true, 0)
	nodes[6].CheckSerializedTransaction(t, true, 0)

	time.Sleep(time.Duration(vtcp.NodePaymentRecoveryAttempts-1) * vtcp.NodePaymentRecoveryTimePeriodSec * time.Second)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0) // Check sync immediately after sleep

	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, true, 0)
	nodes[2].CheckSerializedTransaction(t, true, 0)
	nodes[3].CheckSerializedTransaction(t, true, 0)
	nodes[4].CheckSerializedTransaction(t, true, 0)
	nodes[5].CheckSerializedTransaction(t, true, 0)
	nodes[6].CheckSerializedTransaction(t, false, 0)

	sleepDuration := (vtcp.ObservingCntBlocksForClaiming * vtcp.ObservingCntSecondsPerBlock) - ((vtcp.NodePaymentRecoveryAttempts - 1) * vtcp.NodePaymentRecoveryTimePeriodSec)
	if sleepDuration > 0 {
		time.Sleep(time.Duration(sleepDuration) * time.Second)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[6].CheckPaymentTransaction(t, "", 0, 0, 0, 0)

	nodes[0].CheckExchangeMaxFlow(t, nodes[6], testconfig.Equivalent, []string{testconfig.Equivalent}, "1000")
}

func Test7bLostMsgWithSignatureToCoordinatorExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	if err := nodes[1].SetTestingFlag(vtcp.FlagForbidSendMessageVoteConsistency, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node2: %v", err)
	}
	if err := nodes[5].SetTestingFlag(vtcp.FlagForbidSendMessageVoteConsistency, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node6: %v", err)
	}

	uuid, err := nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusProtocolError)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	nodes[0].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[1].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[2].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[3].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[4].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[5].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[6].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)

	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, true, 0)
	nodes[2].CheckSerializedTransaction(t, true, 0)
	nodes[3].CheckSerializedTransaction(t, true, 0)
	nodes[4].CheckSerializedTransaction(t, true, 0)
	nodes[5].CheckSerializedTransaction(t, true, 0)
	nodes[6].CheckSerializedTransaction(t, true, 0)

	time.Sleep(time.Duration(vtcp.NodePaymentRecoveryAttempts-1) * vtcp.NodePaymentRecoveryTimePeriodSec * time.Second)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, true, 0)
	nodes[2].CheckSerializedTransaction(t, true, 0)
	nodes[3].CheckSerializedTransaction(t, true, 0)
	nodes[4].CheckSerializedTransaction(t, true, 0)
	nodes[5].CheckSerializedTransaction(t, true, 0)
	nodes[6].CheckSerializedTransaction(t, true, 0)

	sleepDuration := (vtcp.ObservingCntBlocksForClaiming * vtcp.ObservingCntSecondsPerBlock) - ((vtcp.NodePaymentRecoveryAttempts - 1) * vtcp.NodePaymentRecoveryTimePeriodSec)
	if sleepDuration > 0 {
		time.Sleep(time.Duration(sleepDuration) * time.Second)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, false, 0)
	nodes[2].CheckSerializedTransaction(t, false, 0)
	nodes[3].CheckSerializedTransaction(t, false, 0)
	nodes[4].CheckSerializedTransaction(t, false, 0)
	nodes[5].CheckSerializedTransaction(t, false, 0)
	nodes[6].CheckSerializedTransaction(t, false, 0)

	nodes[0].CheckExchangeMaxFlow(t, nodes[6], testconfig.Equivalent, []string{testconfig.Equivalent}, "1000")
}

func Test7cLostMsgWithPublicKeyHashFromIntermediateNodeToParticipantsExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_message_on_vote_stage (Corresponds to flag 3)
	if err := nodes[1].SetTestingFlag(vtcp.FlagForbidSendMessageVoteStage, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node2: %v", err)
	}
	if err := nodes[4].SetTestingFlag(vtcp.FlagForbidSendMessageVoteStage, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node5: %v", err)
	}
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusProtocolError)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3) // Default 3s sleep from Python original
}

func Test7dLostMsgWithSignatureFromCoordinatorToAllIntermediateNodesExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_message_on_vote_consistency (Corresponds to flag 4)
	if err := nodes[0].SetTestingFlag(vtcp.FlagForbidSendMessageVoteConsistency, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node1: %v", err)
	}

	uuid, err := nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	time.Sleep(vtcp.WaitingParticipantsVotesSec * time.Second)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	nodes[0].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[1].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[2].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[3].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[4].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[5].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[6].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)

	nodes[0].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 0, 1)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, false, 0)
	nodes[2].CheckSerializedTransaction(t, false, 0)
	nodes[3].CheckSerializedTransaction(t, false, 0)
	nodes[4].CheckSerializedTransaction(t, false, 0)
	nodes[5].CheckSerializedTransaction(t, false, 0)
	nodes[6].CheckSerializedTransaction(t, false, 0)
	// Python: # self.node_1.check_max_flow({self.node_7.address: 0}) -> Not checking max flow here
}

func Test7eLostMsgWithSignatureFromCoordinatorToAllIntermediateNodesAlsoOnRecoveryExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// flag_forbid_send_message_on_recovery_stage = 103 (New)
	if err := nodes[0].SetTestingFlag(vtcp.FlagForbidSendMessageVoteConsistency, "", ""); err != nil { // Applied vote consistency flag
		t.Fatalf("SetTestingFlag failed for node1: %v", err)
	}

	uuid, err := nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	time.Sleep(vtcp.WaitingParticipantsVotesSec * time.Second)

	nodes[0].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[1].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[2].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[3].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[4].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[5].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[6].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)

	nodes[0].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 0, 1)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, true, 0)
	nodes[2].CheckSerializedTransaction(t, true, 0)
	nodes[3].CheckSerializedTransaction(t, true, 0)
	nodes[4].CheckSerializedTransaction(t, true, 0)
	nodes[5].CheckSerializedTransaction(t, true, 0)
	nodes[6].CheckSerializedTransaction(t, true, 0)

	// Python: time.sleep((self.recovering_attempts - 1) * (self.recovering_time_period_sec + self.waiting_for_message))
	// self.waiting_for_message is likely 1 second.
	time.Sleep(time.Duration(vtcp.NodePaymentRecoveryAttempts-1) * (vtcp.NodePaymentRecoveryTimePeriodSec + 1) * time.Second)

	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 0)

	// Python: time.sleep(self.observing_claim_period * 2)
	// self.observing_claim_period might be (ObservingCntBlocksForClaiming * ObservingCntSecondsPerBlock)
	observingClaimPeriod := time.Duration(vtcp.ObservingCntBlocksForClaiming*vtcp.ObservingCntSecondsPerBlock) * time.Second
	time.Sleep(observingClaimPeriod * 2)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	// TODO: Verify observing_ParticipantsVotesPresent_response mapping
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 7, 1, 1) // Using Claimed as placeholder
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 7, 1, 1) // Using Claimed as placeholder
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 7, 1, 1) // Using Claimed as placeholder
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 7, 1, 1) // Using Claimed as placeholder
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 7, 1, 1) // Using Claimed as placeholder
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 7, 1, 0) // Using Claimed as placeholder

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, false, 0)
	nodes[2].CheckSerializedTransaction(t, false, 0)
	nodes[3].CheckSerializedTransaction(t, false, 0)
	nodes[4].CheckSerializedTransaction(t, false, 0)
	nodes[5].CheckSerializedTransaction(t, false, 0)
	nodes[6].CheckSerializedTransaction(t, false, 0)

	time.Sleep(observingClaimPeriod)
	// TODO: Verify observing_ParticipantsVotesPresent_response mapping
	nodes[0].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 7, 0, 1) // Using Claimed as placeholder
	// Python: # self.node_1.check_max_flow({self.node_7.address: 0})
}

func Test8aCrashCoordinatorAfterSendingMessageOnVotingExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_throw_exception_on_vote (Corresponds to flag 6)
	if err := nodes[0].SetTestingFlag(vtcp.FlagThrowExceptionVote, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node1: %v", err)
	}

	uuid, err := nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusProtocolError)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	nodes[0].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[1].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[2].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[3].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[4].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[5].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[6].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)

	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, true, 0)
	nodes[2].CheckSerializedTransaction(t, true, 0)
	nodes[3].CheckSerializedTransaction(t, true, 0)
	nodes[4].CheckSerializedTransaction(t, true, 0)
	nodes[5].CheckSerializedTransaction(t, true, 0)
	nodes[6].CheckSerializedTransaction(t, true, 0)

	time.Sleep(time.Duration(vtcp.NodePaymentRecoveryAttempts-1) * vtcp.NodePaymentRecoveryTimePeriodSec * time.Second)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, true, 0)
	nodes[2].CheckSerializedTransaction(t, true, 0)
	nodes[3].CheckSerializedTransaction(t, true, 0)
	nodes[4].CheckSerializedTransaction(t, true, 0)
	nodes[5].CheckSerializedTransaction(t, true, 0)
	nodes[6].CheckSerializedTransaction(t, true, 0)

	sleepDuration := (vtcp.ObservingCntBlocksForClaiming * vtcp.ObservingCntSecondsPerBlock) - ((vtcp.NodePaymentRecoveryAttempts - 1) * vtcp.NodePaymentRecoveryTimePeriodSec)
	if sleepDuration > 0 {
		time.Sleep(time.Duration(sleepDuration) * time.Second)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, false, 0)
	nodes[2].CheckSerializedTransaction(t, false, 0)
	nodes[3].CheckSerializedTransaction(t, false, 0)
	nodes[4].CheckSerializedTransaction(t, false, 0)
	nodes[5].CheckSerializedTransaction(t, false, 0)
	nodes[6].CheckSerializedTransaction(t, false, 0)

	nodes[0].CheckExchangeMaxFlow(t, nodes[6], testconfig.Equivalent, []string{testconfig.Equivalent}, "1000")
}

func Test8bCrashCoordinatorAfterReceivingMessageWithSignatureExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_throw_exception_on_vote_consistency (Corresponds to flag 7)
	if err := nodes[0].SetTestingFlag(vtcp.FlagThrowExceptionVoteConsistency, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node1: %v", err)
	}

	uuid, err := nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusProtocolError)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	nodes[0].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[1].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[2].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[3].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[4].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[5].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[6].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)

	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, true, 0)
	nodes[2].CheckSerializedTransaction(t, true, 0)
	nodes[3].CheckSerializedTransaction(t, true, 0)
	nodes[4].CheckSerializedTransaction(t, true, 0)
	nodes[5].CheckSerializedTransaction(t, true, 0)
	nodes[6].CheckSerializedTransaction(t, true, 0)

	time.Sleep(time.Duration(vtcp.NodePaymentRecoveryAttempts-1) * vtcp.NodePaymentRecoveryTimePeriodSec * time.Second)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, true, 0)
	nodes[2].CheckSerializedTransaction(t, true, 0)
	nodes[3].CheckSerializedTransaction(t, true, 0)
	nodes[4].CheckSerializedTransaction(t, true, 0)
	nodes[5].CheckSerializedTransaction(t, true, 0)
	nodes[6].CheckSerializedTransaction(t, true, 0)

	sleepDuration := (vtcp.ObservingCntBlocksForClaiming * vtcp.ObservingCntSecondsPerBlock) - ((vtcp.NodePaymentRecoveryAttempts - 1) * vtcp.NodePaymentRecoveryTimePeriodSec)
	if sleepDuration > 0 {
		time.Sleep(time.Duration(sleepDuration) * time.Second)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, false, 0)
	nodes[2].CheckSerializedTransaction(t, false, 0)
	nodes[3].CheckSerializedTransaction(t, false, 0)
	nodes[4].CheckSerializedTransaction(t, false, 0)
	nodes[5].CheckSerializedTransaction(t, false, 0)
	nodes[6].CheckSerializedTransaction(t, false, 0)

	nodes[0].CheckExchangeMaxFlow(t, nodes[6], testconfig.Equivalent, []string{testconfig.Equivalent}, "1000")
}

func Test9aCrashIntermediateNodeRunPreviousNeighborRequestProcessingStageExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_throw_exception_on_previous_neighbor_request (Corresponds to flag 5)
	if err := nodes[2].SetTestingFlag(vtcp.FlagThrowExceptionPreviousNeighborRequest, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node3: %v", err)
	}
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}
func Test9bCrashIntermediateNodeRunCoordinatorRequestProcessingStageExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_throw_exception_on_coordinator_request_processing (New flag)
	if err := nodes[2].SetTestingFlag(vtcp.FlagThrowExceptionCoordinatorRequest, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node3: %v", err)
	}
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test9cCrashIntermediateNodeRunNextNeighborResponseProcessingStageExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_throw_exception_on_next_neighbor_response_processing (New flag)
	if err := nodes[2].SetTestingFlag(vtcp.FlagThrowExceptionNextNeighborResponse, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node3: %v", err)
	}
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test9dCrashIntermediateNodeAfterSignBeforeSendResponseExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_throw_exception_on_vote (Corresponds to flag 6)
	if err := nodes[1].SetTestingFlag(vtcp.FlagThrowExceptionVote, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node2: %v", err)
	}
	if err := nodes[2].SetTestingFlag(vtcp.FlagThrowExceptionVote, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node3: %v", err)
	}
	if err := nodes[4].SetTestingFlag(vtcp.FlagThrowExceptionVote, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node5: %v", err)
	}
	uuid, err := nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusProtocolError)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	nodes[0].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[1].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[2].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[3].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[4].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[5].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[6].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)

	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, true, 0)
	nodes[2].CheckSerializedTransaction(t, true, 0)
	nodes[3].CheckSerializedTransaction(t, true, 0)
	nodes[4].CheckSerializedTransaction(t, true, 0)
	nodes[5].CheckSerializedTransaction(t, true, 0)
	nodes[6].CheckSerializedTransaction(t, true, 0)

	time.Sleep(time.Duration(vtcp.NodePaymentRecoveryAttempts-1) * vtcp.NodePaymentRecoveryTimePeriodSec * time.Second)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, true, 0)
	nodes[2].CheckSerializedTransaction(t, true, 0)
	nodes[3].CheckSerializedTransaction(t, true, 0)
	nodes[4].CheckSerializedTransaction(t, true, 0)
	nodes[5].CheckSerializedTransaction(t, true, 0)
	nodes[6].CheckSerializedTransaction(t, true, 0)

	sleepDuration := (vtcp.ObservingCntBlocksForClaiming * vtcp.ObservingCntSecondsPerBlock) - ((vtcp.NodePaymentRecoveryAttempts - 1) * vtcp.NodePaymentRecoveryTimePeriodSec)
	if sleepDuration > 0 {
		time.Sleep(time.Duration(sleepDuration) * time.Second)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, false, 0)
	nodes[2].CheckSerializedTransaction(t, false, 0)
	nodes[3].CheckSerializedTransaction(t, false, 0)
	nodes[4].CheckSerializedTransaction(t, false, 0)
	nodes[5].CheckSerializedTransaction(t, false, 0)
	nodes[6].CheckSerializedTransaction(t, false, 0)

	nodes[0].CheckExchangeMaxFlow(t, nodes[6], testconfig.Equivalent, []string{testconfig.Equivalent}, "1000")
}
func Test9eStopProcessIntermediateNodeAfterVotesReceivingBeforeCommittingExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_terminate_process_on_vote_consistency (Corresponds to flag 11)
	if err := nodes[1].SetTestingFlag(vtcp.FlagTerminateProcessVoteConsistency, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node2: %v", err)
	}
	if err := nodes[2].SetTestingFlag(vtcp.FlagTerminateProcessVoteConsistency, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node3: %v", err)
	}
	if err := nodes[4].SetTestingFlag(vtcp.FlagTerminateProcessVoteConsistency, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node5: %v", err)
	}
	uuid, err := nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	nodes[0].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[1].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[2].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[3].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[4].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[5].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[6].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)

	nodes[0].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 0, 1)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, false, 0)
	nodes[2].CheckSerializedTransaction(t, false, 0)
	nodes[3].CheckSerializedTransaction(t, false, 0)
	nodes[4].CheckSerializedTransaction(t, false, 0)
	nodes[5].CheckSerializedTransaction(t, false, 0)
	nodes[6].CheckSerializedTransaction(t, false, 0)
	// Python: # self.node_1.check_max_flow({self.node_7.address: 0})
}

func Test10aStopProcessCoordinatorAfterSendingMessageOnVotingExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_terminate_process_on_vote (Corresponds to flag 10)
	if err := nodes[0].SetTestingFlag(vtcp.FlagTerminateProcessVote, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node1: %v", err)
	}
	// Python status_code=None implies OK or that the call might not return/complete normally due to process termination.
	// Assuming OK for transaction creation initiation. The actual outcome is process termination.
	uuid, err := nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	nodes[0].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[1].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[2].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[3].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[4].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[5].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[6].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)

	// Python: self.node_1.check_payment_transaction(outgoing_receipts_cnt=1)
	// Other nodes default to 0 counts.
	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 1)
	nodes[1].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[2].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[3].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[4].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[5].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[6].CheckPaymentTransaction(t, "", 0, 0, 0, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, false, 0)
	nodes[2].CheckSerializedTransaction(t, false, 0)
	nodes[3].CheckSerializedTransaction(t, false, 0)
	nodes[4].CheckSerializedTransaction(t, false, 0)
	nodes[5].CheckSerializedTransaction(t, false, 0)
	nodes[6].CheckSerializedTransaction(t, false, 0)

	nodes[0].CheckExchangeMaxFlow(t, nodes[6], testconfig.Equivalent, []string{testconfig.Equivalent}, "0")
}

func Test10bStopProcessCoordinatorAfterReceivingMessageWithSignaturesExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_terminate_process_on_vote_consistency (Corresponds to flag 11)
	if err := nodes[0].SetTestingFlag(vtcp.FlagTerminateProcessVoteConsistency, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node1: %v", err)
	}
	uuid, err := nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK) // Python status_code=None
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	nodes[0].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[1].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[2].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[3].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[4].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[5].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[6].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)

	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 1) // Python: outgoing_receipts_cnt=1
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, true, 0)
	nodes[2].CheckSerializedTransaction(t, true, 0)
	nodes[3].CheckSerializedTransaction(t, true, 0)
	nodes[4].CheckSerializedTransaction(t, true, 0)
	nodes[5].CheckSerializedTransaction(t, true, 0)
	nodes[6].CheckSerializedTransaction(t, true, 0)

	time.Sleep(time.Duration(vtcp.NodePaymentRecoveryAttempts-1) * vtcp.NodePaymentRecoveryTimePeriodSec * time.Second)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 1) // Python: outgoing_receipts_cnt=1
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, true, 0)
	nodes[2].CheckSerializedTransaction(t, true, 0)
	nodes[3].CheckSerializedTransaction(t, true, 0)
	nodes[4].CheckSerializedTransaction(t, true, 0)
	nodes[5].CheckSerializedTransaction(t, true, 0)
	nodes[6].CheckSerializedTransaction(t, true, 0)

	sleepDuration := (vtcp.ObservingCntBlocksForClaiming * vtcp.ObservingCntSecondsPerBlock) - ((vtcp.NodePaymentRecoveryAttempts - 1) * vtcp.NodePaymentRecoveryTimePeriodSec)
	if sleepDuration > 0 {
		time.Sleep(time.Duration(sleepDuration) * time.Second)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 1) // Python: outgoing_receipts_cnt=1
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, false, 0)
	nodes[2].CheckSerializedTransaction(t, false, 0)
	nodes[3].CheckSerializedTransaction(t, false, 0)
	nodes[4].CheckSerializedTransaction(t, false, 0)
	nodes[5].CheckSerializedTransaction(t, false, 0)
	nodes[6].CheckSerializedTransaction(t, false, 0)

	nodes[0].CheckExchangeMaxFlow(t, nodes[6], testconfig.Equivalent, []string{testconfig.Equivalent}, "1000")
}

func Test11aStopProcessIntermediateNodeRunPreviousNeighborRequestProcessingStageExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_terminate_process_on_previous_neighbor_request (Corresponds to flag 9)
	if err := nodes[2].SetTestingFlag(vtcp.FlagTerminateProcessPreviousNeighborRequest, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node3: %v", err)
	}
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test11bStopProcessIntermediateNodeRunCoordinatorRequestProcessingStageExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_terminate_process_on_coordinator_request_processing (New flag)
	// TODO: Verify flag value for flag_terminate_process_on_coordinator_request_processing
	if err := nodes[2].SetTestingFlag(vtcp.FlagTerminateProcessCoordinatorRequest, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node3: %v", err)
	}
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test11cStopProcessIntermediateNodeRunNextNeighborResponseProcessingStageExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_terminate_process_on_next_neighbor_response_processing (New flag)
	// TODO: Verify flag value for flag_terminate_process_on_next_neighbor_response_processing
	if err := nodes[2].SetTestingFlag(vtcp.FlagTerminateProcessNextNeighborResponse, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node3: %v", err)
	}
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test11dStopProcessIntermediateNodeAfterSignBeforeSendingExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_terminate_process_on_vote (Corresponds to flag 10)
	if err := nodes[1].SetTestingFlag(vtcp.FlagTerminateProcessVote, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node2: %v", err)
	}
	if err := nodes[2].SetTestingFlag(vtcp.FlagTerminateProcessVote, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node3: %v", err)
	}
	if err := nodes[4].SetTestingFlag(vtcp.FlagTerminateProcessVote, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node5: %v", err)
	}
	uuid, err := nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusProtocolError)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	nodes[0].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[1].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[2].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[3].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[4].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[5].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[6].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)

	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, true, 0)
	nodes[2].CheckSerializedTransaction(t, true, 0)
	nodes[3].CheckSerializedTransaction(t, true, 0)
	nodes[4].CheckSerializedTransaction(t, true, 0)
	nodes[5].CheckSerializedTransaction(t, true, 0)
	nodes[6].CheckSerializedTransaction(t, true, 0)

	time.Sleep(time.Duration(vtcp.NodePaymentRecoveryAttempts-1) * vtcp.NodePaymentRecoveryTimePeriodSec * time.Second)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, true, 0)
	nodes[2].CheckSerializedTransaction(t, true, 0)
	nodes[3].CheckSerializedTransaction(t, true, 0)
	nodes[4].CheckSerializedTransaction(t, true, 0)
	nodes[5].CheckSerializedTransaction(t, true, 0)
	nodes[6].CheckSerializedTransaction(t, true, 0)

	sleepDuration := (vtcp.ObservingCntBlocksForClaiming * vtcp.ObservingCntSecondsPerBlock) - ((vtcp.NodePaymentRecoveryAttempts - 1) * vtcp.NodePaymentRecoveryTimePeriodSec)
	if sleepDuration > 0 {
		time.Sleep(time.Duration(sleepDuration) * time.Second)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, false, 0)
	nodes[2].CheckSerializedTransaction(t, false, 0)
	nodes[3].CheckSerializedTransaction(t, false, 0)
	nodes[4].CheckSerializedTransaction(t, false, 0)
	nodes[5].CheckSerializedTransaction(t, false, 0)
	nodes[6].CheckSerializedTransaction(t, false, 0)

	nodes[0].CheckExchangeMaxFlow(t, nodes[6], testconfig.Equivalent, []string{testconfig.Equivalent}, "1000")
}

func Test11eStopProcessIntermediateNodeAfterVotesReceivingBeforeCommittingExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_terminate_process_on_vote_consistency (Corresponds to flag 11)
	if err := nodes[1].SetTestingFlag(vtcp.FlagTerminateProcessVoteConsistency, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node2: %v", err)
	}
	if err := nodes[2].SetTestingFlag(vtcp.FlagTerminateProcessVoteConsistency, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node3: %v", err)
	}
	if err := nodes[4].SetTestingFlag(vtcp.FlagTerminateProcessVoteConsistency, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node5: %v", err)
	}
	uuid, err := nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	nodes[0].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[1].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[2].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[3].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[4].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[5].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[6].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)

	nodes[0].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 0, 1)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, false, 0)
	nodes[2].CheckSerializedTransaction(t, false, 0)
	nodes[3].CheckSerializedTransaction(t, false, 0)
	nodes[4].CheckSerializedTransaction(t, false, 0)
	nodes[5].CheckSerializedTransaction(t, false, 0)
	nodes[6].CheckSerializedTransaction(t, false, 0)
	// Python: # self.node_1.check_max_flow({self.node_7.address: 0})
}
