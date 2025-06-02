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
	severalPaymentsAtTheSameTimeNextNodeIndex = 1
)

func getNextIPForSeveralPaymentsAtTheSameTimeTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForSeveralPaymentsAtTheSameTimeTest, severalPaymentsAtTheSameTimeNextNodeIndex)
	severalPaymentsAtTheSameTimeNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForSeveralPaymentsAtTheSameTimeTest(t *testing.T, count int) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, count)
	for i := range count {
		nodes[i] = vtcp.NewNode(t, getNextIPForSeveralPaymentsAtTheSameTimeTest(), fmt.Sprintf("node%c", i+1))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings) // Assuming clusterSettings is defined globally or passed
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes)
	return nodes, cluster
}

func Test1WithCoordinatorSettlementLineCommon(t *testing.T) {
	nodes, _ := setupNodesForSeveralPaymentsAtTheSameTimeTest(t, 9)

	nodes[1].OpenChannelAndCheck(t, nodes[0])
	nodes[4].OpenChannelAndCheck(t, nodes[3])
	nodes[7].OpenChannelAndCheck(t, nodes[6])
	nodes[2].OpenChannelAndCheck(t, nodes[1])
	nodes[3].OpenChannelAndCheck(t, nodes[1])
	nodes[5].OpenChannelAndCheck(t, nodes[1])
	nodes[6].OpenChannelAndCheck(t, nodes[1])
	nodes[8].OpenChannelAndCheck(t, nodes[1])

	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "1000")
	nodes[2].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "200")
	nodes[3].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "300")
	nodes[4].CreateAndSetSettlementLineAndCheck(t, nodes[3], testconfig.Equivalent, "100")
	nodes[5].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "300")
	nodes[6].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "700")
	nodes[7].CreateAndSetSettlementLineAndCheck(t, nodes[6], testconfig.Equivalent, "200")
	nodes[8].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.Equivalent, "100")

	expectedMaxFlows := []vtcp.MaxFlowBatchCheck{
		{Node: nodes[2], ExpectedMaxFlow: "200"},
		{Node: nodes[4], ExpectedMaxFlow: "100"},
		{Node: nodes[5], ExpectedMaxFlow: "300"},
		{Node: nodes[7], ExpectedMaxFlow: "200"},
		{Node: nodes[8], ExpectedMaxFlow: "100"},
	}
	nodes[0].CheckMaxFlowBatch(t, expectedMaxFlows, testconfig.Equivalent)

	nodes[0].CreateTransactionCheckStatus(t, nodes[2], testconfig.Equivalent, "200", vtcp.StatusOK)
	nodes[0].CreateTransactionCheckStatus(t, nodes[4], testconfig.Equivalent, "100", vtcp.StatusOK)
	nodes[0].CreateTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "300", vtcp.StatusOK)
	nodes[0].CreateTransactionCheckStatus(t, nodes[7], testconfig.Equivalent, "200", vtcp.StatusOK)
	nodes[0].CreateTransactionCheckStatus(t, nodes[8], testconfig.Equivalent, "100", vtcp.StatusOK)
	time.Sleep(5 * time.Second)

	for _, node := range nodes {
		node.CheckSerializedTransaction(t, false, 0)
	}

	expectedMaxFlows = []vtcp.MaxFlowBatchCheck{
		{Node: nodes[2], ExpectedMaxFlow: "0"},
		{Node: nodes[4], ExpectedMaxFlow: "0"},
		{Node: nodes[5], ExpectedMaxFlow: "0"},
		{Node: nodes[7], ExpectedMaxFlow: "0"},
		{Node: nodes[8], ExpectedMaxFlow: "0"},
	}
	nodes[0].CheckMaxFlowBatch(t, expectedMaxFlows, testconfig.Equivalent)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test2WithReceiverSettlementLineCommon(t *testing.T) {
	nodes, _ := setupNodesForSeveralPaymentsAtTheSameTimeTest(t, 9)

	nodes[8].OpenChannelAndCheck(t, nodes[7])
	nodes[5].OpenChannelAndCheck(t, nodes[4])
	nodes[2].OpenChannelAndCheck(t, nodes[1])
	nodes[7].OpenChannelAndCheck(t, nodes[5])
	nodes[7].OpenChannelAndCheck(t, nodes[5])
	nodes[7].OpenChannelAndCheck(t, nodes[3])
	nodes[7].OpenChannelAndCheck(t, nodes[2])
	nodes[7].OpenChannelAndCheck(t, nodes[0])

	nodes[8].SetSettlementLineAndCheck(t, nodes[7], testconfig.Equivalent, "1000")
	nodes[7].SetSettlementLineAndCheck(t, nodes[6], testconfig.Equivalent, "200")
	nodes[7].SetSettlementLineAndCheck(t, nodes[5], testconfig.Equivalent, "300")
	nodes[5].SetSettlementLineAndCheck(t, nodes[4], testconfig.Equivalent, "100")
	nodes[7].SetSettlementLineAndCheck(t, nodes[3], testconfig.Equivalent, "700")
	nodes[3].SetSettlementLineAndCheck(t, nodes[2], testconfig.Equivalent, "200")
	nodes[7].SetSettlementLineAndCheck(t, nodes[0], testconfig.Equivalent, "100")

	nodes[0].CheckMaxFlow(t, nodes[8], testconfig.Equivalent, "100")
	nodes[1].CheckMaxFlow(t, nodes[8], testconfig.Equivalent, "200")
	nodes[3].CheckMaxFlow(t, nodes[8], testconfig.Equivalent, "300")
	nodes[4].CheckMaxFlow(t, nodes[8], testconfig.Equivalent, "100")
	nodes[6].CheckMaxFlow(t, nodes[8], testconfig.Equivalent, "200")

	nodes[0].CreateTransactionCheckStatus(t, nodes[8], testconfig.Equivalent, "100", vtcp.StatusOK)
	nodes[1].CreateTransactionCheckStatus(t, nodes[8], testconfig.Equivalent, "200", vtcp.StatusOK)
	nodes[3].CreateTransactionCheckStatus(t, nodes[8], testconfig.Equivalent, "300", vtcp.StatusOK)
	nodes[4].CreateTransactionCheckStatus(t, nodes[8], testconfig.Equivalent, "100", vtcp.StatusOK)
	nodes[6].CreateTransactionCheckStatus(t, nodes[8], testconfig.Equivalent, "200", vtcp.StatusOK)

	time.Sleep(20 * time.Second)

	for _, node := range nodes {
		node.CheckSerializedTransaction(t, false, 0)
	}

	nodes[0].CreateTransactionCheckStatus(t, nodes[8], testconfig.Equivalent, "0", vtcp.StatusOK)
	nodes[1].CreateTransactionCheckStatus(t, nodes[8], testconfig.Equivalent, "0", vtcp.StatusOK)
	nodes[3].CreateTransactionCheckStatus(t, nodes[8], testconfig.Equivalent, "0", vtcp.StatusOK)
	nodes[4].CreateTransactionCheckStatus(t, nodes[8], testconfig.Equivalent, "0", vtcp.StatusOK)
	nodes[6].CreateTransactionCheckStatus(t, nodes[8], testconfig.Equivalent, "0", vtcp.StatusOK)

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test3WithIntermediateNodeCommon(t *testing.T) {
	nodes, _ := setupNodesForSeveralPaymentsAtTheSameTimeTest(t, 14)

	// Assign specific node variables for clarity
	node1 := nodes[0]
	node2 := nodes[1]
	node3 := nodes[2]
	node4 := nodes[3]
	node5 := nodes[4]
	node6 := nodes[5]
	node7 := nodes[6]
	node8 := nodes[7]
	node9 := nodes[8]
	node10 := nodes[9]
	node11 := nodes[10]
	node12 := nodes[11]
	node13 := nodes[12]
	node14 := nodes[13]

	node1.OpenChannelAndCheck(t, node4)
	node2.OpenChannelAndCheck(t, node3)
	node1.OpenChannelAndCheck(t, node6)
	node8.OpenChannelAndCheck(t, node7)
	node1.OpenChannelAndCheck(t, node9)
	node12.OpenChannelAndCheck(t, node13)
	node1.OpenChannelAndCheck(t, node11)
	node1.OpenChannelAndCheck(t, node14)
	node1.OpenChannelAndCheck(t, node3)
	node1.OpenChannelAndCheck(t, node5)
	node1.OpenChannelAndCheck(t, node8)
	node1.OpenChannelAndCheck(t, node10)
	node1.OpenChannelAndCheck(t, node13)

	node4.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "10000")
	node6.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "10000")
	node9.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "10000")
	node11.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "10000")
	node14.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "10000")
	node3.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "10000")
	node5.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "10000")
	node8.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "10000")
	node10.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "10000")
	node13.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "10000")
	node3.CreateAndSetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "500")
	node8.CreateAndSetSettlementLineAndCheck(t, node7, testconfig.Equivalent, "500")
	node13.CreateAndSetSettlementLineAndCheck(t, node12, testconfig.Equivalent, "500")

	node1.CreateTransactionCheckStatus(t, node3, testconfig.Equivalent, "100", vtcp.StatusOK)
	node1.CreateTransactionCheckStatus(t, node5, testconfig.Equivalent, "100", vtcp.StatusOK)
	node1.CreateTransactionCheckStatus(t, node8, testconfig.Equivalent, "100", vtcp.StatusOK)
	node1.CreateTransactionCheckStatus(t, node10, testconfig.Equivalent, "100", vtcp.StatusOK)
	node1.CreateTransactionCheckStatus(t, node13, testconfig.Equivalent, "100", vtcp.StatusOK)
	time.Sleep(20 * time.Second)

	node2.CheckMaxFlow(t, node4, testconfig.Equivalent, "100")
	node5.CheckMaxFlow(t, node6, testconfig.Equivalent, "100")
	node7.CheckMaxFlow(t, node9, testconfig.Equivalent, "100")
	node10.CheckMaxFlow(t, node11, testconfig.Equivalent, "100")
	node12.CheckMaxFlow(t, node14, testconfig.Equivalent, "100")

	node2.CreateTransactionCheckStatus(t, node4, testconfig.Equivalent, "50", vtcp.StatusOK)
	node5.CreateTransactionCheckStatus(t, node6, testconfig.Equivalent, "50", vtcp.StatusOK)
	node7.CreateTransactionCheckStatus(t, node9, testconfig.Equivalent, "50", vtcp.StatusOK)
	node10.CreateTransactionCheckStatus(t, node11, testconfig.Equivalent, "50", vtcp.StatusOK)
	node12.CreateTransactionCheckStatus(t, node14, testconfig.Equivalent, "50", vtcp.StatusOK)
	time.Sleep(20 * time.Second)

	for _, node := range nodes {
		node.CheckSerializedTransaction(t, false, 0)
	}

	node2.CheckMaxFlow(t, node4, testconfig.Equivalent, "50")
	node5.CheckMaxFlow(t, node6, testconfig.Equivalent, "50")
	node7.CheckMaxFlow(t, node9, testconfig.Equivalent, "50")
	node10.CheckMaxFlow(t, node11, testconfig.Equivalent, "50")
	node12.CheckMaxFlow(t, node14, testconfig.Equivalent, "50")

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test4WithTwoCommonNodes(t *testing.T) {
	nodes, _ := setupNodesForSeveralPaymentsAtTheSameTimeTest(t, 2)

	node1 := nodes[0]
	node2 := nodes[1]

	trustSize := "1000"
	cntPayments := 20
	paymentSize := "50"

	node2.OpenChannelAndCheck(t, node1)
	node2.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, trustSize)

	for i := 0; i < cntPayments; i++ {
		node1.CreateTransactionCheckStatus(t, node2, testconfig.Equivalent, paymentSize, vtcp.StatusOK)
		time.Sleep(50 * time.Millisecond)
	}
	time.Sleep(40 * time.Second)

	node1.CheckSerializedTransaction(t, false, 0)
	node2.CheckSerializedTransaction(t, false, 0)

	// trust_size - cnt_payments * payment_size = 1000 - 20 * 50 = 0
	node1.CheckMaxFlow(t, node2, testconfig.Equivalent, "0")
	// cnt_payments * payment_size = 20 * 50 = 1000
	node2.CheckMaxFlow(t, node1, testconfig.Equivalent, "1000")

	vtcp.CheckSettlementLineForSyncBatch(t, nodes, testconfig.Equivalent, 3)
}

func Test5ThreeNodesWithGatewayKeysExhausting(t *testing.T) {
	nodes, _ := setupNodesForSeveralPaymentsAtTheSameTimeTest(t, 3)

	node1 := nodes[0]
	node2 := nodes[1]
	node3 := nodes[2]

	node1.OpenChannelAndCheck(t, node2)
	node2.OpenChannelAndCheck(t, node3)

	node2.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "10000")
	node3.CreateAndSetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "10000")

	node3.SetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "10001")
	node3.SetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "10002")
	node3.SetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "10003")
	node3.SetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "10004")
	node3.SetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "10005")
	node3.SetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "10006")
	node3.SetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "10007")
	node3.SetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "10000")

	cntPayments := 5

	for i := 0; i < cntPayments; i++ {
		node2.CreateTransactionCheckStatus(t, node3, testconfig.Equivalent, "100", vtcp.StatusOK)
		time.Sleep(500 * time.Millisecond)
	}

	node1.CreateTransactionCheckStatus(t, node3, testconfig.Equivalent, "200", vtcp.StatusOK)
	node1.CreateTransactionCheckStatus(t, node3, testconfig.Equivalent, "300", vtcp.StatusOK)
	node1.CreateTransactionCheckStatus(t, node3, testconfig.Equivalent, "400", vtcp.StatusOK)
	node1.CreateTransactionCheckStatus(t, node3, testconfig.Equivalent, "500", vtcp.StatusOK)
	node1.CreateTransactionCheckStatus(t, node3, testconfig.Equivalent, "600", vtcp.StatusOK)

	time.Sleep(5 * time.Second)
	node1.CheckMaxFlow(t, node3, testconfig.Equivalent, "7500")

	node1.CreateTransactionCheckStatus(t, node3, testconfig.Equivalent, "50", vtcp.StatusOK)
}

func Test6ThreeNodes(t *testing.T) {
	nodes, _ := setupNodesForSeveralPaymentsAtTheSameTimeTest(t, 3)

	node1 := nodes[0]
	node2 := nodes[1]
	node3 := nodes[2]

	tlAmount := "100000"

	node1.OpenChannelAndCheck(t, node2)
	node2.OpenChannelAndCheck(t, node3)
	node2.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, tlAmount)
	node3.CreateAndSetSettlementLineAndCheck(t, node2, testconfig.Equivalent, tlAmount)

	cntPayments := 10
	paymentAmount := 10
	totalAmount := 0

	for i := 0; i < cntPayments; i++ {
		node1.CreateTransactionCheckStatus(t, node3, testconfig.Equivalent, fmt.Sprintf("%d", paymentAmount), vtcp.StatusOK)
		totalAmount += paymentAmount
		paymentAmount++
		time.Sleep(400 * time.Millisecond)
	}

	time.Sleep(5 * time.Second)
	node1.CheckSettlementLineForSync(t, node2, testconfig.Equivalent)
	node2.CheckSettlementLineForSync(t, node3, testconfig.Equivalent)

	// tl_amount - total_amount = 100000 - 55 = 99945
	node1.CheckMaxFlow(t, node3, testconfig.Equivalent, "99945")

	node1.CreateTransactionCheckStatus(t, node3, testconfig.Equivalent, "50", vtcp.StatusOK)
}

func Test7OneHubManyToMany(t *testing.T) {
	coordinatorCount := 10
	receiverCount := 10
	totalNodes := 1 + coordinatorCount + receiverCount // 1 hub + coordinators + receivers
	nodes, _ := setupNodesForSeveralPaymentsAtTheSameTimeTest(t, totalNodes)

	hub := nodes[0]
	coordinators := make([]*vtcp.Node, coordinatorCount)
	receivers := make([]*vtcp.Node, receiverCount)

	// Assign coordinators
	for i := 0; i < coordinatorCount; i++ {
		coordinators[i] = nodes[1+i]
	}

	// Assign receivers
	for i := 0; i < receiverCount; i++ {
		receivers[i] = nodes[1+coordinatorCount+i]
	}

	nodeTlAmount := "100000"

	// Setup coordinators
	for i := 0; i < coordinatorCount; i++ {
		coordinators[i].OpenChannelAndCheck(t, hub)
		coordinators[i].CreateAndSetSettlementLineAndCheck(t, hub, testconfig.Equivalent, nodeTlAmount)
		hub.SetSettlementLineAndCheck(t, coordinators[i], testconfig.Equivalent, nodeTlAmount)
	}

	// Setup receivers
	for i := 0; i < receiverCount; i++ {
		receivers[i].OpenChannelAndCheck(t, hub)
		receivers[i].CreateAndSetSettlementLineAndCheck(t, hub, testconfig.Equivalent, nodeTlAmount)
	}

	cntPayments := 100
	paymentAmount := "10"

	idx := 0
	for x := 0; x < cntPayments; x++ {
		coordinators[idx].CreateTransactionCheckStatus(t, receivers[idx], testconfig.Equivalent, paymentAmount, vtcp.StatusOK)
		time.Sleep(400 * time.Millisecond)
		idx++
		if idx >= coordinatorCount {
			idx = 0
		}
	}

	time.Sleep(5 * time.Second)

	// Check max flows - each coordinator should have sent payment_amount * (cnt_payments / cnt_nodes) = 10 * 10 = 100
	// So remaining should be 100000 - 100 = 99900
	for i := 0; i < coordinatorCount; i++ {
		coordinators[i].CheckMaxFlow(t, receivers[i], testconfig.Equivalent, "99900")
	}
}

func Test8TwoHubsSenderReceiverDifferentHubs(t *testing.T) {
	nodes, _ := setupNodesForSeveralPaymentsAtTheSameTimeTest(t, 8) // 2 hubs + 2 nodes + 4 additional nodes

	node1 := nodes[0]
	node2 := nodes[1]
	hub1 := nodes[2]
	hub2 := nodes[3]
	// Additional nodes for each hub (nodes[4-7])

	tlAmount := "100000"

	hub1.OpenChannelAndCheck(t, hub2)
	hub1.CreateAndSetSettlementLineAndCheck(t, hub2, testconfig.Equivalent, tlAmount)
	hub2.SetSettlementLineAndCheck(t, hub1, testconfig.Equivalent, tlAmount)

	// Setup additional nodes for hub1
	for i := 4; i < 6; i++ {
		nodes[i].OpenChannelAndCheck(t, hub1)
		nodes[i].CreateAndSetSettlementLineAndCheck(t, hub1, testconfig.Equivalent, tlAmount)
	}

	// Setup additional nodes for hub2
	for i := 6; i < 8; i++ {
		nodes[i].OpenChannelAndCheck(t, hub2)
		nodes[i].CreateAndSetSettlementLineAndCheck(t, hub2, testconfig.Equivalent, tlAmount)
	}

	node1.OpenChannelAndCheck(t, hub1)
	node2.OpenChannelAndCheck(t, hub2)
	node1.CreateAndSetSettlementLineAndCheck(t, hub1, testconfig.Equivalent, tlAmount)
	node2.CreateAndSetSettlementLineAndCheck(t, hub2, testconfig.Equivalent, tlAmount)

	hub1.SetSettlementLineAndCheck(t, node1, testconfig.Equivalent, tlAmount)

	cntPayments := 30
	paymentAmount := 10
	totalAmount := 0

	for x := 0; x < cntPayments; x++ {
		node1.CreateTransactionCheckStatus(t, node2, testconfig.Equivalent, fmt.Sprintf("%d", paymentAmount), vtcp.StatusOK)
		totalAmount += paymentAmount
		paymentAmount++
		time.Sleep(400 * time.Millisecond)
	}

	time.Sleep(5 * time.Second)

	// Check trustlines
	hub1.CheckSettlementLineForSync(t, hub2, testconfig.Equivalent)
	hub1.CheckSettlementLineForSync(t, node1, testconfig.Equivalent)
	hub2.CheckSettlementLineForSync(t, node2, testconfig.Equivalent)

	// total_amount = sum from 10 to 39 = 735
	node1.CheckMaxFlow(t, node2, testconfig.Equivalent, "99265")
}

func Test9TwoHubsSenderReceiverTheSameHub(t *testing.T) {
	nodes, _ := setupNodesForSeveralPaymentsAtTheSameTimeTest(t, 8) // 2 hubs + 2 nodes + 4 additional nodes

	node1 := nodes[0]
	node2 := nodes[1]
	hub1 := nodes[2]
	hub2 := nodes[3]

	tlAmount := "100000"

	hub1.OpenChannelAndCheck(t, hub2)
	time.Sleep(1 * time.Second)
	hub1.CreateAndSetSettlementLineAndCheck(t, hub2, testconfig.Equivalent, tlAmount)
	hub2.SetSettlementLineAndCheck(t, hub1, testconfig.Equivalent, tlAmount)

	// Setup additional nodes for hub1
	for i := 4; i < 6; i++ {
		nodes[i].OpenChannelAndCheck(t, hub1)
		nodes[i].CreateAndSetSettlementLineAndCheck(t, hub1, testconfig.Equivalent, tlAmount)
	}

	// Setup additional nodes for hub2
	for i := 6; i < 8; i++ {
		nodes[i].OpenChannelAndCheck(t, hub2)
		nodes[i].CreateAndSetSettlementLineAndCheck(t, hub2, testconfig.Equivalent, tlAmount)
	}

	node1.OpenChannelAndCheck(t, hub1)
	node2.OpenChannelAndCheck(t, hub1) // Both nodes connect to hub
	node1.CreateAndSetSettlementLineAndCheck(t, hub1, testconfig.Equivalent, tlAmount)
	node2.CreateAndSetSettlementLineAndCheck(t, hub1, testconfig.Equivalent, tlAmount)

	hub1.SetSettlementLineAndCheck(t, node1, testconfig.Equivalent, tlAmount)

	cntPayments := 10
	paymentAmount := 10
	totalAmount := 0

	for x := 0; x < cntPayments; x++ {
		node1.CreateTransactionCheckStatus(t, node2, testconfig.Equivalent, fmt.Sprintf("%d", paymentAmount), vtcp.StatusOK)
		totalAmount += paymentAmount
		paymentAmount++
		time.Sleep(400 * time.Millisecond)
	}

	time.Sleep(5 * time.Second)

	// Check trustlines
	hub1.CheckSettlementLineForSync(t, hub2, testconfig.Equivalent)
	hub1.CheckSettlementLineForSync(t, node1, testconfig.Equivalent)
	hub1.CheckSettlementLineForSync(t, node2, testconfig.Equivalent)

	// total_amount = sum from 10 to 19 = 145
	node1.CheckMaxFlow(t, node2, testconfig.Equivalent, "99855")
}

func Test10TwoHubsManyToManyThroughSeveralHubs(t *testing.T) {
	coordinatorCount := 10
	receiverCount := 10
	totalNodes := 2 + coordinatorCount + receiverCount // 2 hubs + coordinators + receivers
	nodes, _ := setupNodesForSeveralPaymentsAtTheSameTimeTest(t, totalNodes)

	hub1 := nodes[0]
	hub2 := nodes[1]
	coordinators := make([]*vtcp.Node, coordinatorCount)
	receivers := make([]*vtcp.Node, receiverCount)

	// Assign coordinators
	for i := 0; i < coordinatorCount; i++ {
		coordinators[i] = nodes[2+i]
	}

	// Assign receivers
	for i := 0; i < receiverCount; i++ {
		receivers[i] = nodes[2+coordinatorCount+i]
	}

	hubTlAmount := "10000000"
	nodeTlAmount := "100000"

	hub1.OpenChannelAndCheck(t, hub2)
	hub1.CreateAndSetSettlementLineAndCheck(t, hub2, testconfig.Equivalent, hubTlAmount)
	hub2.SetSettlementLineAndCheck(t, hub1, testconfig.Equivalent, hubTlAmount)

	// Setup coordinators with hub1
	for i := 0; i < coordinatorCount; i++ {
		coordinators[i].OpenChannelAndCheck(t, hub1)
		coordinators[i].CreateAndSetSettlementLineAndCheck(t, hub1, testconfig.Equivalent, nodeTlAmount)
		hub1.SetSettlementLineAndCheck(t, coordinators[i], testconfig.Equivalent, nodeTlAmount)
	}

	// Setup receivers with hub2
	for i := 0; i < receiverCount; i++ {
		receivers[i].OpenChannelAndCheck(t, hub2)
		receivers[i].CreateAndSetSettlementLineAndCheck(t, hub2, testconfig.Equivalent, nodeTlAmount)
	}

	cntPayments := 100
	paymentAmount := "10"

	idx := 0
	for x := 0; x < cntPayments; x++ {
		coordinators[idx].CreateTransactionCheckStatus(t, receivers[idx], testconfig.Equivalent, paymentAmount, vtcp.StatusOK)
		time.Sleep(400 * time.Millisecond)
		idx++
		if idx >= coordinatorCount {
			idx = 0
		}
	}

	time.Sleep(5 * time.Second)

	// Check trustlines
	hub1.CheckSettlementLineForSync(t, hub2, testconfig.Equivalent)
	for i := 0; i < coordinatorCount; i++ {
		hub1.CheckSettlementLineForSync(t, coordinators[i], testconfig.Equivalent)
	}
	for i := 0; i < receiverCount; i++ {
		hub2.CheckSettlementLineForSync(t, receivers[i], testconfig.Equivalent)
	}

	// Each coordinator sent payment_amount * (cnt_payments / cnt_nodes) = 10 * 10 = 100
	// So remaining should be 100000 - 100 = 99900
	for i := 0; i < coordinatorCount; i++ {
		coordinators[i].CheckMaxFlow(t, receivers[i], testconfig.Equivalent, "99900")
	}
}

func Test11TwoHubsManyToManyThroughSeveralHubsHubsTlLimit(t *testing.T) {
	coordinatorCount := 10
	receiverCount := 10
	totalNodes := 2 + coordinatorCount + receiverCount // 2 hubs + coordinators + receivers
	nodes, _ := setupNodesForSeveralPaymentsAtTheSameTimeTest(t, totalNodes)

	hub1 := nodes[0]
	hub2 := nodes[1]
	coordinators := make([]*vtcp.Node, coordinatorCount)
	receivers := make([]*vtcp.Node, receiverCount)

	// Assign coordinators
	for i := 0; i < coordinatorCount; i++ {
		coordinators[i] = nodes[2+i]
	}

	// Assign receivers
	for i := 0; i < receiverCount; i++ {
		receivers[i] = nodes[2+coordinatorCount+i]
	}

	hubTlAmount := "10000"
	nodeTlAmount := "100000"

	hub1.OpenChannelAndCheck(t, hub2)
	hub1.CreateAndSetSettlementLineAndCheck(t, hub2, testconfig.Equivalent, hubTlAmount)
	hub2.SetSettlementLineAndCheck(t, hub1, testconfig.Equivalent, hubTlAmount)

	// Setup coordinators with hub1
	for i := 0; i < coordinatorCount; i++ {
		coordinators[i].OpenChannelAndCheck(t, hub1)
		coordinators[i].CreateAndSetSettlementLineAndCheck(t, hub1, testconfig.Equivalent, nodeTlAmount)
		hub1.SetSettlementLineAndCheck(t, coordinators[i], testconfig.Equivalent, nodeTlAmount)
	}

	// Setup receivers with hub2
	for i := 0; i < receiverCount; i++ {
		receivers[i].OpenChannelAndCheck(t, hub2)
		receivers[i].CreateAndSetSettlementLineAndCheck(t, hub2, testconfig.Equivalent, nodeTlAmount)
	}

	cntPayments := 30
	paymentAmount := "10"

	idx := 0
	for x := 0; x < cntPayments; x++ {
		coordinators[idx].CreateTransactionCheckStatus(t, receivers[idx], testconfig.Equivalent, paymentAmount, vtcp.StatusOK)
		time.Sleep(400 * time.Millisecond)
		idx++
		if idx >= coordinatorCount {
			idx = 0
		}
	}

	time.Sleep(5 * time.Second)

	// Check trustlines
	hub1.CheckSettlementLineForSync(t, hub2, testconfig.Equivalent)
	for i := 0; i < coordinatorCount; i++ {
		hub1.CheckSettlementLineForSync(t, coordinators[i], testconfig.Equivalent)
	}
	for i := 0; i < receiverCount; i++ {
		hub2.CheckSettlementLineForSync(t, receivers[i], testconfig.Equivalent)
	}

	// hub_tl_amount - payment_amount * cnt_payments = 10000 - 10 * 30 = 9700
	for i := 0; i < coordinatorCount; i++ {
		coordinators[i].CheckMaxFlow(t, receivers[i], testconfig.Equivalent, "9700")
	}
}
