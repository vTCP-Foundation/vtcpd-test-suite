package tests

import (
	"context"
	"strconv"
	"testing"
	"time"

	// Assuming your go.mod module path is 'github.com/vTCP-Foundation/vtcpd-test-suite'
	// Adjust this path if your module name is different.
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

func TestTrustLineSet(t *testing.T) {
	nodeA := vtcp.NewNode(t, "172.18.0.2", "nodeA")
	nodeB := vtcp.NewNode(t, "172.18.0.3", "nodeB2")

	//time.Sleep(30 * time.Second)

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, clusterSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB})

	nodeA.OpenChannelAndCheck(t, nodeB)

	equivalent := "1001"
	maxPositiveBalance := 100000000
	nodeA.CreateSettlementLineAndCheck(t, nodeB, equivalent, strconv.Itoa(maxPositiveBalance))

	transactionAmount := 10000
	nodeB.CreateTransactionCheckStatus(t, nodeA, equivalent, strconv.Itoa(transactionAmount), vtcp.StatusOK)
	time.Sleep(1 * time.Second)

	nodeA.CheckActiveSettlementLine(t, nodeB, equivalent, strconv.Itoa(maxPositiveBalance), "0", strconv.Itoa(transactionAmount))
	nodeB.CheckActiveSettlementLine(t, nodeA, equivalent, "0", strconv.Itoa(maxPositiveBalance), strconv.Itoa(-transactionAmount))
}
