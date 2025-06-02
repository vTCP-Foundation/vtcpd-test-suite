package main

import (
	"context"
	"testing"

	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
	"github.com/vTCP-Foundation/vtcpd-test-suite/tests/testconfig"
)

func TestMaxFlowThrough6Hops(t *testing.T) {
	node1 := vtcp.NewNode(t, testconfig.StaticContainerIPPartForMaxFlowTest+"1", "node1")
	node2 := vtcp.NewNode(t, testconfig.StaticContainerIPPartForMaxFlowTest+"2", "node2")
	node3 := vtcp.NewNode(t, testconfig.StaticContainerIPPartForMaxFlowTest+"3", "node31")
	node4 := vtcp.NewNode(t, testconfig.StaticContainerIPPartForMaxFlowTest+"4", "node4")
	node5 := vtcp.NewNode(t, testconfig.StaticContainerIPPartForMaxFlowTest+"5", "node5")
	node6 := vtcp.NewNode(t, testconfig.StaticContainerIPPartForMaxFlowTest+"6", "node6")
	node7 := vtcp.NewNode(t, testconfig.StaticContainerIPPartForMaxFlowTest+"7", "node72")

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

	node2.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "1000")
	node3.CreateAndSetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "800")
	node4.CreateAndSetSettlementLineAndCheck(t, node3, testconfig.Equivalent, "900")
	node5.CreateAndSetSettlementLineAndCheck(t, node4, testconfig.Equivalent, "700")
	node6.CreateAndSetSettlementLineAndCheck(t, node5, testconfig.Equivalent, "900")
	node7.CreateAndSetSettlementLineAndCheck(t, node6, testconfig.Equivalent, "1000")

	node1.CheckMaxFlow(t, node7, testconfig.Equivalent, "700")

	node1.CreateTransactionCheckStatus(t, node7, testconfig.Equivalent, "200", vtcp.StatusOK)

	node1.CheckMaxFlow(t, node7, testconfig.Equivalent, "500")

}
