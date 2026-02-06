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
	nodes[0].SetTestingFlag(t, vtcp.FlagForbidSendRequestToIntermediateReservation, "", "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4bLostAskNeighborToApproveFurtherNodeReservationMsgFromCoordinatorToFirstIntermediateNodeExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_message_to_coordinator_on_reservation (New flag)
	nodes[0].SetTestingFlag(t, vtcp.FlagForbidSendMessageToCoordinatorReservation, "", "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4cLostAskRemoteNodeToApproveReservationMsgFromCoordinatorToLastIntermediateNodeExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_message_to_coordinator_on_reservation (New flag)
	nodes[0].SetTestingFlag(t, vtcp.FlagForbidSendMessageToCoordinatorReservation, nodes[2].GetIPAddressForRequests(), "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4dLostProcessNeighborAmountReservationResponseMsgFromFirstIntermediateNodeToCoordinatorExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_response_to_intermediate_node_on_reservation (Corresponds to flag 2)
	nodes[1].SetTestingFlag(t, vtcp.FlagForbidSendResponseToIntemediateOnReservation, "", "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4eLostMsgFromFirstIntermediateNodeToNextIntermediateNodeExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_request_to_intermediate_node_on_reservation (Corresponds to flag 1)
	nodes[1].SetTestingFlag(t, vtcp.FlagForbidSendRequestToIntermediateReservation, "", "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4fLostMsgFromNextIntermediateNodeToPreviousExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_response_to_intermediate_node_on_reservation (Corresponds to flag 2)
	nodes[2].SetTestingFlag(t, vtcp.FlagForbidSendResponseToIntemediateOnReservation, "", "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4gLostMsgFromLastIntermediateNodeReceiverExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_request_to_intermediate_node_on_reservation (Corresponds to flag 1)
	nodes[5].SetTestingFlag(t, vtcp.FlagForbidSendRequestToIntermediateReservation, "", "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4hLostMsgReceiverToPreviousExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_response_to_intermediate_node_on_reservation (Corresponds to flag 2)
	nodes[6].SetTestingFlag(t, vtcp.FlagForbidSendResponseToIntemediateOnReservation, "", "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4jLostProcessNeighborFurtherReservationResponseMsgFromFirstIntermediateNodeToCoordinatorExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_message_to_coordinator_on_reservation (New flag)
	nodes[0].SetTestingFlag(t, vtcp.FlagForbidSendMessageToCoordinatorReservation, "", "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4kLostProcessRemoteNodeResponseMsgFromLastIntermediateNodeToCoordinatorExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_message_to_coordinator_on_reservation (New flag)
	nodes[5].SetTestingFlag(t, vtcp.FlagForbidSendMessageToCoordinatorReservation, "", "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test5LostMessageWithPathFinalConfigurationExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	nodes[0].SetTestingFlag(t, vtcp.FlagForbidSendMessageFinalPathConfig, "", "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK) // Default status in Pyt	hon
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test6aLostMessageWithFinalConfigurationToIntermediateNodeExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	nodes[0].SetTestingFlag(t, vtcp.FlagForbidSendMessageFinalAmountClarification, "", "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusProtocolError)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test6bLostMessageWithFinalConfigurationToCoordinatorExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	nodes[3].SetTestingFlag(t, vtcp.FlagForbidSendMessageFinalAmountClarification, "", "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusProtocolError)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test7aLostMsgWithPublicKeysToFirstIntermediateNodeExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	nodes[0].SetTestingFlag(t, vtcp.FlagForbidSendMessageVoteStage, "", "")

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

	nodes[1].SetTestingFlag(t, vtcp.FlagForbidSendMessageVoteConsistency, "", "")
	nodes[5].SetTestingFlag(t, vtcp.FlagForbidSendMessageVoteConsistency, "", "")

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

	nodes[1].SetTestingFlag(t, vtcp.FlagForbidSendMessageVoteStage, "", "")
	nodes[4].SetTestingFlag(t, vtcp.FlagForbidSendMessageVoteStage, "", "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusProtocolError)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3) // Default 3s sleep from Python original
}

func Test7dLostMsgWithSignatureFromCoordinatorToAllIntermediateNodesExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	nodes[0].SetTestingFlag(t, vtcp.FlagForbidSendMessageVoteConsistency, "", "")

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

	nodes[0].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 0, 1)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, false, 0)
	nodes[2].CheckSerializedTransaction(t, false, 0)
	nodes[3].CheckSerializedTransaction(t, false, 0)
	nodes[4].CheckSerializedTransaction(t, false, 0)
	nodes[5].CheckSerializedTransaction(t, false, 0)
	nodes[6].CheckSerializedTransaction(t, false, 0)

	// TODO: uncomment this when topology cashing is fixed
	// nodes[0].CheckExchangeMaxFlow(t, nodes[6], testconfig.Equivalent, []string{testconfig.Equivalent}, "0")
}

func Test7eLostMsgWithSignatureFromCoordinatorToAllIntermediateNodesAlsoOnRecoveryExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	nodes[0].SetTestingFlag(t, vtcp.FlagForbidSendMessageVoteConsistency, "", "")
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

func Test7fLostMsgWithSignatureFromCoordinatorToAllIntermediateNodesIncludinfRecoveryStageExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_message_on_vote_consistency (Corresponds to flag)
	nodes[0].SetTestingFlag(t, vtcp.FlagForbidSendMessageVoteConsistency, "", "")
	nodes[1].SetTestingFlag(t, vtcp.FlagForbidSendMessageRecoveryStage, "", "")
	nodes[2].SetTestingFlag(t, vtcp.FlagForbidSendMessageRecoveryStage, "", "")
	nodes[3].SetTestingFlag(t, vtcp.FlagForbidSendMessageRecoveryStage, "", "")
	nodes[4].SetTestingFlag(t, vtcp.FlagForbidSendMessageRecoveryStage, "", "")
	nodes[5].SetTestingFlag(t, vtcp.FlagForbidSendMessageRecoveryStage, "", "")
	nodes[6].SetTestingFlag(t, vtcp.FlagForbidSendMessageRecoveryStage, "", "")

	uuid, err := nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	// Waiting for 15 minutes to simulate the time period of the recovery stage with observing
	time.Sleep(15 * time.Minute)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	nodes[0].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[1].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[2].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[3].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[4].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[5].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[6].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)

	nodes[0].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 0, 1)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, false, 0)
	nodes[2].CheckSerializedTransaction(t, false, 0)
	nodes[3].CheckSerializedTransaction(t, false, 0)
	nodes[4].CheckSerializedTransaction(t, false, 0)
	nodes[5].CheckSerializedTransaction(t, false, 0)
	nodes[6].CheckSerializedTransaction(t, false, 0)

	// TODO: uncomment this when topology cashing is fixed
	// nodes[0].CheckExchangeMaxFlow(t, nodes[6], testconfig.Equivalent, []string{testconfig.Equivalent}, "0")
}

func Test8aCrashCoordinatorAfterSendingMessageOnVotingExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	nodes[0].SetTestingFlag(t, vtcp.FlagThrowExceptionVote, "", "")
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

	nodes[0].SetTestingFlag(t, vtcp.FlagThrowExceptionVoteConsistency, "", "")
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

	nodes[2].SetTestingFlag(t, vtcp.FlagThrowExceptionPreviousNeighborRequest, "", "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}
func Test9bCrashIntermediateNodeRunCoordinatorRequestProcessingStageExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	nodes[2].SetTestingFlag(t, vtcp.FlagThrowExceptionCoordinatorRequest, "", "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test9cCrashIntermediateNodeRunNextNeighborResponseProcessingStageExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	nodes[2].SetTestingFlag(t, vtcp.FlagThrowExceptionNextNeighborResponse, "", "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test9dCrashIntermediateNodeAfterSignBeforeSendResponseExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	nodes[1].SetTestingFlag(t, vtcp.FlagThrowExceptionVote, "", "")
	nodes[2].SetTestingFlag(t, vtcp.FlagThrowExceptionVote, "", "")
	nodes[4].SetTestingFlag(t, vtcp.FlagThrowExceptionVote, "", "")
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
func Test9eCrashProcessIntermediateNodeAfterVotesReceivingBeforeCommittingExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	nodes[1].SetTestingFlag(t, vtcp.FlagThrowExceptionVoteConsistency, "", "")
	nodes[2].SetTestingFlag(t, vtcp.FlagThrowExceptionVoteConsistency, "", "")
	nodes[4].SetTestingFlag(t, vtcp.FlagThrowExceptionVoteConsistency, "", "")
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

func Test9fCrashProcessIntermediateNodeBeforeSubmittingClaime(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	nodes[0].SetTestingFlag(t, vtcp.FlagForbidSendMessageVoteConsistency, "", "")
	nodes[1].SetTestingFlag(t, vtcp.FlagForbidSendMessageRecoveryStage+vtcp.FlagThrowExceptionOnObservingSubmitClaimStage, "", "")
	uuid, err := nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}

	// Waiting for 18 minutes to simulate the time period of the recovery stage with observing
	time.Sleep(18 * time.Minute)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	nodes[0].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[1].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[2].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[3].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[4].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[5].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[6].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)

	nodes[0].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 0, 1)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 0)

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

	nodes[0].SetTestingFlag(t, vtcp.FlagTerminateProcessVote, "", "")
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

	nodes[0].SetTestingFlag(t, vtcp.FlagTerminateProcessVoteConsistency, "", "")
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

func Test10fTerminateProcessIntermediateNodeBeforeSubmittingClaime(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	nodes[0].SetTestingFlag(t, vtcp.FlagForbidSendMessageVoteConsistency, "", "")
	nodes[1].SetTestingFlag(t, vtcp.FlagForbidSendMessageRecoveryStage+vtcp.FlagTerminateProcessOnObservingSubmitClaimStage, "", "")
	uuid, err := nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}

	// Waiting for 15 minutes to simulate the time period of the recovery stage with observing
	time.Sleep(20 * time.Minute)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	nodes[0].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[1].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[2].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[3].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[4].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[5].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[6].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)

	nodes[0].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 0, 1)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, false, 0)
	nodes[2].CheckSerializedTransaction(t, false, 0)
	nodes[3].CheckSerializedTransaction(t, false, 0)
	nodes[4].CheckSerializedTransaction(t, false, 0)
	nodes[5].CheckSerializedTransaction(t, false, 0)
	nodes[6].CheckSerializedTransaction(t, false, 0)
	// Python: # self.node_1.check_max_flow({self.node_7.address: 0})
}

func Test11aStopProcessIntermediateNodeRunPreviousNeighborRequestProcessingStageExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	nodes[2].SetTestingFlag(t, vtcp.FlagTerminateProcessPreviousNeighborRequest, "", "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test11bStopProcessIntermediateNodeRunCoordinatorRequestProcessingStageExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// TODO: Verify flag value for flag_terminate_process_on_coordinator_request_processing
	nodes[2].SetTestingFlag(t, vtcp.FlagTerminateProcessCoordinatorRequest, "", "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test11cStopProcessIntermediateNodeRunNextNeighborResponseProcessingStageExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// TODO: Verify flag value for flag_terminate_process_on_next_neighbor_response_processing
	nodes[2].SetTestingFlag(t, vtcp.FlagTerminateProcessNextNeighborResponse, "", "")
	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test11dStopProcessIntermediateNodeAfterSignBeforeSendingExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_terminate_process_on_vote (Corresponds to flag 10)
	nodes[1].SetTestingFlag(t, vtcp.FlagTerminateProcessVote, "", "")
	nodes[2].SetTestingFlag(t, vtcp.FlagTerminateProcessVote, "", "")
	nodes[4].SetTestingFlag(t, vtcp.FlagTerminateProcessVote, "", "")
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
	nodes[1].SetTestingFlag(t, vtcp.FlagTerminateProcessVoteConsistency, "", "")
	nodes[2].SetTestingFlag(t, vtcp.FlagTerminateProcessVoteConsistency, "", "")
	nodes[4].SetTestingFlag(t, vtcp.FlagTerminateProcessVoteConsistency, "", "")
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

func Test12aAuditDuringRecoveryStageExchange(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	// self.flag_forbid_send_message_on_vote_consistency (Corresponds to flag)
	nodes[0].SetTestingFlag(t, vtcp.FlagForbidSendMessageVoteConsistency, "", "")
	nodes[1].SetTestingFlag(t, vtcp.FlagForbidSendMessageRecoveryStage, "", "")

	uuid, err := nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "1000", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	// Waiting for 15 minutes to simulate the time period of the recovery stage with observing
	time.Sleep(5 * time.Minute)

	nodes[0].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[1].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[2].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[3].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[4].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[5].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[6].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)

	nodes[0].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 0, 1)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, true, 0)
	nodes[2].CheckSerializedTransaction(t, false, 0)
	nodes[3].CheckSerializedTransaction(t, false, 0)
	nodes[4].CheckSerializedTransaction(t, false, 0)
	nodes[5].CheckSerializedTransaction(t, false, 0)
	nodes[6].CheckSerializedTransaction(t, false, 0)

	nodes[1].SetSettlementLine(t, nodes[0], testconfig.Equivalent, "4000", vtcp.StatusOK)

	time.Sleep(15 * time.Second)

	nodes[2].SetSettlementLine(t, nodes[1], testconfig.Equivalent, "3000", vtcp.StatusOK)

	time.Sleep(15 * time.Minute)

	nodes[0].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 0, 1)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 1)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 1, 7, 1, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, false, 0)
	nodes[2].CheckSerializedTransaction(t, false, 0)
	nodes[3].CheckSerializedTransaction(t, false, 0)
	nodes[4].CheckSerializedTransaction(t, false, 0)
	nodes[5].CheckSerializedTransaction(t, false, 0)
	nodes[6].CheckSerializedTransaction(t, false, 0)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)
}

func Test12bAuditDuringRecoveryStageExchangeWithOnePaymentBefore(t *testing.T) {
	nodes := setupNodesForDirectPaymentSevenNodesTest(t)

	uuid, err := nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "100", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}

	// self.flag_forbid_send_message_on_vote_consistency (Corresponds to flag)
	nodes[0].SetTestingFlag(t, vtcp.FlagForbidSendMessageVoteConsistency, "", "")
	nodes[1].SetTestingFlag(t, vtcp.FlagForbidSendMessageRecoveryStage, "", "")

	uuid, err = nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "100", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	time.Sleep(5 * time.Minute)

	nodes[0].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	nodes[1].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[2].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[3].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[4].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[5].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	nodes[6].CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)

	nodes[0].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 2, 14, 0, 2)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 2, 7, 2, 2)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 2, 14, 2, 2)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 2, 14, 2, 2)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 2, 14, 2, 2)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 2, 14, 2, 2)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 2, 14, 2, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, true, 0)
	nodes[2].CheckSerializedTransaction(t, false, 0)
	nodes[3].CheckSerializedTransaction(t, false, 0)
	nodes[4].CheckSerializedTransaction(t, false, 0)
	nodes[5].CheckSerializedTransaction(t, false, 0)
	nodes[6].CheckSerializedTransaction(t, false, 0)

	nodes[1].SetSettlementLine(t, nodes[0], testconfig.Equivalent, "4000", vtcp.StatusOK)
	time.Sleep(15 * time.Second)
	nodes[2].SetSettlementLine(t, nodes[1], testconfig.Equivalent, "3000", vtcp.StatusOK)

	// Waiting for 15 minutes to simulate the time period of the recovery stage with observing
	time.Sleep(15 * time.Minute)

	nodes[0].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 2, 14, 0, 2)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 2, 14, 2, 2)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 2, 14, 2, 2)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 2, 14, 2, 2)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 2, 14, 2, 2)
	nodes[5].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 2, 14, 2, 2)
	nodes[6].CheckPaymentTransaction(t, vtcp.PaymentObservingStateCommitted, 2, 14, 2, 0)

	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, false, 0)
	nodes[2].CheckSerializedTransaction(t, false, 0)
	nodes[3].CheckSerializedTransaction(t, false, 0)
	nodes[4].CheckSerializedTransaction(t, false, 0)
	nodes[5].CheckSerializedTransaction(t, false, 0)
	nodes[6].CheckSerializedTransaction(t, false, 0)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	nodes[0].SetTestingFlag(t, 0, "", "")
	nodes[1].SetTestingFlag(t, 0, "", "")

	uuid, err = nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[6], testconfig.Equivalent, "100", testconfig.Equivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)
}
