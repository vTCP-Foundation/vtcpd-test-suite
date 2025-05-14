package tests

import (
	"context"
	"testing"

	"github.com/vTCP-Foundation/vtcpd-test-suite/internal/conf"
	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
)

var (
	// Get settings from the central configuration package
	testCfg         = conf.GetConfig()
	clusterSettings = &vtcp.ClusterSettings{
		NodeImageName: testCfg.NodeImageName,
		NetworkName:   testCfg.NetworkName,
	}
)

func TestMaxFlowThrough6Hops(t *testing.T) {
	node1 := vtcp.NewNode(t, "172.18.0.2", "node1")
	node2 := vtcp.NewNode(t, "172.18.0.3", "node2")
	node3 := vtcp.NewNode(t, "172.18.0.4", "node31")
	node4 := vtcp.NewNode(t, "172.18.0.5", "node4")
	node5 := vtcp.NewNode(t, "172.18.0.6", "node5")
	node6 := vtcp.NewNode(t, "172.18.0.7", "node6")
	node7 := vtcp.NewNode(t, "172.18.0.8", "node72")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, clusterSettings)
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

	equivalent := "1"

	node2.CreateSettlementLineAndCheck(t, node1, equivalent, "1000")
	node3.CreateSettlementLineAndCheck(t, node2, equivalent, "800")
	node4.CreateSettlementLineAndCheck(t, node3, equivalent, "900")
	node5.CreateSettlementLineAndCheck(t, node4, equivalent, "700")
	node6.CreateSettlementLineAndCheck(t, node5, equivalent, "900")
	node7.CreateSettlementLineAndCheck(t, node6, equivalent, "1000")

	node1.CheckMaxFlow(t, node7, equivalent, "700")

	node1.CreateTransactionCheckStatus(t, node7, equivalent, "200", vtcp.StatusOK)

	node1.CheckMaxFlow(t, node7, equivalent, "500")

}
