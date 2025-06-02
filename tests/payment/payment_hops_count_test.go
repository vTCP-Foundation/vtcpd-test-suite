package main

import (
	"context"
	"fmt"
	"testing"

	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
	"github.com/vTCP-Foundation/vtcpd-test-suite/tests/testconfig"
)

var (
	paymentHopsCountNextNodeIndex = 1
)

func getNextIPForPaymentHopsCountTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForPaymentHopsCountTest, paymentHopsCountNextNodeIndex)
	paymentHopsCountNextNodeIndex++
	return ip
}

func TestPaymentHopsCount1(t *testing.T) {
	node1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_1")
	node2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_2")
	node3 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_3")
	node4 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_4")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{node1, node2, node3, node4})

	node1.SetHopsCount(1)

	node1.OpenChannelAndCheck(t, node2)
	node2.OpenChannelAndCheck(t, node3)
	node3.OpenChannelAndCheck(t, node4)

	node2.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "1000")
	node3.CreateAndSetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "1000")
	node4.CreateAndSetSettlementLineAndCheck(t, node3, testconfig.Equivalent, "1000")

	node1.CheckMaxFlow(t, node2, testconfig.Equivalent, "1000")
	node1.CheckMaxFlow(t, node3, testconfig.Equivalent, "1000")
	node1.CheckMaxFlow(t, node4, testconfig.Equivalent, "1000")

}

func TestPaymentHopsCount11(t *testing.T) {
	node1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_1")
	node2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_2")
	node3 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_3")
	node4 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_4")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{node1, node2, node3, node4})

	node1.SetHopsCount(0)

	node1.OpenChannelAndCheck(t, node2)
	node2.OpenChannelAndCheck(t, node3)
	node3.OpenChannelAndCheck(t, node4)

	node2.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "1000")
	node3.CreateAndSetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "1000")
	node4.CreateAndSetSettlementLineAndCheck(t, node3, testconfig.Equivalent, "1000")

	node1.CheckMaxFlow(t, node2, testconfig.Equivalent, "0")
	node1.CheckMaxFlow(t, node3, testconfig.Equivalent, "0")
	node1.CheckMaxFlow(t, node4, testconfig.Equivalent, "0")
}

func TestPaymentHopsCount12(t *testing.T) {
	node1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_1")
	node2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_2")
	node3 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_3")
	node4 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_4")
	node5 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_5")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{node1, node2, node3, node4, node5})

	node1.SetHopsCount(0)

	node1.OpenChannelAndCheck(t, node2)
	node2.OpenChannelAndCheck(t, node3)
	node3.OpenChannelAndCheck(t, node4)
	node4.OpenChannelAndCheck(t, node5)

	node2.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "1000")
	node3.CreateAndSetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "1000")
	node4.CreateAndSetSettlementLineAndCheck(t, node3, testconfig.Equivalent, "1000")
	node5.CreateAndSetSettlementLineAndCheck(t, node4, testconfig.Equivalent, "1000")

	node1.CheckMaxFlow(t, node5, testconfig.Equivalent, "0")
}

func TestPaymentHopsCount2(t *testing.T) {
	node1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_1")
	node2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_2")
	node3 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_3")
	node4 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_4")
	node5 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_5")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{node1, node2, node3, node4, node5})

	node1.SetHopsCount(2)

	node1.OpenChannelAndCheck(t, node2)
	node2.OpenChannelAndCheck(t, node3)
	node3.OpenChannelAndCheck(t, node4)
	node4.OpenChannelAndCheck(t, node5)

	node2.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "1000")
	node3.CreateAndSetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "1000")
	node4.CreateAndSetSettlementLineAndCheck(t, node3, testconfig.Equivalent, "1000")
	node5.CreateAndSetSettlementLineAndCheck(t, node4, testconfig.Equivalent, "1000")

	node1.CheckMaxFlow(t, node2, testconfig.Equivalent, "1000")
	node1.CheckMaxFlow(t, node3, testconfig.Equivalent, "1000")
	node1.CheckMaxFlow(t, node4, testconfig.Equivalent, "1000")
	node1.CheckMaxFlow(t, node5, testconfig.Equivalent, "0")
}

func TestPaymentHopsCount3(t *testing.T) {
	node1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_1")
	node2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_2")
	node3 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_3")
	node4 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_4")
	node5 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_5")
	node6 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_6")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{node1, node2, node3, node4, node5, node6})

	node1.SetHopsCount(3)

	node1.OpenChannelAndCheck(t, node2)
	node2.OpenChannelAndCheck(t, node3)
	node3.OpenChannelAndCheck(t, node4)
	node4.OpenChannelAndCheck(t, node5)
	node5.OpenChannelAndCheck(t, node6)

	node2.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "1000")
	node3.CreateAndSetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "1000")
	node4.CreateAndSetSettlementLineAndCheck(t, node3, testconfig.Equivalent, "1000")
	node5.CreateAndSetSettlementLineAndCheck(t, node4, testconfig.Equivalent, "1000")
	node6.CreateAndSetSettlementLineAndCheck(t, node5, testconfig.Equivalent, "500")

	node1.CheckMaxFlow(t, node6, testconfig.Equivalent, "0")
}

func TestPaymentHopsCount51(t *testing.T) {
	node1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_1")
	node2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_2")
	node3 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_3")
	node4 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_4")
	node5 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_5")
	node6 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_6")
	node7 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_7")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{node1, node2, node3, node4, node5, node6, node7})

	node1.SetHopsCount(5)

	node1.OpenChannelAndCheck(t, node2)
	node2.OpenChannelAndCheck(t, node3)
	node3.OpenChannelAndCheck(t, node4)
	node4.OpenChannelAndCheck(t, node5)
	node5.OpenChannelAndCheck(t, node6)
	node6.OpenChannelAndCheck(t, node7)

	node2.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "1000")
	node3.CreateAndSetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "1000")
	node4.CreateAndSetSettlementLineAndCheck(t, node3, testconfig.Equivalent, "1000")
	node5.CreateAndSetSettlementLineAndCheck(t, node4, testconfig.Equivalent, "1000")
	node6.CreateAndSetSettlementLineAndCheck(t, node5, testconfig.Equivalent, "500")
	node7.CreateAndSetSettlementLineAndCheck(t, node6, testconfig.Equivalent, "500")

	node1.CheckMaxFlow(t, node7, testconfig.Equivalent, "500")
}

func TestPaymentHopsCount52(t *testing.T) {
	node1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_1")
	node2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_2")
	node3 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_3")
	node4 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_4")
	node5 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_5")
	node6 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_6")
	node7 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_7")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{node1, node2, node3, node4, node5, node6, node7})

	node1.SetHopsCount(4)

	node1.OpenChannelAndCheck(t, node2)
	node2.OpenChannelAndCheck(t, node3)
	node3.OpenChannelAndCheck(t, node4)
	node4.OpenChannelAndCheck(t, node5)
	node5.OpenChannelAndCheck(t, node6)
	node6.OpenChannelAndCheck(t, node7)

	node2.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "1000")
	node3.CreateAndSetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "1000")
	node4.CreateAndSetSettlementLineAndCheck(t, node3, testconfig.Equivalent, "1000")
	node5.CreateAndSetSettlementLineAndCheck(t, node4, testconfig.Equivalent, "1000")
	node6.CreateAndSetSettlementLineAndCheck(t, node5, testconfig.Equivalent, "500")
	node7.CreateAndSetSettlementLineAndCheck(t, node6, testconfig.Equivalent, "500")

	node1.CheckMaxFlow(t, node7, testconfig.Equivalent, "0")
}

func TestPayment1HopsCountHop11(t *testing.T) {
	node1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_1")
	node2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_2")
	hop1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_1")
	hop2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_2")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{node1, node2, hop1, hop2})

	node1.SetHopsCount(2)
	hop1.MakeHub(testconfig.Equivalent)
	hop2.MakeHub(testconfig.Equivalent)

	node1.OpenChannelAndCheck(t, hop1)
	hop1.OpenChannelAndCheck(t, hop2)
	hop2.OpenChannelAndCheck(t, node2)

	hop1.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "1000")
	hop2.CreateAndSetSettlementLineAndCheck(t, hop1, testconfig.Equivalent, "1000")
	node2.CreateAndSetSettlementLineAndCheck(t, hop2, testconfig.Equivalent, "1000")

	node1.CheckMaxFlow(t, node2, testconfig.Equivalent, "1000")
}

func TestPayment1HopsCountHop12(t *testing.T) {
	node1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_1")
	node2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_2")
	hop1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_1")
	hop2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_2")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{node1, node2, hop1, hop2})

	node1.SetHopsCount(1)
	hop1.MakeHub(testconfig.Equivalent)
	hop2.MakeHub(testconfig.Equivalent)

	node1.OpenChannelAndCheck(t, hop1)
	hop1.OpenChannelAndCheck(t, hop2)
	hop2.OpenChannelAndCheck(t, node2)

	hop1.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "1000")
	hop2.CreateAndSetSettlementLineAndCheck(t, hop1, testconfig.Equivalent, "1000")
	node2.CreateAndSetSettlementLineAndCheck(t, hop2, testconfig.Equivalent, "1000")

	node1.CheckMaxFlow(t, node2, testconfig.Equivalent, "1000")
}

func TestPayment1HopsCountHop13(t *testing.T) {
	node1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_1")
	node2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_2")
	hop1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_1")
	hop2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_2")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{node1, node2, hop1, hop2})

	node1.SetHopsCount(0)
	hop1.MakeHub(testconfig.Equivalent)
	hop2.MakeHub(testconfig.Equivalent)

	node1.OpenChannelAndCheck(t, hop1)
	hop1.OpenChannelAndCheck(t, hop2)
	hop2.OpenChannelAndCheck(t, node2)

	hop1.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "1000")
	hop2.CreateAndSetSettlementLineAndCheck(t, hop1, testconfig.Equivalent, "1000")
	node2.CreateAndSetSettlementLineAndCheck(t, hop2, testconfig.Equivalent, "1000")

	node1.CheckMaxFlow(t, node2, testconfig.Equivalent, "0")
}

func TestPayment2HopsCountHop21(t *testing.T) {
	node1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_1")
	node2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_2")
	hop1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_1")
	hop2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_2")
	hop3 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_3")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{node1, node2, hop1, hop2, hop3})

	node1.SetHopsCount(3)
	hop1.MakeHub(testconfig.Equivalent)
	hop2.MakeHub(testconfig.Equivalent)
	hop3.MakeHub(testconfig.Equivalent)

	node1.OpenChannelAndCheck(t, hop1)
	hop1.OpenChannelAndCheck(t, hop2)
	hop2.OpenChannelAndCheck(t, hop3)
	hop3.OpenChannelAndCheck(t, node2)

	hop1.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "10000")
	hop2.CreateAndSetSettlementLineAndCheck(t, hop1, testconfig.Equivalent, "10000")
	hop3.CreateAndSetSettlementLineAndCheck(t, hop2, testconfig.Equivalent, "10000")
	node2.CreateAndSetSettlementLineAndCheck(t, hop3, testconfig.Equivalent, "1000")

	node1.CheckMaxFlow(t, node2, testconfig.Equivalent, "1000")
}

func TestPayment2HopsCountHop22(t *testing.T) {
	node1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_1")
	node2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_2")
	hop1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_1")
	hop2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_2")
	hop3 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_3")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{node1, node2, hop1, hop2, hop3})

	node1.SetHopsCount(2)
	hop1.MakeHub(testconfig.Equivalent)
	hop2.MakeHub(testconfig.Equivalent)
	hop3.MakeHub(testconfig.Equivalent)

	node1.OpenChannelAndCheck(t, hop1)
	hop1.OpenChannelAndCheck(t, hop2)
	hop2.OpenChannelAndCheck(t, hop3)
	hop3.OpenChannelAndCheck(t, node2)

	hop1.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "10000")
	hop2.CreateAndSetSettlementLineAndCheck(t, hop1, testconfig.Equivalent, "10000")
	hop3.CreateAndSetSettlementLineAndCheck(t, hop2, testconfig.Equivalent, "10000")
	node2.CreateAndSetSettlementLineAndCheck(t, hop3, testconfig.Equivalent, "1000")

	node1.CheckMaxFlow(t, node2, testconfig.Equivalent, "0")
}

func TestPayment3HopsCountHop31(t *testing.T) {
	node1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_1")
	node2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_2")
	hop1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_1")
	hop2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_2")
	hop3 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_3")
	hop4 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_4")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{node1, node2, hop1, hop2, hop3, hop4})

	node1.SetHopsCount(4)
	hop1.MakeHub(testconfig.Equivalent)
	hop2.MakeHub(testconfig.Equivalent)
	hop3.MakeHub(testconfig.Equivalent)
	hop4.MakeHub(testconfig.Equivalent)

	node1.OpenChannelAndCheck(t, hop1)
	hop1.OpenChannelAndCheck(t, hop2)
	hop2.OpenChannelAndCheck(t, hop3)
	hop3.OpenChannelAndCheck(t, hop4)
	hop4.OpenChannelAndCheck(t, node2)

	hop1.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "10000")
	hop2.CreateAndSetSettlementLineAndCheck(t, hop1, testconfig.Equivalent, "10000")
	hop3.CreateAndSetSettlementLineAndCheck(t, hop2, testconfig.Equivalent, "10000")
	hop4.CreateAndSetSettlementLineAndCheck(t, hop3, testconfig.Equivalent, "1000")
	node2.CreateAndSetSettlementLineAndCheck(t, hop4, testconfig.Equivalent, "1000")

	node1.CheckMaxFlow(t, node2, testconfig.Equivalent, "1000")
}

func TestPayment3HopsCountHop32(t *testing.T) {
	node1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_1")
	node2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_2")
	hop1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_1")
	hop2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_2")
	hop3 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_3")
	hop4 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_4")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{node1, node2, hop1, hop2, hop3, hop4})

	node1.SetHopsCount(3)
	hop1.MakeHub(testconfig.Equivalent)
	hop2.MakeHub(testconfig.Equivalent)
	hop3.MakeHub(testconfig.Equivalent)
	hop4.MakeHub(testconfig.Equivalent)

	node1.OpenChannelAndCheck(t, hop1)
	hop1.OpenChannelAndCheck(t, hop2)
	hop2.OpenChannelAndCheck(t, hop3)
	hop3.OpenChannelAndCheck(t, hop4)
	hop4.OpenChannelAndCheck(t, node2)

	hop1.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "10000")
	hop2.CreateAndSetSettlementLineAndCheck(t, hop1, testconfig.Equivalent, "10000")
	hop3.CreateAndSetSettlementLineAndCheck(t, hop2, testconfig.Equivalent, "10000")
	hop4.CreateAndSetSettlementLineAndCheck(t, hop3, testconfig.Equivalent, "1000")
	node2.CreateAndSetSettlementLineAndCheck(t, hop4, testconfig.Equivalent, "1000")

	node1.CheckMaxFlow(t, node2, testconfig.Equivalent, "0")
}

func TestPayment4HopsCountHop41(t *testing.T) {
	node1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_1")
	node2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_2")
	hop1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_1")
	hop2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_2")
	hop3 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_3")
	hop4 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_4")
	hop5 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_5")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{node1, node2, hop1, hop2, hop3, hop4, hop5})

	node1.SetHopsCount(5)
	hop1.MakeHub(testconfig.Equivalent)
	hop2.MakeHub(testconfig.Equivalent)
	hop3.MakeHub(testconfig.Equivalent)
	hop4.MakeHub(testconfig.Equivalent)
	hop5.MakeHub(testconfig.Equivalent)

	node1.OpenChannelAndCheck(t, hop1)
	hop1.OpenChannelAndCheck(t, hop2)
	hop2.OpenChannelAndCheck(t, hop3)
	hop3.OpenChannelAndCheck(t, hop4)
	hop4.OpenChannelAndCheck(t, hop5)
	hop5.OpenChannelAndCheck(t, node2)

	hop1.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "10000")
	hop2.CreateAndSetSettlementLineAndCheck(t, hop1, testconfig.Equivalent, "10000")
	hop3.CreateAndSetSettlementLineAndCheck(t, hop2, testconfig.Equivalent, "10000")
	hop4.CreateAndSetSettlementLineAndCheck(t, hop3, testconfig.Equivalent, "10000")
	hop5.CreateAndSetSettlementLineAndCheck(t, hop4, testconfig.Equivalent, "10000")
	node2.CreateAndSetSettlementLineAndCheck(t, hop5, testconfig.Equivalent, "1000")

	node1.CheckMaxFlow(t, node2, testconfig.Equivalent, "1000")
}

func TestPayment4HopsCountHop42(t *testing.T) {
	node1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_1")
	node2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_2")
	hop1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_1")
	hop2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_2")
	hop3 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_3")
	hop4 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_4")
	hop5 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_5")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{node1, node2, hop1, hop2, hop3, hop4, hop5})

	node1.SetHopsCount(4)
	hop1.MakeHub(testconfig.Equivalent)
	hop2.MakeHub(testconfig.Equivalent)
	hop3.MakeHub(testconfig.Equivalent)
	hop4.MakeHub(testconfig.Equivalent)
	hop5.MakeHub(testconfig.Equivalent)

	node1.OpenChannelAndCheck(t, hop1)
	hop1.OpenChannelAndCheck(t, hop2)
	hop2.OpenChannelAndCheck(t, hop3)
	hop3.OpenChannelAndCheck(t, hop4)
	hop4.OpenChannelAndCheck(t, hop5)
	hop5.OpenChannelAndCheck(t, node2)

	hop1.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "10000")
	hop2.CreateAndSetSettlementLineAndCheck(t, hop1, testconfig.Equivalent, "10000")
	hop3.CreateAndSetSettlementLineAndCheck(t, hop2, testconfig.Equivalent, "10000")
	hop4.CreateAndSetSettlementLineAndCheck(t, hop3, testconfig.Equivalent, "10000")
	hop5.CreateAndSetSettlementLineAndCheck(t, hop4, testconfig.Equivalent, "10000")
	node2.CreateAndSetSettlementLineAndCheck(t, hop5, testconfig.Equivalent, "1000")

	node1.CheckMaxFlow(t, node2, testconfig.Equivalent, "0")
}

func TestPayment5HopsCountHop(t *testing.T) {
	node1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_1")
	node2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_2")
	hop1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_1")
	hop2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_2")
	hop3 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_3")
	hop4 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_4")
	hop5 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_5")
	hop6 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_6")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{node1, node2, hop1, hop2, hop3, hop4, hop5, hop6})

	node1.SetHopsCount(5)
	hop1.MakeHub(testconfig.Equivalent)
	hop2.MakeHub(testconfig.Equivalent)
	hop3.MakeHub(testconfig.Equivalent)
	hop4.MakeHub(testconfig.Equivalent)
	hop5.MakeHub(testconfig.Equivalent)
	hop6.MakeHub(testconfig.Equivalent)

	node1.OpenChannelAndCheck(t, hop1)
	hop1.OpenChannelAndCheck(t, hop2)
	hop2.OpenChannelAndCheck(t, hop3)
	hop3.OpenChannelAndCheck(t, hop4)
	hop4.OpenChannelAndCheck(t, hop5)
	hop5.OpenChannelAndCheck(t, hop6)
	hop6.OpenChannelAndCheck(t, node2)

	hop1.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "10000")
	hop2.CreateAndSetSettlementLineAndCheck(t, hop1, testconfig.Equivalent, "10000")
	hop3.CreateAndSetSettlementLineAndCheck(t, hop2, testconfig.Equivalent, "10000")
	hop4.CreateAndSetSettlementLineAndCheck(t, hop3, testconfig.Equivalent, "10000")
	hop5.CreateAndSetSettlementLineAndCheck(t, hop4, testconfig.Equivalent, "10000")
	hop6.CreateAndSetSettlementLineAndCheck(t, hop5, testconfig.Equivalent, "10000")
	node2.CreateAndSetSettlementLineAndCheck(t, hop6, testconfig.Equivalent, "1000")

	node1.CheckMaxFlow(t, node2, testconfig.Equivalent, "0")
}

func TestPayment6HopsCountHop5(t *testing.T) {
	node1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_1")
	node2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_2")
	node3 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "node_3")
	hop1 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_1")
	hop2 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_2")
	hop3 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_3")
	hop4 := vtcp.NewNode(t, getNextIPForPaymentHopsCountTest(), "hop_4")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{node1, node2, node3, hop1, hop2, hop3, hop4})

	node1.SetHopsCount(5)
	hop1.MakeHub(testconfig.Equivalent)
	hop2.MakeHub(testconfig.Equivalent)
	hop3.MakeHub(testconfig.Equivalent)
	hop4.MakeHub(testconfig.Equivalent)

	node1.OpenChannelAndCheck(t, hop1)
	hop1.OpenChannelAndCheck(t, hop2)
	hop2.OpenChannelAndCheck(t, node2)
	node2.OpenChannelAndCheck(t, hop3)
	hop3.OpenChannelAndCheck(t, hop4)
	hop4.OpenChannelAndCheck(t, node3)

	hop1.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "1000")
	hop2.CreateAndSetSettlementLineAndCheck(t, hop1, testconfig.Equivalent, "10000")
	node2.CreateAndSetSettlementLineAndCheck(t, hop2, testconfig.Equivalent, "10000")
	hop3.CreateAndSetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "10000")
	hop4.CreateAndSetSettlementLineAndCheck(t, hop3, testconfig.Equivalent, "10000")
	node3.CreateAndSetSettlementLineAndCheck(t, hop4, testconfig.Equivalent, "1000")

	node1.CheckMaxFlow(t, node3, testconfig.Equivalent, "0")
	node1.CheckMaxFlow(t, node2, testconfig.Equivalent, "1000")
	node2.CheckMaxFlow(t, node3, testconfig.Equivalent, "1000")
}
