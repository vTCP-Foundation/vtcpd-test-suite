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
	keysSharingNextCntPaymentsAuditRuleSettlementLineNextNodeIndex = 1
)

func getNextIPForKeysSharingNextCntPaymentsAuditRuleSettlementLineTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForKeysSharingNextCntPaymentsAuditRuleSettlementLineTest, keysSharingNextCntPaymentsAuditRuleSettlementLineNextNodeIndex)
	keysSharingNextCntPaymentsAuditRuleSettlementLineNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForKeysSharingNextCntPaymentsAuditRuleSettlementLineTest(t *testing.T, count int) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, count)
	for i := range count {
		nodes[i] = vtcp.NewNode(t, getNextIPForKeysSharingNextCntPaymentsAuditRuleSettlementLineTest(), fmt.Sprintf("node%c", 'A'+i))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings) // Assuming clusterSettings is defined globally or passed
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes)

	nodes[0].OpenChannelAndCheck(t, nodes[1])
	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "1000")
	nodes[0].SetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "2000")

	return nodes, cluster
}

func TestSettlementLineKeysSharing(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	for x := 0; x < vtcp.DefaultAuditPaymentsCount-1; x++ {
		nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "50", vtcp.StatusOK)
		time.Sleep(1 * time.Second)
	}

	for x := 0; x < vtcp.DefaultKeysCount-vtcp.DefaultCriticalKeysCount-(vtcp.DefaultAuditPaymentsCount-1)-3-1; x++ {
		nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "3000", vtcp.StatusOK)
	}

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "50", vtcp.StatusOK)
	time.Sleep(vtcp.DefaultKeysSharingSeconds * time.Second)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineKeysPresence(t, nodeB, true, true)
	nodeB.CheckSettlementLineKeysPresence(t, nodeA, true, true)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
}

func TestSettlementLinesKeysSharingByPaymentOnCoordinator(t *testing.T) {
	nodes, _ := setupNodesForKeysSharingNextCntPaymentsAuditRuleSettlementLineTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	for x := 0; x < vtcp.DefaultAuditPaymentsCount-1; x++ {
		nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "50", vtcp.StatusOK)
		time.Sleep(1 * time.Second)
	}

	for x := 0; x < vtcp.DefaultKeysCount-vtcp.DefaultCriticalKeysCount-(vtcp.DefaultAuditPaymentsCount-1)-3-1; x++ {
		nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "3000", vtcp.StatusOK)
		time.Sleep(500 * time.Millisecond)
	}

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "50", vtcp.StatusOK)
	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "50", vtcp.StatusOK)
	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "50", vtcp.StatusOK)
	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "50", vtcp.StatusOK)
	time.Sleep(15 * time.Second)
	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckSettlementLineKeysPresence(t, nodeB, true, true)
	nodeB.CheckSettlementLineKeysPresence(t, nodeA, true, true)
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)

	// TODO : not finished
	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "50", vtcp.StatusOK)
	// nodeA.CheckValidKeys(t)
	// nodeB.CheckValidKeys(t)
}
