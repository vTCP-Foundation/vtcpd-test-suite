package main

import (
	"context"
	"testing"
	"time"

	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
	"github.com/vTCP-Foundation/vtcpd-test-suite/tests/testconfig"
)

func TestMaxBatchFlowThrough6Hops(t *testing.T) {
	node1 := vtcp.NewNode(t, testconfig.StaticContainerIPPartForMaxFlowBatchTest+"1", "node1")
	node2 := vtcp.NewNode(t, testconfig.StaticContainerIPPartForMaxFlowBatchTest+"2", "node2")
	node3 := vtcp.NewNode(t, testconfig.StaticContainerIPPartForMaxFlowBatchTest+"3", "node3")
	node4 := vtcp.NewNode(t, testconfig.StaticContainerIPPartForMaxFlowBatchTest+"4", "node4")
	node5 := vtcp.NewNode(t, testconfig.StaticContainerIPPartForMaxFlowBatchTest+"5", "node5")
	node6 := vtcp.NewNode(t, testconfig.StaticContainerIPPartForMaxFlowBatchTest+"6", "node6")
	node7 := vtcp.NewNode(t, testconfig.StaticContainerIPPartForMaxFlowBatchTest+"7", "node7")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{node1, node2, node3, node4, node5, node6, node7})

	node1.OpenChannelAndCheck(t, node2)
	node3.OpenChannelAndCheck(t, node4)
	node5.OpenChannelAndCheck(t, node6)

	node2.OpenChannelAndCheck(t, node3)
	node4.OpenChannelAndCheck(t, node5)
	node6.OpenChannelAndCheck(t, node7)

	node2.CreateSettlementLineAndCheck(t, node1, testconfig.Equivalent, "1000")
	node3.CreateSettlementLineAndCheck(t, node2, testconfig.Equivalent, "800")
	node4.CreateSettlementLineAndCheck(t, node3, testconfig.Equivalent, "900")
	node5.CreateSettlementLineAndCheck(t, node4, testconfig.Equivalent, "700")
	node6.CreateSettlementLineAndCheck(t, node5, testconfig.Equivalent, "900")
	node7.CreateSettlementLineAndCheck(t, node6, testconfig.Equivalent, "1000")

	expectedMaxFlows := []vtcp.MaxFlowBatchCheck{
		{Node: node2, ExpectedMaxFlow: "1000"},
		{Node: node3, ExpectedMaxFlow: "800"},
		{Node: node4, ExpectedMaxFlow: "800"},
		{Node: node5, ExpectedMaxFlow: "700"},
		{Node: node6, ExpectedMaxFlow: "700"},
		{Node: node7, ExpectedMaxFlow: "700"},
	}
	node1.CheckMaxFlowBatch(t, expectedMaxFlows, testconfig.Equivalent)

	node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "200", vtcp.StatusOK)
	time.Sleep(5 * time.Second)

	expectedMaxFlows = []vtcp.MaxFlowBatchCheck{
		{Node: node2, ExpectedMaxFlow: "800"},
		{Node: node3, ExpectedMaxFlow: "600"},
		{Node: node4, ExpectedMaxFlow: "600"},
		{Node: node5, ExpectedMaxFlow: "500"},
		{Node: node6, ExpectedMaxFlow: "500"},
		{Node: node7, ExpectedMaxFlow: "500"},
	}
	node1.CheckMaxFlowBatch(t, expectedMaxFlows, testconfig.Equivalent)

}
