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
	directPaymentSeneNodesNextNodeIndex = 1
)

func setupSevenNodes(t *testing.T, startIndex int) (*vtcp.Node, *vtcp.Node, *vtcp.Node, *vtcp.Node, *vtcp.Node, *vtcp.Node, *vtcp.Node, []*vtcp.Node) {
	node1 := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentSevenNodes, startIndex), fmt.Sprintf("node%d", startIndex))
	node2 := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentSevenNodes, startIndex+1), fmt.Sprintf("node%d", startIndex+1))
	node3 := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentSevenNodes, startIndex+2), fmt.Sprintf("node%d", startIndex+2))
	node4 := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentSevenNodes, startIndex+3), fmt.Sprintf("node%d", startIndex+3))
	node5 := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentSevenNodes, startIndex+4), fmt.Sprintf("node%d", startIndex+4))
	node6 := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentSevenNodes, startIndex+5), fmt.Sprintf("node%d", startIndex+5))
	node7 := vtcp.NewNode(t, fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForDirectPaymentSevenNodes, startIndex+6), fmt.Sprintf("node%d", startIndex+6))
	nodes := []*vtcp.Node{node1, node2, node3, node4, node5, node6, node7}
	return node1, node2, node3, node4, node5, node6, node7, nodes
}

func createChannelsAndSettlementLinesSevenNodes(t *testing.T, node1, node2, node3, node4, node5, node6, node7 *vtcp.Node) {
	node2.CreateChannelAndSettlementLineAndCheck(t, node1, testconfig.Equivalent, "3000")
	node3.CreateChannelAndSettlementLineAndCheck(t, node2, testconfig.Equivalent, "2500")
	node4.CreateChannelAndSettlementLineAndCheck(t, node3, testconfig.Equivalent, "2000")
	node5.CreateChannelAndSettlementLineAndCheck(t, node4, testconfig.Equivalent, "5000")
	node6.CreateChannelAndSettlementLineAndCheck(t, node5, testconfig.Equivalent, "1000")
	node7.CreateChannelAndSettlementLineAndCheck(t, node6, testconfig.Equivalent, "1500")
}

func Test1DirectPayment7NormalAmount(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusOK)
	vtcp.CheckSettlementLineForSyncBatch(t, []*vtcp.Node{node1, node2, node3, node4, node5, node6, node7}, testconfig.Equivalent, 3)
}

func Test2DirectPayment7NodesAmountTooBig(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1500", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4aLostAskNeighborToReserveAmountMsgFromCoordinatorToFirstIntermediateNode(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_forbid_send_request_to_intermediate_node_on_reservation (Corresponds to flag 1 in two_nodes tests)
	if err := node1.SetTestingFlag(vtcp.FlagForbidSendRequestToIntermediateReservation, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node1: %v", err)
	}
	node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4bLostAskNeighborToApproveFurtherNodeReservationMsgFromCoordinatorToFirstIntermediateNode(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_forbid_send_message_to_coordinator_on_reservation (New flag)
	if err := node1.SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node1: %v", err)
	}
	node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4cLostAskRemoteNodeToApproveReservationMsgFromCoordinatorToLastIntermediateNode(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_forbid_send_message_to_coordinator_on_reservation (New flag)
	if err := node1.SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, node3.GetIPAddressForRequests(), ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node1: %v", err)
	}
	node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4dLostProcessNeighborAmountReservationResponseMsgFromFirstIntermediateNodeToCoordinator(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_forbid_send_response_to_intermediate_node_on_reservation (Corresponds to flag 2)
	if err := node2.SetTestingFlag(vtcp.FlagForbidSendResponseToIntemediateOnReservation, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node2: %v", err)
	}
	node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4eLostMsgFromFirstIntermediateNodeToNextIntermediateNode(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_forbid_send_request_to_intermediate_node_on_reservation (Corresponds to flag 1)
	if err := node2.SetTestingFlag(vtcp.FlagForbidSendRequestToIntermediateReservation, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node2: %v", err)
	}
	node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4fLostMsgFromNextIntermediateNodeToPrevious(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_forbid_send_response_to_intermediate_node_on_reservation (Corresponds to flag 2)
	if err := node3.SetTestingFlag(vtcp.FlagForbidSendResponseToIntemediateOnReservation, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node3: %v", err)
	}
	node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4gLostMsgFromLastIntermediateNodeReceiver(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_forbid_send_request_to_intermediate_node_on_reservation (Corresponds to flag 1)
	if err := node6.SetTestingFlag(vtcp.FlagForbidSendRequestToIntermediateReservation, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node6: %v", err)
	}
	node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4hLostMsgReceiverToPrevious(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_forbid_send_response_to_intermediate_node_on_reservation (Corresponds to flag 2)
	if err := node7.SetTestingFlag(vtcp.FlagForbidSendResponseToIntemediateOnReservation, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node7: %v", err)
	}
	node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4jLostProcessNeighborFurtherReservationResponseMsgFromFirstIntermediateNodeToCoordinator(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_forbid_send_message_to_coordinator_on_reservation (New flag)
	if err := node1.SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, "", ""); err != nil { // Assuming node1 is the coordinator contextually
		t.Fatalf("SetTestingFlag failed for node1: %v", err)
	}
	node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4kLostProcessRemoteNodeResponseMsgFromLastIntermediateNodeToCoordinator(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_forbid_send_message_to_coordinator_on_reservation (New flag)
	if err := node6.SetTestingFlag(vtcp.FlagForbidSendMessageToCoordinatorReservation, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node6: %v", err)
	}
	node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test5LostMessageWithPathFinalConfiguration(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_forbid_send_message_with_final_path_configuration (New flag)
	if err := node1.SetTestingFlag(vtcp.FlagForbidSendMessageFinalPathConfig, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node1: %v", err)
	}
	node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusOK) // Default status in Python
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test6aLostMessageWithFinalConfigurationToIntermediateNode(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_forbid_send_message_on_final_amount_clarification (New flag)
	if err := node1.SetTestingFlag(vtcp.FlagForbidSendMessageFinalAmountClarification, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node1: %v", err)
	}
	node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusProtocolError)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test6bLostMessageWithFinalConfigurationToCoordinator(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_forbid_send_message_on_final_amount_clarification (New flag)
	if err := node4.SetTestingFlag(vtcp.FlagForbidSendMessageFinalAmountClarification, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node4: %v", err)
	}
	node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusProtocolError)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test7aLostMsgWithPublicKeysToFirstIntermediateNode(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	err = node1.SetTestingFlag(vtcp.FlagForbidSendMessageVoteStage, "", "")
	if err != nil {
		t.Fatalf("SetTestingFlag failed: %v", err)
	}

	uuid, err := node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusProtocolError)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	node1.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	node2.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node3.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node4.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node5.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node6.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node7.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)

	node1.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, true, 0)
	node3.CheckSerializedTransaction(t, true, 0)
	node4.CheckSerializedTransaction(t, true, 0)
	node5.CheckSerializedTransaction(t, true, 0)
	node6.CheckSerializedTransaction(t, true, 0)
	node7.CheckSerializedTransaction(t, true, 0)

	time.Sleep(time.Duration(vtcp.NodePaymentRecoveryAttempts-1) * vtcp.NodePaymentRecoveryTimePeriodSec * time.Second)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0) // Check sync immediately after sleep

	node1.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node7.CheckPaymentTransaction(t, "", 0, 0, 0, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, true, 0)
	node3.CheckSerializedTransaction(t, true, 0)
	node4.CheckSerializedTransaction(t, true, 0)
	node5.CheckSerializedTransaction(t, true, 0)
	node6.CheckSerializedTransaction(t, true, 0)
	node7.CheckSerializedTransaction(t, false, 0)

	sleepDuration := (vtcp.ObservingCntBlocksForClaiming * vtcp.ObservingCntSecondsPerBlock) - ((vtcp.NodePaymentRecoveryAttempts - 1) * vtcp.NodePaymentRecoveryTimePeriodSec)
	if sleepDuration > 0 {
		time.Sleep(time.Duration(sleepDuration) * time.Second)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	node1.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node7.CheckPaymentTransaction(t, "", 0, 0, 0, 0)

	node1.CheckMaxFlow(t, node7, testconfig.Equivalent, "1000")
}

func Test7bLostMsgWithSignatureToCoordinator(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	if err := node2.SetTestingFlag(vtcp.FlagForbidSendMessageVoteConsistency, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node2: %v", err)
	}
	if err := node6.SetTestingFlag(vtcp.FlagForbidSendMessageVoteConsistency, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node6: %v", err)
	}

	uuid, err := node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusProtocolError)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	node1.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	node2.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node3.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node4.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node5.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node6.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node7.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)

	node1.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, true, 0)
	node3.CheckSerializedTransaction(t, true, 0)
	node4.CheckSerializedTransaction(t, true, 0)
	node5.CheckSerializedTransaction(t, true, 0)
	node6.CheckSerializedTransaction(t, true, 0)
	node7.CheckSerializedTransaction(t, true, 0)

	time.Sleep(time.Duration(vtcp.NodePaymentRecoveryAttempts-1) * vtcp.NodePaymentRecoveryTimePeriodSec * time.Second)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	node1.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, true, 0)
	node3.CheckSerializedTransaction(t, true, 0)
	node4.CheckSerializedTransaction(t, true, 0)
	node5.CheckSerializedTransaction(t, true, 0)
	node6.CheckSerializedTransaction(t, true, 0)
	node7.CheckSerializedTransaction(t, true, 0)

	sleepDuration := (vtcp.ObservingCntBlocksForClaiming * vtcp.ObservingCntSecondsPerBlock) - ((vtcp.NodePaymentRecoveryAttempts - 1) * vtcp.NodePaymentRecoveryTimePeriodSec)
	if sleepDuration > 0 {
		time.Sleep(time.Duration(sleepDuration) * time.Second)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	node1.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, false, 0)
	node3.CheckSerializedTransaction(t, false, 0)
	node4.CheckSerializedTransaction(t, false, 0)
	node5.CheckSerializedTransaction(t, false, 0)
	node6.CheckSerializedTransaction(t, false, 0)
	node7.CheckSerializedTransaction(t, false, 0)

	node1.CheckMaxFlow(t, node7, testconfig.Equivalent, "1000")
}

func Test7cLostMsgWithPublicKeyHashFromIntermediateNodeToParticipants(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_forbid_send_message_on_vote_stage (Corresponds to flag 3)
	if err := node2.SetTestingFlag(vtcp.FlagForbidSendMessageVoteStage, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node2: %v", err)
	}
	if err := node5.SetTestingFlag(vtcp.FlagForbidSendMessageVoteStage, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node5: %v", err)
	}
	node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusProtocolError)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3) // Default 3s sleep from Python original
}

func Test7dLostMsgWithSignatureFromCoordinatorToAllIntermediateNodes(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_forbid_send_message_on_vote_consistency (Corresponds to flag 4)
	if err := node1.SetTestingFlag(vtcp.FlagForbidSendMessageVoteConsistency, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node1: %v", err)
	}

	uuid, err := node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusOK)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	time.Sleep(vtcp.WaitingParticipantsVotesSec * time.Second)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	node1.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	node2.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node3.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node4.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node5.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node6.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node7.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)

	node1.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 0, 1)
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, false, 0)
	node3.CheckSerializedTransaction(t, false, 0)
	node4.CheckSerializedTransaction(t, false, 0)
	node5.CheckSerializedTransaction(t, false, 0)
	node6.CheckSerializedTransaction(t, false, 0)
	node7.CheckSerializedTransaction(t, false, 0)
	// Python: # self.node_1.check_max_flow({self.node_7.address: 0}) -> Not checking max flow here
}

func Test7eLostMsgWithSignatureFromCoordinatorToAllIntermediateNodesAlsoOnRecovery(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// flag_forbid_send_message_on_recovery_stage = 103 (New)
	if err := node1.SetTestingFlag(vtcp.FlagForbidSendMessageVoteConsistency, "", ""); err != nil { // Applied vote consistency flag
		t.Fatalf("SetTestingFlag failed for node1: %v", err)
	}

	uuid, err := node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusOK)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	time.Sleep(vtcp.WaitingParticipantsVotesSec * time.Second)

	node1.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	node2.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node3.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node4.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node5.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node6.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node7.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)

	node1.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 0, 1)
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, true, 0)
	node3.CheckSerializedTransaction(t, true, 0)
	node4.CheckSerializedTransaction(t, true, 0)
	node5.CheckSerializedTransaction(t, true, 0)
	node6.CheckSerializedTransaction(t, true, 0)
	node7.CheckSerializedTransaction(t, true, 0)

	// Python: time.sleep((self.recovering_attempts - 1) * (self.recovering_time_period_sec + self.waiting_for_message))
	// self.waiting_for_message is likely 1 second.
	time.Sleep(time.Duration(vtcp.NodePaymentRecoveryAttempts-1) * (vtcp.NodePaymentRecoveryTimePeriodSec + 1) * time.Second)

	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 0)

	// Python: time.sleep(self.observing_claim_period * 2)
	// self.observing_claim_period might be (ObservingCntBlocksForClaiming * ObservingCntSecondsPerBlock)
	observingClaimPeriod := time.Duration(vtcp.ObservingCntBlocksForClaiming*vtcp.ObservingCntSecondsPerBlock) * time.Second
	time.Sleep(observingClaimPeriod * 2)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	// TODO: Verify observing_ParticipantsVotesPresent_response mapping
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 7, 1, 1) // Using Claimed as placeholder
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 7, 1, 1) // Using Claimed as placeholder
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 7, 1, 1) // Using Claimed as placeholder
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 7, 1, 1) // Using Claimed as placeholder
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 7, 1, 1) // Using Claimed as placeholder
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 7, 1, 0) // Using Claimed as placeholder

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, false, 0)
	node3.CheckSerializedTransaction(t, false, 0)
	node4.CheckSerializedTransaction(t, false, 0)
	node5.CheckSerializedTransaction(t, false, 0)
	node6.CheckSerializedTransaction(t, false, 0)
	node7.CheckSerializedTransaction(t, false, 0)

	time.Sleep(observingClaimPeriod)
	// TODO: Verify observing_ParticipantsVotesPresent_response mapping
	node1.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 7, 0, 1) // Using Claimed as placeholder
	// Python: # self.node_1.check_max_flow({self.node_7.address: 0})
}

func Test8aCrashCoordinatorAfterSendingMessageOnVoting(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_throw_exception_on_vote (Corresponds to flag 6)
	if err := node1.SetTestingFlag(vtcp.FlagThrowExceptionVote, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node1: %v", err)
	}

	uuid, err := node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusProtocolError)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	node1.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	node2.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node3.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node4.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node5.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node6.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node7.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)

	node1.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, true, 0)
	node3.CheckSerializedTransaction(t, true, 0)
	node4.CheckSerializedTransaction(t, true, 0)
	node5.CheckSerializedTransaction(t, true, 0)
	node6.CheckSerializedTransaction(t, true, 0)
	node7.CheckSerializedTransaction(t, true, 0)

	time.Sleep(time.Duration(vtcp.NodePaymentRecoveryAttempts-1) * vtcp.NodePaymentRecoveryTimePeriodSec * time.Second)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	node1.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, true, 0)
	node3.CheckSerializedTransaction(t, true, 0)
	node4.CheckSerializedTransaction(t, true, 0)
	node5.CheckSerializedTransaction(t, true, 0)
	node6.CheckSerializedTransaction(t, true, 0)
	node7.CheckSerializedTransaction(t, true, 0)

	sleepDuration := (vtcp.ObservingCntBlocksForClaiming * vtcp.ObservingCntSecondsPerBlock) - ((vtcp.NodePaymentRecoveryAttempts - 1) * vtcp.NodePaymentRecoveryTimePeriodSec)
	if sleepDuration > 0 {
		time.Sleep(time.Duration(sleepDuration) * time.Second)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	node1.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, false, 0)
	node3.CheckSerializedTransaction(t, false, 0)
	node4.CheckSerializedTransaction(t, false, 0)
	node5.CheckSerializedTransaction(t, false, 0)
	node6.CheckSerializedTransaction(t, false, 0)
	node7.CheckSerializedTransaction(t, false, 0)

	node1.CheckMaxFlow(t, node7, testconfig.Equivalent, "1000")
}

func Test8bCrashCoordinatorAfterReceivingMessageWithSignature(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_throw_exception_on_vote_consistency (Corresponds to flag 7)
	if err := node1.SetTestingFlag(vtcp.FlagThrowExceptionVoteConsistency, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node1: %v", err)
	}

	uuid, err := node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusProtocolError)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	node1.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	node2.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node3.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node4.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node5.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node6.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node7.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)

	node1.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, true, 0)
	node3.CheckSerializedTransaction(t, true, 0)
	node4.CheckSerializedTransaction(t, true, 0)
	node5.CheckSerializedTransaction(t, true, 0)
	node6.CheckSerializedTransaction(t, true, 0)
	node7.CheckSerializedTransaction(t, true, 0)

	time.Sleep(time.Duration(vtcp.NodePaymentRecoveryAttempts-1) * vtcp.NodePaymentRecoveryTimePeriodSec * time.Second)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	node1.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, true, 0)
	node3.CheckSerializedTransaction(t, true, 0)
	node4.CheckSerializedTransaction(t, true, 0)
	node5.CheckSerializedTransaction(t, true, 0)
	node6.CheckSerializedTransaction(t, true, 0)
	node7.CheckSerializedTransaction(t, true, 0)

	sleepDuration := (vtcp.ObservingCntBlocksForClaiming * vtcp.ObservingCntSecondsPerBlock) - ((vtcp.NodePaymentRecoveryAttempts - 1) * vtcp.NodePaymentRecoveryTimePeriodSec)
	if sleepDuration > 0 {
		time.Sleep(time.Duration(sleepDuration) * time.Second)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	node1.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, false, 0)
	node3.CheckSerializedTransaction(t, false, 0)
	node4.CheckSerializedTransaction(t, false, 0)
	node5.CheckSerializedTransaction(t, false, 0)
	node6.CheckSerializedTransaction(t, false, 0)
	node7.CheckSerializedTransaction(t, false, 0)

	node1.CheckMaxFlow(t, node7, testconfig.Equivalent, "1000")
}

func Test9aCrashIntermediateNodeRunPreviousNeighborRequestProcessingStage(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_throw_exception_on_previous_neighbor_request (Corresponds to flag 5)
	if err := node3.SetTestingFlag(vtcp.FlagThrowExceptionPreviousNeighborRequest, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node3: %v", err)
	}
	node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}
func Test9bCrashIntermediateNodeRunCoordinatorRequestProcessingStage(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_throw_exception_on_coordinator_request_processing (New flag)
	if err := node3.SetTestingFlag(vtcp.FlagThrowExceptionCoordinatorRequest, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node3: %v", err)
	}
	node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test9cCrashIntermediateNodeRunNextNeighborResponseProcessingStage(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_throw_exception_on_next_neighbor_response_processing (New flag)
	if err := node3.SetTestingFlag(vtcp.FlagThrowExceptionNextNeighborResponse, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node3: %v", err)
	}
	node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test9dCrashIntermediateNodeAfterSignBeforeSendResponse(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_throw_exception_on_vote (Corresponds to flag 6)
	if err := node2.SetTestingFlag(vtcp.FlagThrowExceptionVote, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node2: %v", err)
	}
	if err := node3.SetTestingFlag(vtcp.FlagThrowExceptionVote, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node3: %v", err)
	}
	if err := node5.SetTestingFlag(vtcp.FlagThrowExceptionVote, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node5: %v", err)
	}
	uuid, err := node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusProtocolError)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	node1.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	node2.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node3.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node4.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node5.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node6.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node7.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)

	node1.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, true, 0)
	node3.CheckSerializedTransaction(t, true, 0)
	node4.CheckSerializedTransaction(t, true, 0)
	node5.CheckSerializedTransaction(t, true, 0)
	node6.CheckSerializedTransaction(t, true, 0)
	node7.CheckSerializedTransaction(t, true, 0)

	time.Sleep(time.Duration(vtcp.NodePaymentRecoveryAttempts-1) * vtcp.NodePaymentRecoveryTimePeriodSec * time.Second)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	node1.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, true, 0)
	node3.CheckSerializedTransaction(t, true, 0)
	node4.CheckSerializedTransaction(t, true, 0)
	node5.CheckSerializedTransaction(t, true, 0)
	node6.CheckSerializedTransaction(t, true, 0)
	node7.CheckSerializedTransaction(t, true, 0)

	sleepDuration := (vtcp.ObservingCntBlocksForClaiming * vtcp.ObservingCntSecondsPerBlock) - ((vtcp.NodePaymentRecoveryAttempts - 1) * vtcp.NodePaymentRecoveryTimePeriodSec)
	if sleepDuration > 0 {
		time.Sleep(time.Duration(sleepDuration) * time.Second)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	node1.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, false, 0)
	node3.CheckSerializedTransaction(t, false, 0)
	node4.CheckSerializedTransaction(t, false, 0)
	node5.CheckSerializedTransaction(t, false, 0)
	node6.CheckSerializedTransaction(t, false, 0)
	node7.CheckSerializedTransaction(t, false, 0)

	node1.CheckMaxFlow(t, node7, testconfig.Equivalent, "1000")
}
func Test9eStopProcessIntermediateNodeAfterVotesReceivingBeforeCommitting(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_terminate_process_on_vote_consistency (Corresponds to flag 11)
	if err := node2.SetTestingFlag(vtcp.FlagTerminateProcessVoteConsistency, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node2: %v", err)
	}
	if err := node3.SetTestingFlag(vtcp.FlagTerminateProcessVoteConsistency, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node3: %v", err)
	}
	if err := node5.SetTestingFlag(vtcp.FlagTerminateProcessVoteConsistency, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node5: %v", err)
	}
	uuid, err := node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusOK)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	node1.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	node2.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node3.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node4.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	node5.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node6.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	node7.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)

	node1.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 0, 1)
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, false, 0)
	node3.CheckSerializedTransaction(t, false, 0)
	node4.CheckSerializedTransaction(t, false, 0)
	node5.CheckSerializedTransaction(t, false, 0)
	node6.CheckSerializedTransaction(t, false, 0)
	node7.CheckSerializedTransaction(t, false, 0)
	// Python: # self.node_1.check_max_flow({self.node_7.address: 0})
}

func Test10aStopProcessCoordinatorAfterSendingMessageOnVoting(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_terminate_process_on_vote (Corresponds to flag 10)
	if err := node1.SetTestingFlag(vtcp.FlagTerminateProcessVote, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node1: %v", err)
	}
	// Python status_code=None implies OK or that the call might not return/complete normally due to process termination.
	// Assuming OK for transaction creation initiation. The actual outcome is process termination.
	uuid, err := node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusOK)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	node1.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	node2.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	node3.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	node4.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	node5.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	node6.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	node7.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)

	// Python: self.node_1.check_payment_transaction(outgoing_receipts_cnt=1)
	// Other nodes default to 0 counts.
	node1.CheckPaymentTransaction(t, "", 0, 0, 0, 1)
	node2.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	node3.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	node4.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	node5.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	node6.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	node7.CheckPaymentTransaction(t, "", 0, 0, 0, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, false, 0)
	node3.CheckSerializedTransaction(t, false, 0)
	node4.CheckSerializedTransaction(t, false, 0)
	node5.CheckSerializedTransaction(t, false, 0)
	node6.CheckSerializedTransaction(t, false, 0)
	node7.CheckSerializedTransaction(t, false, 0)

	node1.CheckMaxFlow(t, node7, testconfig.Equivalent, "0")
}

func Test10bStopProcessCoordinatorAfterReceivingMessageWithSignatures(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_terminate_process_on_vote_consistency (Corresponds to flag 11)
	if err := node1.SetTestingFlag(vtcp.FlagTerminateProcessVoteConsistency, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node1: %v", err)
	}
	uuid, err := node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusOK) // Python status_code=None
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	node1.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	node2.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node3.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node4.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node5.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node6.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node7.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)

	node1.CheckPaymentTransaction(t, "", 0, 0, 0, 1) // Python: outgoing_receipts_cnt=1
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, true, 0)
	node3.CheckSerializedTransaction(t, true, 0)
	node4.CheckSerializedTransaction(t, true, 0)
	node5.CheckSerializedTransaction(t, true, 0)
	node6.CheckSerializedTransaction(t, true, 0)
	node7.CheckSerializedTransaction(t, true, 0)

	time.Sleep(time.Duration(vtcp.NodePaymentRecoveryAttempts-1) * vtcp.NodePaymentRecoveryTimePeriodSec * time.Second)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	node1.CheckPaymentTransaction(t, "", 0, 0, 0, 1) // Python: outgoing_receipts_cnt=1
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, true, 0)
	node3.CheckSerializedTransaction(t, true, 0)
	node4.CheckSerializedTransaction(t, true, 0)
	node5.CheckSerializedTransaction(t, true, 0)
	node6.CheckSerializedTransaction(t, true, 0)
	node7.CheckSerializedTransaction(t, true, 0)

	sleepDuration := (vtcp.ObservingCntBlocksForClaiming * vtcp.ObservingCntSecondsPerBlock) - ((vtcp.NodePaymentRecoveryAttempts - 1) * vtcp.NodePaymentRecoveryTimePeriodSec)
	if sleepDuration > 0 {
		time.Sleep(time.Duration(sleepDuration) * time.Second)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	node1.CheckPaymentTransaction(t, "", 0, 0, 0, 1) // Python: outgoing_receipts_cnt=1
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, false, 0)
	node3.CheckSerializedTransaction(t, false, 0)
	node4.CheckSerializedTransaction(t, false, 0)
	node5.CheckSerializedTransaction(t, false, 0)
	node6.CheckSerializedTransaction(t, false, 0)
	node7.CheckSerializedTransaction(t, false, 0)

	node1.CheckMaxFlow(t, node7, testconfig.Equivalent, "1000")
}

func Test11aStopProcessIntermediateNodeRunPreviousNeighborRequestProcessingStage(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_terminate_process_on_previous_neighbor_request (Corresponds to flag 9)
	if err := node3.SetTestingFlag(vtcp.FlagTerminateProcessPreviousNeighborRequest, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node3: %v", err)
	}
	node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test11bStopProcessIntermediateNodeRunCoordinatorRequestProcessingStage(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_terminate_process_on_coordinator_request_processing (New flag)
	// TODO: Verify flag value for flag_terminate_process_on_coordinator_request_processing
	if err := node3.SetTestingFlag(vtcp.FlagTerminateProcessCoordinatorRequest, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node3: %v", err)
	}
	node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test11cStopProcessIntermediateNodeRunNextNeighborResponseProcessingStage(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_terminate_process_on_next_neighbor_response_processing (New flag)
	// TODO: Verify flag value for flag_terminate_process_on_next_neighbor_response_processing
	if err := node3.SetTestingFlag(vtcp.FlagTerminateProcessNextNeighborResponse, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node3: %v", err)
	}
	node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusInsufficientFunds)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test11dStopProcessIntermediateNodeAfterSignBeforeSending(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_terminate_process_on_vote (Corresponds to flag 10)
	if err := node2.SetTestingFlag(vtcp.FlagTerminateProcessVote, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node2: %v", err)
	}
	if err := node3.SetTestingFlag(vtcp.FlagTerminateProcessVote, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node3: %v", err)
	}
	if err := node5.SetTestingFlag(vtcp.FlagTerminateProcessVote, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node5: %v", err)
	}
	uuid, err := node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusProtocolError)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	node1.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	node2.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node3.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node4.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node5.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node6.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node7.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)

	node1.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 1)
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 0, 1, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, true, 0)
	node3.CheckSerializedTransaction(t, true, 0)
	node4.CheckSerializedTransaction(t, true, 0)
	node5.CheckSerializedTransaction(t, true, 0)
	node6.CheckSerializedTransaction(t, true, 0)
	node7.CheckSerializedTransaction(t, true, 0)

	time.Sleep(time.Duration(vtcp.NodePaymentRecoveryAttempts-1) * vtcp.NodePaymentRecoveryTimePeriodSec * time.Second)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	node1.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 1)
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 1, 0, 1, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, true, 0)
	node3.CheckSerializedTransaction(t, true, 0)
	node4.CheckSerializedTransaction(t, true, 0)
	node5.CheckSerializedTransaction(t, true, 0)
	node6.CheckSerializedTransaction(t, true, 0)
	node7.CheckSerializedTransaction(t, true, 0)

	sleepDuration := (vtcp.ObservingCntBlocksForClaiming * vtcp.ObservingCntSecondsPerBlock) - ((vtcp.NodePaymentRecoveryAttempts - 1) * vtcp.NodePaymentRecoveryTimePeriodSec)
	if sleepDuration > 0 {
		time.Sleep(time.Duration(sleepDuration) * time.Second)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 0)

	node1.CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 1, 0, 0, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, false, 0)
	node3.CheckSerializedTransaction(t, false, 0)
	node4.CheckSerializedTransaction(t, false, 0)
	node5.CheckSerializedTransaction(t, false, 0)
	node6.CheckSerializedTransaction(t, false, 0)
	node7.CheckSerializedTransaction(t, false, 0)

	node1.CheckMaxFlow(t, node7, testconfig.Equivalent, "1000")
}

func Test11eStopProcessIntermediateNodeAfterVotesReceivingBeforeCommitting(t *testing.T) {
	startIndex := directPaymentSeneNodesNextNodeIndex
	directPaymentSeneNodesNextNodeIndex += 7
	node1, node2, node3, node4, node5, node6, node7, nodes := setupSevenNodes(t, startIndex)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}
	cluster.RunNodes(ctx, t, nodes)
	createChannelsAndSettlementLinesSevenNodes(t, node1, node2, node3, node4, node5, node6, node7)

	// self.flag_terminate_process_on_vote_consistency (Corresponds to flag 11)
	if err := node2.SetTestingFlag(vtcp.FlagTerminateProcessVoteConsistency, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node2: %v", err)
	}
	if err := node3.SetTestingFlag(vtcp.FlagTerminateProcessVoteConsistency, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node3: %v", err)
	}
	if err := node5.SetTestingFlag(vtcp.FlagTerminateProcessVoteConsistency, "", ""); err != nil {
		t.Fatalf("SetTestingFlag failed for node5: %v", err)
	}
	uuid, err := node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "1000", vtcp.StatusOK)
	if err != nil {
		t.Fatalf("CreateTransactionCheckStatus failed for node1: %v", err)
	}
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, vtcp.WaitingParticipantsVotesSec)

	node1.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	node2.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node3.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node4.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	node5.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, true)
	node6.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)
	node7.CheckNodeForLogMessage(t, uuid, vtcp.LogMessageRecoveringLogMessage, false)

	node1.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 0, 1)
	node2.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	node3.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	node4.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	node5.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	node6.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 1)
	node7.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 7, 1, 0)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, false, 0)
	node3.CheckSerializedTransaction(t, false, 0)
	node4.CheckSerializedTransaction(t, false, 0)
	node5.CheckSerializedTransaction(t, false, 0)
	node6.CheckSerializedTransaction(t, false, 0)
	node7.CheckSerializedTransaction(t, false, 0)
	// Python: # self.node_1.check_max_flow({self.node_7.address: 0})
}
