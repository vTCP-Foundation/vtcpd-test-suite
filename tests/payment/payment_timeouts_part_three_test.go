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
	paymentTimeoutsPartThreeNextNodeIndex = 1
)

func getNextIPForPaymentTimeoutsPartThreeTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForPaymentTimeoutsPartThreeTest, paymentTimeoutsPartThreeNextNodeIndex)
	paymentTimeoutsPartThreeNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForPaymentTimeoutsPartThreeTest(t *testing.T) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, 5)
	for i := range 5 {
		nodes[i] = vtcp.NewNode(t, getNextIPForPaymentTimeoutsPartThreeTest(), fmt.Sprintf("node%d", i+1))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes, false)

	nodes[1].OpenChannelAndCheck(t, nodes[0])
	nodes[3].OpenChannelAndCheck(t, nodes[2])
	nodes[2].OpenChannelAndCheck(t, nodes[1])
	nodes[4].OpenChannelAndCheck(t, nodes[3])

	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "700")
	nodes[2].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "1000")
	nodes[3].CreateAndSetSettlementLineAndCheck(t, nodes[2], testconfig.Equivalent, "800")
	nodes[4].CreateAndSetSettlementLineAndCheck(t, nodes[3], testconfig.Equivalent, "500")

	nodes[0].CheckMaxFlow(t, nodes[4], testconfig.Equivalent, "500")

	return nodes, cluster
}

func TestTimeoutsPartThree1TimesleepIntermediateNodeBeforeLastConfiguration(t *testing.T) {
	nodes, _ := setupNodesForPaymentTimeoutsPartThreeTest(t)

	nodes[2].SetTestingFlag(vtcp.FlagSleepOnFinalAmountClarification, "", "")

	nodes[0].CreateTransactionCheckStatus(t, nodes[4], testconfig.Equivalent, "500", vtcp.StatusNoConsensusError)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 60)
}

func TestTimeoutsPartThree2TimesleepOneIntermediateNodeWhileSigning(t *testing.T) {
	nodes, _ := setupNodesForPaymentTimeoutsPartThreeTest(t)

	nodes[2].SetTestingFlag(vtcp.FlagSleepOnVoteConsistencyStage, "", "")

	nodes[0].CreateTransactionCheckStatus(t, nodes[4], testconfig.Equivalent, "500", vtcp.StatusNoConsensusError)

	time.Sleep(time.Duration(vtcp.WaitingParticipantsVotesSec+12) * time.Second)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 60)

	// Check for recovery log messages
	nodes[0].CheckNodeForLogMessage(t, "", vtcp.LogMessageRecoveringLogMessage, false)
	nodes[1].CheckNodeForLogMessage(t, "", vtcp.LogMessageRecoveringLogMessage, true)
	nodes[2].CheckNodeForLogMessage(t, "", vtcp.LogMessageRecoveringLogMessage, true)
	nodes[3].CheckNodeForLogMessage(t, "", vtcp.LogMessageRecoveringLogMessage, true)
	nodes[4].CheckNodeForLogMessage(t, "", vtcp.LogMessageRecoveringLogMessage, true)

	// Check payment transactions with observing responses
	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 0, 1, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 0, 1, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 0, 1, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 0, 1, 0, 1)

	// Check serialized transactions
	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, true, 0)
	nodes[2].CheckSerializedTransaction(t, true, 0)
	nodes[3].CheckSerializedTransaction(t, true, 0)
	nodes[4].CheckSerializedTransaction(t, true, 0)

	// Wait for recovery attempts
	time.Sleep(time.Duration((vtcp.NodePaymentRecoveryAttempts-1)*vtcp.NodePaymentRecoveryTimePeriodSec) * time.Second)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 60)

	// Check payment transactions after recovery
	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 0, 1, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 0, 1, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 0, 1, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 0, 1, 0, 1)

	// Check serialized transactions after recovery
	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, true, 0)
	nodes[2].CheckSerializedTransaction(t, true, 0)
	nodes[3].CheckSerializedTransaction(t, true, 0)
	nodes[4].CheckSerializedTransaction(t, true, 0)

	// Wait for observing period
	time.Sleep(time.Duration(vtcp.ObservingCntBlocksForClaiming*vtcp.ObservingCntSecondsPerBlock-(vtcp.NodePaymentRecoveryAttempts-1)*vtcp.NodePaymentRecoveryTimePeriodSec) * time.Second)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 60)

	// Check final payment transactions
	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 0, 0, 0, 0)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 0, 0, 0, 0)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 0, 0, 0, 0)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 0, 0, 0, 0)

	// Check final serialized transactions
	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, false, 0)
	nodes[2].CheckSerializedTransaction(t, false, 0)
	nodes[3].CheckSerializedTransaction(t, false, 0)
	nodes[4].CheckSerializedTransaction(t, false, 0)
}

func TestTimeoutsPartThree3TimesleepAllIntermediateNodeBeforeSendingToCoordinator(t *testing.T) {
	nodes, _ := setupNodesForPaymentTimeoutsPartThreeTest(t)

	nodes[1].SetTestingFlag(vtcp.FlagSleepOnVoteConsistencyStage, "", "")
	nodes[2].SetTestingFlag(vtcp.FlagSleepOnVoteConsistencyStage, "", "")
	nodes[3].SetTestingFlag(vtcp.FlagSleepOnVoteConsistencyStage, "", "")
	nodes[4].SetTestingFlag(vtcp.FlagSleepOnVoteConsistencyStage, "", "")

	nodes[0].CreateTransactionCheckStatus(t, nodes[4], testconfig.Equivalent, "500", vtcp.StatusNoConsensusError)

	time.Sleep(time.Duration(vtcp.WaitingParticipantsVotesSec) * time.Second)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 60)

	// Check for recovery log messages
	nodes[0].CheckNodeForLogMessage(t, "", vtcp.LogMessageRecoveringLogMessage, false)
	nodes[1].CheckNodeForLogMessage(t, "", vtcp.LogMessageRecoveringLogMessage, true)
	nodes[2].CheckNodeForLogMessage(t, "", vtcp.LogMessageRecoveringLogMessage, true)
	nodes[3].CheckNodeForLogMessage(t, "", vtcp.LogMessageRecoveringLogMessage, true)
	nodes[4].CheckNodeForLogMessage(t, "", vtcp.LogMessageRecoveringLogMessage, true)

	// Check payment transactions with observing responses
	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 0, 1, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 0, 1, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 0, 1, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 0, 1, 0, 1)

	// Check serialized transactions
	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, true, 0)
	nodes[2].CheckSerializedTransaction(t, true, 0)
	nodes[3].CheckSerializedTransaction(t, true, 0)
	nodes[4].CheckSerializedTransaction(t, true, 0)

	// Wait for recovery attempts
	time.Sleep(time.Duration((vtcp.NodePaymentRecoveryAttempts-1)*vtcp.NodePaymentRecoveryTimePeriodSec) * time.Second)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 60)

	// Check payment transactions after recovery
	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 0, 1, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 0, 1, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 0, 1, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateClaimed, 0, 1, 0, 1)

	// Check serialized transactions after recovery
	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, true, 0)
	nodes[2].CheckSerializedTransaction(t, true, 0)
	nodes[3].CheckSerializedTransaction(t, true, 0)
	nodes[4].CheckSerializedTransaction(t, true, 0)

	// Wait for observing period
	time.Sleep(time.Duration(vtcp.ObservingCntBlocksForClaiming*vtcp.ObservingCntSecondsPerBlock-(vtcp.NodePaymentRecoveryAttempts-1)*vtcp.NodePaymentRecoveryTimePeriodSec) * time.Second)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 60)

	// Check final payment transactions
	nodes[0].CheckPaymentTransaction(t, "", 0, 0, 0, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 0, 0, 0, 0)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 0, 0, 0, 0)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 0, 0, 0, 0)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateRejected, 0, 0, 0, 0)

	// Check final serialized transactions
	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, false, 0)
	nodes[2].CheckSerializedTransaction(t, false, 0)
	nodes[3].CheckSerializedTransaction(t, false, 0)
	nodes[4].CheckSerializedTransaction(t, false, 0)
}

func TestTimeoutsPartThree4TimesleepCoordinatorBeforeSendingFinalConfigurationMessage(t *testing.T) {
	nodes, _ := setupNodesForPaymentTimeoutsPartThreeTest(t)

	nodes[0].SetTestingFlag(vtcp.FlagSleepOnVoteConsistencyStage, "", "")

	nodes[0].CreateTransactionCheckStatus(t, nodes[4], testconfig.Equivalent, "500", vtcp.StatusOK)

	time.Sleep(time.Duration(vtcp.WaitingParticipantsVotesSec+vtcp.NodePaymentRecoveryTimePeriodSec) * time.Second)
	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 60)

	// Check for recovery log messages
	nodes[1].CheckNodeForLogMessage(t, "", vtcp.LogMessageRecoveringLogMessage, true)
	nodes[2].CheckNodeForLogMessage(t, "", vtcp.LogMessageRecoveringLogMessage, true)
	nodes[3].CheckNodeForLogMessage(t, "", vtcp.LogMessageRecoveringLogMessage, true)
	nodes[4].CheckNodeForLogMessage(t, "", vtcp.LogMessageRecoveringLogMessage, true)

	// Check payment transactions with observing responses
	nodes[0].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 5, 0, 1, 0)
	nodes[1].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 5, 1, 1, 1)
	nodes[2].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 5, 1, 1, 1)
	nodes[3].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 5, 1, 1, 1)
	nodes[4].CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 5, 1, 0, 1)

	// Check serialized transactions
	nodes[0].CheckSerializedTransaction(t, false, 0)
	nodes[1].CheckSerializedTransaction(t, false, 0)
	nodes[2].CheckSerializedTransaction(t, false, 0)
	nodes[3].CheckSerializedTransaction(t, false, 0)
	nodes[4].CheckSerializedTransaction(t, false, 0)
}
