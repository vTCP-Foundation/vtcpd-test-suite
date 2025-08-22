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
	setSettlementLineBadInternetNextNodeIndex = 1
)

func getNextIPForSetSettlementLineBadInternetTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForSetSettlementLineBadInternetTest, setSettlementLineBadInternetNextNodeIndex)
	setSettlementLineBadInternetNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForSetSettlementLineBadInternetTest(t *testing.T, count int) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, count)
	for i := range count {
		nodes[i] = vtcp.NewNode(t, getNextIPForSetSettlementLineBadInternetTest(), fmt.Sprintf("node%c", 'A'+i))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings) // Assuming clusterSettings is defined globally or passed
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes)
	nodes[0].OpenChannelAndCheck(t, nodes[1])
	nodes[0].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "1000")

	return nodes, cluster
}

func waitSetSettlementLineActive(t *testing.T, nodeA *vtcp.Node, nodeB *vtcp.Node) {
	currentAttempt := 0
	maxAttempts := 30
	for {
		time.Sleep(5 * time.Second)
		t.Logf("Attempt %d", currentAttempt)
		nodeInfo, stausCode, err := nodeA.GetSettlementsLineInfoByAddress(nodeB, testconfig.Equivalent)
		if err != nil {
			t.Logf("failed to get settlements line info: %v", err)
			currentAttempt++
			if currentAttempt > maxAttempts {
				t.Fatalf("exceeded max attempts to get settlements line info: %v", err)
			}
			continue
		}
		if stausCode != vtcp.StatusOK {
			t.Logf("failed to get settlements line info, wrong response status code: %d", stausCode)
			currentAttempt++
			if currentAttempt > maxAttempts {
				t.Fatalf("exceeded max attempts to get settlements line info: %v", err)
			}
			continue
		}
		if nodeInfo.State != vtcp.SettlementLineStateActive {
			t.Logf("Settlements line state: %s", nodeInfo.State)
			currentAttempt++
			if currentAttempt > maxAttempts {
				t.Fatalf("exceeded max attempts to get settlements line info: %v", err)
			}
			continue
		}
		break
	}
}

func TestSetSettlementLine256kbBandwidthInitiatorNode(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineBadInternetTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Configure network conditions with 256kb bandwidth
	conditions := &vtcp.NetworkConditions{
		Bandwidth: "256kbit",
	}
	err := cluster.ConfigureNetworkConditions(nodeA, conditions, "eth0")
	if err != nil {
		t.Fatalf("failed to configure network conditions: %v", err)
	}

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "2000", vtcp.StatusOK)

	waitSetSettlementLineActive(t, nodeA, nodeB)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t) // Added based on python test
	nodeB.CheckValidKeys(t) // Added based on python test
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckSettlementLineForSync(t, nodeA, testconfig.Equivalent)
}

func TestSetSettlementLine256kbBandwidthContractorNode(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineBadInternetTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Configure network conditions with 256kb bandwidth
	conditions := &vtcp.NetworkConditions{
		Bandwidth: "256kbit",
	}
	err := cluster.ConfigureNetworkConditions(nodeB, conditions, "eth0")
	if err != nil {
		t.Fatalf("failed to configure network conditions: %v", err)
	}

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "2000", vtcp.StatusOK)

	waitSetSettlementLineActive(t, nodeA, nodeB)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t) // Added based on python test
	nodeB.CheckValidKeys(t) // Added based on python test
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckSettlementLineForSync(t, nodeA, testconfig.Equivalent)
}

func TestSetSettlementLine256kbBandwidthBothNodes(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineBadInternetTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Configure network conditions with 256kb bandwidth
	conditions := &vtcp.NetworkConditions{
		Bandwidth: "256kbit",
	}
	err := cluster.ConfigureNetworkConditions(nodeA, conditions, "eth0")
	if err != nil {
		t.Fatalf("failed to configure network conditions: %v", err)
	}
	err = cluster.ConfigureNetworkConditions(nodeB, conditions, "eth0")
	if err != nil {
		t.Fatalf("failed to configure network conditions: %v", err)
	}

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "2000", vtcp.StatusOK)

	waitSetSettlementLineActive(t, nodeA, nodeB)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t) // Added based on python test
	nodeB.CheckValidKeys(t) // Added based on python test
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckSettlementLineForSync(t, nodeA, testconfig.Equivalent)
}

func TestSetSettlementLine10PercentPacketLossInitiatorNode(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineBadInternetTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Configure network conditions with 10% packet loss
	conditions := &vtcp.NetworkConditions{
		LossPercent: 10.0,
	}
	err := cluster.ConfigureNetworkConditions(nodeA, conditions, "eth0")
	if err != nil {
		t.Fatalf("failed to configure network conditions: %v", err)
	}

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "2000", vtcp.StatusOK)

	waitSetSettlementLineActive(t, nodeA, nodeB)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t) // Added based on python test
	nodeB.CheckValidKeys(t) // Added based on python test
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckSettlementLineForSync(t, nodeA, testconfig.Equivalent)
}

func TestSetSettlementLine10PercentPacketLossContractorNode(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineBadInternetTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Configure network conditions with 10% packet loss
	conditions := &vtcp.NetworkConditions{
		LossPercent: 10.0,
	}
	err := cluster.ConfigureNetworkConditions(nodeB, conditions, "eth0")
	if err != nil {
		t.Fatalf("failed to configure network conditions: %v", err)
	}

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "2000", vtcp.StatusOK)

	waitSetSettlementLineActive(t, nodeA, nodeB)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t) // Added based on python test
	nodeB.CheckValidKeys(t) // Added based on python test
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckSettlementLineForSync(t, nodeA, testconfig.Equivalent)
}

func TestSetSettlementLine10PercentPacketLossBothNodes(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineBadInternetTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Configure network conditions with 10% packet loss
	conditions := &vtcp.NetworkConditions{
		LossPercent: 10.0,
	}
	err := cluster.ConfigureNetworkConditions(nodeA, conditions, "eth0")
	if err != nil {
		t.Fatalf("failed to configure network conditions: %v", err)
	}
	err = cluster.ConfigureNetworkConditions(nodeB, conditions, "eth0")
	if err != nil {
		t.Fatalf("failed to configure network conditions: %v", err)
	}

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "2000", vtcp.StatusOK)

	waitSetSettlementLineActive(t, nodeA, nodeB)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t) // Added based on python test
	nodeB.CheckValidKeys(t) // Added based on python test
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckSettlementLineForSync(t, nodeA, testconfig.Equivalent)
}

func TestSetSettlementLine10PercentPacketCorruptionInitiatorNode(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineBadInternetTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Configure network conditions with 10% packet corruption
	conditions := &vtcp.NetworkConditions{
		CorruptPercent: 10.0,
	}
	err := cluster.ConfigureNetworkConditions(nodeA, conditions, "eth0")
	if err != nil {
		t.Fatalf("failed to configure network conditions: %v", err)
	}

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "2000", vtcp.StatusOK)

	waitSetSettlementLineActive(t, nodeA, nodeB)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t) // Added based on python test
	nodeB.CheckValidKeys(t) // Added based on python test
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckSettlementLineForSync(t, nodeA, testconfig.Equivalent)
}

func TestSetSettlementLine10PercentPacketCorruptionContractorNode(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineBadInternetTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Configure network conditions with 10% packet corruption
	conditions := &vtcp.NetworkConditions{
		CorruptPercent: 10.0,
	}
	err := cluster.ConfigureNetworkConditions(nodeB, conditions, "eth0")
	if err != nil {
		t.Fatalf("failed to configure network conditions: %v", err)
	}

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "2000", vtcp.StatusOK)

	waitSetSettlementLineActive(t, nodeA, nodeB)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t) // Added based on python test
	nodeB.CheckValidKeys(t) // Added based on python test
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckSettlementLineForSync(t, nodeA, testconfig.Equivalent)
}

func TestSetSettlementLine10PercentPacketCorruptionBothNodes(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineBadInternetTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Configure network conditions with 10% packet corruption
	conditions := &vtcp.NetworkConditions{
		CorruptPercent: 10.0,
	}
	err := cluster.ConfigureNetworkConditions(nodeA, conditions, "eth0")
	if err != nil {
		t.Fatalf("failed to configure network conditions: %v", err)
	}
	err = cluster.ConfigureNetworkConditions(nodeB, conditions, "eth0")
	if err != nil {
		t.Fatalf("failed to configure network conditions: %v", err)
	}

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "2000", vtcp.StatusOK)

	waitSetSettlementLineActive(t, nodeA, nodeB)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t) // Added based on python test
	nodeB.CheckValidKeys(t) // Added based on python test
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckSettlementLineForSync(t, nodeA, testconfig.Equivalent)
}

func TestSetSettlementLine10PercentPacketReorderingInitiatorNode(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineBadInternetTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Configure network conditions with 10% packet reordering
	conditions := &vtcp.NetworkConditions{
		ReorderPercent: 10.0,
		DelayMs:        10,
	}
	err := cluster.ConfigureNetworkConditions(nodeA, conditions, "eth0")
	if err != nil {
		t.Fatalf("failed to configure network conditions: %v", err)
	}

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "2000", vtcp.StatusOK)

	waitSetSettlementLineActive(t, nodeA, nodeB)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t) // Added based on python test
	nodeB.CheckValidKeys(t) // Added based on python test
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckSettlementLineForSync(t, nodeA, testconfig.Equivalent)
}

func TestSetSettlementLine10PercentPacketReorderingContractorNode(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineBadInternetTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Configure network conditions with 10% packet reordering
	conditions := &vtcp.NetworkConditions{
		ReorderPercent: 10.0,
		DelayMs:        10,
	}
	err := cluster.ConfigureNetworkConditions(nodeB, conditions, "eth0")
	if err != nil {
		t.Fatalf("failed to configure network conditions: %v", err)
	}

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "2000", vtcp.StatusOK)

	waitSetSettlementLineActive(t, nodeA, nodeB)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t) // Added based on python test
	nodeB.CheckValidKeys(t) // Added based on python test
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckSettlementLineForSync(t, nodeA, testconfig.Equivalent)
}

func TestSetSettlementLine10PercentPacketReorderingBothNodes(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineBadInternetTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Configure network conditions with 10% packet reordering
	conditions := &vtcp.NetworkConditions{
		ReorderPercent: 10.0,
		DelayMs:        10,
	}
	err := cluster.ConfigureNetworkConditions(nodeA, conditions, "eth0")
	if err != nil {
		t.Fatalf("failed to configure network conditions: %v", err)
	}
	err = cluster.ConfigureNetworkConditions(nodeB, conditions, "eth0")
	if err != nil {
		t.Fatalf("failed to configure network conditions: %v", err)
	}

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "2000", vtcp.StatusOK)

	waitSetSettlementLineActive(t, nodeA, nodeB)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t) // Added based on python test
	nodeB.CheckValidKeys(t) // Added based on python test
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckSettlementLineForSync(t, nodeA, testconfig.Equivalent)
}

func TestSetSettlementLine10PercentPacketDublicationInitiatorNode(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineBadInternetTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Configure network conditions with 10% packet duplication
	conditions := &vtcp.NetworkConditions{
		DuplicatePercent: 10.0,
	}
	err := cluster.ConfigureNetworkConditions(nodeA, conditions, "eth0")
	if err != nil {
		t.Fatalf("failed to configure network conditions: %v", err)
	}

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "2000", vtcp.StatusOK)

	waitSetSettlementLineActive(t, nodeA, nodeB)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t) // Added based on python test
	nodeB.CheckValidKeys(t) // Added based on python test
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckSettlementLineForSync(t, nodeA, testconfig.Equivalent)
}

func TestSetSettlementLine10PercentPacketDublicationContractorNode(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineBadInternetTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Configure network conditions with 10% packet duplication
	conditions := &vtcp.NetworkConditions{
		DuplicatePercent: 10.0,
	}
	err := cluster.ConfigureNetworkConditions(nodeB, conditions, "eth0")
	if err != nil {
		t.Fatalf("failed to configure network conditions: %v", err)
	}

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "2000", vtcp.StatusOK)

	waitSetSettlementLineActive(t, nodeA, nodeB)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t) // Added based on python test
	nodeB.CheckValidKeys(t) // Added based on python test
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckSettlementLineForSync(t, nodeA, testconfig.Equivalent)
}

func TestSetSettlementLine10PercentPacketDublicationBothNodes(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineBadInternetTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Configure network conditions with 10% packet duplication
	conditions := &vtcp.NetworkConditions{
		DuplicatePercent: 10.0,
	}
	err := cluster.ConfigureNetworkConditions(nodeA, conditions, "eth0")
	if err != nil {
		t.Fatalf("failed to configure network conditions: %v", err)
	}
	err = cluster.ConfigureNetworkConditions(nodeB, conditions, "eth0")
	if err != nil {
		t.Fatalf("failed to configure network conditions: %v", err)
	}

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "2000", vtcp.StatusOK)

	waitSetSettlementLineActive(t, nodeA, nodeB)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t) // Added based on python test
	nodeB.CheckValidKeys(t) // Added based on python test
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckSettlementLineForSync(t, nodeA, testconfig.Equivalent)
}

func TestSetSettlementLine200msDelayInitiatorNode(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineBadInternetTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Configure network conditions with 200ms delay
	conditions := &vtcp.NetworkConditions{
		DelayMs:  200,
		JitterMs: 50,
	}
	err := cluster.ConfigureNetworkConditions(nodeA, conditions, "eth0")
	if err != nil {
		t.Fatalf("failed to configure network conditions: %v", err)
	}

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "2000", vtcp.StatusOK)

	waitSetSettlementLineActive(t, nodeA, nodeB)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t) // Added based on python test
	nodeB.CheckValidKeys(t) // Added based on python test
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckSettlementLineForSync(t, nodeA, testconfig.Equivalent)
}

func TestSetSettlementLine200msDelayContractorNode(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineBadInternetTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Configure network conditions with 200ms delay
	conditions := &vtcp.NetworkConditions{
		DelayMs:  200,
		JitterMs: 50,
	}
	err := cluster.ConfigureNetworkConditions(nodeB, conditions, "eth0")
	if err != nil {
		t.Fatalf("failed to configure network conditions: %v", err)
	}

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "2000", vtcp.StatusOK)

	waitSetSettlementLineActive(t, nodeA, nodeB)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t) // Added based on python test
	nodeB.CheckValidKeys(t) // Added based on python test
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckSettlementLineForSync(t, nodeA, testconfig.Equivalent)
}

func TestSetSettlementLine200msDelayBothNodes(t *testing.T) {
	nodes, cluster := setupNodesForSetSettlementLineBadInternetTest(t, 2)
	nodeA, nodeB := nodes[0], nodes[1]

	// Configure network conditions with 200ms delay
	conditions := &vtcp.NetworkConditions{
		DelayMs:  200,
		JitterMs: 50,
	}
	err := cluster.ConfigureNetworkConditions(nodeA, conditions, "eth0")
	if err != nil {
		t.Fatalf("failed to configure network conditions: %v", err)
	}
	err = cluster.ConfigureNetworkConditions(nodeB, conditions, "eth0")
	if err != nil {
		t.Fatalf("failed to configure network conditions: %v", err)
	}

	nodeA.SetSettlementLine(t, nodeB, testconfig.Equivalent, "2000", vtcp.StatusOK)

	waitSetSettlementLineActive(t, nodeA, nodeB)

	nodeA.CheckSerializedTransaction(t, false, 0)
	nodeB.CheckSerializedTransaction(t, false, 0)
	nodeA.CheckSettlementLineState(t, nodeB, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeB.CheckSettlementLineState(t, nodeA, testconfig.Equivalent, vtcp.SettlementLineStateActive)
	nodeA.CheckValidKeys(t) // Added based on python test
	nodeB.CheckValidKeys(t) // Added based on python test
	nodeA.CheckSettlementLineForSync(t, nodeB, testconfig.Equivalent)
	nodeB.CheckSettlementLineForSync(t, nodeA, testconfig.Equivalent)
}
