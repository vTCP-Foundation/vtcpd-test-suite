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
	settlementLineArchivedNextNodeIndex = 1
)

func getNextIPForSettlementLineArchivedTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForSettlementLineArchivedTest, settlementLineArchivedNextNodeIndex)
	settlementLineArchivedNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForSettlementLineArchivedTest(t *testing.T, count int) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, count)
	for i := range count {
		nodes[i] = vtcp.NewNode(t, getNextIPForSettlementLineArchivedTest(), fmt.Sprintf("node%c", 'A'+i))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings) // Assuming clusterSettings is defined globally or passed
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes)

	nodes[0].OpenChannelAndCheck(t, nodes[1])
	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "1000")
	return nodes, cluster
}

func TestSettlementLineArchivedCloseOutgoing(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineArchivedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeB.SetSettlementLine(t, nodeA, testconfig.Equivalent, "0", vtcp.StatusOK)
	time.Sleep(3 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)

	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
}

func TestSettlementLineArchivedCloseIncoming(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineArchivedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.CloseMaxNegativeBalance(t, nodeB, testconfig.Equivalent)
	time.Sleep(3 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)

	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
}

func TestSettlementLineArchivedPayment(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineArchivedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	nodeA.CreateTransactionCheckStatus(t, nodeB, testconfig.Equivalent, "500", vtcp.StatusOK)
	time.Sleep(1 * time.Second)
	nodeB.SetSettlementLine(t, nodeA, testconfig.Equivalent, "0", vtcp.StatusOK)
	time.Sleep(1 * time.Second)
	nodeB.CreateTransactionCheckStatus(t, nodeA, testconfig.Equivalent, "500", vtcp.StatusOK)
	time.Sleep(2 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)

	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
}

func TestSettlementLineArchivedOpenAgain(t *testing.T) {
	nodes, _ := setupNodesForSettlementLineArchivedTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Close the settlement line first
	nodeB.SetSettlementLine(t, nodeA, testconfig.Equivalent, "0", vtcp.StatusOK)
	time.Sleep(1 * time.Second)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateClosed)
	nodeA.CheckValidKeys(t, 0, 0)
	nodeB.CheckValidKeys(t, 0, 0)

	// Open settlement line again
	nodeB.CreateAndSetSettlementLineAndCheck(t, nodeA, testconfig.Equivalent, "1000")

	// Check keys after reopening (keys_count-1 = 10-1 = 9)
	nodeA.CheckValidKeys(t, 9, 9)
	nodeB.CheckValidKeys(t, 9, 9)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)

	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
}
