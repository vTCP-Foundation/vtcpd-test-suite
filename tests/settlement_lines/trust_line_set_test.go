package main

import (
	"context"
	"strconv"
	"testing"
	"time"

	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
	"github.com/vTCP-Foundation/vtcpd-test-suite/tests/testconfig"
)

func TestTrustLineSet(t *testing.T) {
	nodeA := vtcp.NewNode(t, "172.18.100.2", "nodeA")
	nodeB := vtcp.NewNode(t, "172.18.100.3", "nodeB")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{nodeA, nodeB})

	nodeA.OpenChannelAndCheck(t, nodeB)

	equivalent := "1001"
	maxPositiveBalance := 100000000
	nodeA.CreateSettlementLineAndCheck(t, nodeB, equivalent, strconv.Itoa(maxPositiveBalance))

	nodeA.CheckSettlementLineForSync(t, nodeB, equivalent)
	nodeB.CheckSettlementLineForSync(t, nodeA, equivalent)
	vtcp.CheckSettlementLineForSyncBatch(t, []*vtcp.Node{nodeA, nodeB}, equivalent, 0)

	time.Sleep(3 * time.Second)
	transactionAmount := 10000
	nodeB.CreateTransactionCheckStatus(t, nodeA, equivalent, strconv.Itoa(transactionAmount), vtcp.StatusOK)
	time.Sleep(1 * time.Second)

	nodeA.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 2, 1, 0)
	nodeB.CheckPaymentTransaction(t, vtcp.PaymentObservingStateNoInfo, 1, 2, 0, 1)

	nodeA.CheckActiveSettlementLine(t, nodeB, equivalent, strconv.Itoa(maxPositiveBalance), "0", strconv.Itoa(transactionAmount))
	nodeB.CheckActiveSettlementLine(t, nodeA, equivalent, "0", strconv.Itoa(maxPositiveBalance), strconv.Itoa(-transactionAmount))
}
