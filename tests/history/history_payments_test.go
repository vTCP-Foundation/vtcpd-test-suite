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
	historyPaymentsNextNodeIndex = 1
)

func getNextIPForHistoryPaymentsTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForHistoryPaymentsTest, historyPaymentsNextNodeIndex)
	historyPaymentsNextNodeIndex++
	return ip
}

func TestHistoryAdditionalPayments(t *testing.T) {
	node1 := vtcp.NewNode(t, getNextIPForHistoryPaymentsTest(), "node_1")
	node2 := vtcp.NewNode(t, getNextIPForHistoryPaymentsTest(), "node_2")
	node3 := vtcp.NewNode(t, getNextIPForHistoryPaymentsTest(), "node_3")
	node4 := vtcp.NewNode(t, getNextIPForHistoryPaymentsTest(), "node_4")
	node5 := vtcp.NewNode(t, getNextIPForHistoryPaymentsTest(), "node_5")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{node1, node2, node3, node4, node5})

	node1.SetHopsCount(4)

	// Open channels
	node1.OpenChannelAndCheck(t, node2)
	time.Sleep(1 * time.Second)
	node1.OpenChannelAndCheck(t, node3)
	time.Sleep(1 * time.Second)
	node2.OpenChannelAndCheck(t, node4)
	time.Sleep(1 * time.Second)
	node3.OpenChannelAndCheck(t, node4)
	time.Sleep(1 * time.Second)
	node2.OpenChannelAndCheck(t, node5)
	time.Sleep(1 * time.Second)
	node4.OpenChannelAndCheck(t, node5)
	time.Sleep(1 * time.Second)

	// Set settlement lines
	node2.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "500")
	node3.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "500")
	node4.CreateAndSetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "100")
	node4.CreateAndSetSettlementLineAndCheck(t, node3, testconfig.Equivalent, "800")
	node5.CreateAndSetSettlementLineAndCheck(t, node4, testconfig.Equivalent, "1000")
	node5.CreateAndSetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "400")

	time.Sleep(2 * time.Second)

	// Check max flow and perform transactions
	node1.CheckMaxFlow(t, node5, testconfig.Equivalent, "1000")
	node1.CreateTransactionCheckStatus(t, node5, testconfig.Equivalent, "500", vtcp.StatusOK)

	node1.CheckMaxFlow(t, node5, testconfig.Equivalent, "500")
	node5.CreateTransactionCheckStatus(t, node1, testconfig.Equivalent, "500", vtcp.StatusOK)

	node1.CheckMaxFlow(t, node5, testconfig.Equivalent, "1000")
	node1.CreateTransactionCheckStatus(t, node5, testconfig.Equivalent, "700", vtcp.StatusOK)

	node1.CheckMaxFlow(t, node5, testconfig.Equivalent, "300")

	// Check history additional payments for node2
	jsonRes := node2.HistoryAdditionalPayments(t)
	if jsonRes["data"].(map[string]interface{})["count"].(float64) != 3 {
		t.Errorf("Expected count 3, got %v", jsonRes["data"].(map[string]interface{})["count"])
	}

	records := jsonRes["data"].(map[string]interface{})["records"].([]interface{})
	if int(records[0].(map[string]interface{})["amount"].(float64)) != 500 {
		t.Errorf("Expected amount 500, got %v", records[0].(map[string]interface{})["amount"])
	}
	if int(records[1].(map[string]interface{})["amount"].(float64)) != 500 {
		t.Errorf("Expected amount 500, got %v", records[1].(map[string]interface{})["amount"])
	}
	if int(records[2].(map[string]interface{})["amount"].(float64)) != 500 {
		t.Errorf("Expected amount 500, got %v", records[2].(map[string]interface{})["amount"])
	}

	// Check history additional payments for node3
	jsonRes = node3.HistoryAdditionalPayments(t)
	if int(jsonRes["data"].(map[string]interface{})["count"].(float64)) != 1 {
		t.Errorf("Expected count 1, got %v", jsonRes["data"].(map[string]interface{})["count"])
	}

	records = jsonRes["data"].(map[string]interface{})["records"].([]interface{})
	if int(records[0].(map[string]interface{})["amount"].(float64)) != 200 {
		t.Errorf("Expected amount 200, got %v", records[0].(map[string]interface{})["amount"])
	}
}

func TestHistoryPayments(t *testing.T) {
	node1 := vtcp.NewNode(t, getNextIPForHistoryPaymentsTest(), "node_1")
	node2 := vtcp.NewNode(t, getNextIPForHistoryPaymentsTest(), "node_2")
	node3 := vtcp.NewNode(t, getNextIPForHistoryPaymentsTest(), "node_3")
	node4 := vtcp.NewNode(t, getNextIPForHistoryPaymentsTest(), "node_4")
	node5 := vtcp.NewNode(t, getNextIPForHistoryPaymentsTest(), "node_5")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{node1, node2, node3, node4, node5})

	node1.SetHopsCount(4)

	// Open channels
	node1.OpenChannelAndCheck(t, node2)
	time.Sleep(1 * time.Second)
	node1.OpenChannelAndCheck(t, node3)
	time.Sleep(1 * time.Second)
	node2.OpenChannelAndCheck(t, node4)
	time.Sleep(1 * time.Second)
	node3.OpenChannelAndCheck(t, node4)
	time.Sleep(1 * time.Second)
	node2.OpenChannelAndCheck(t, node5)
	time.Sleep(1 * time.Second)
	node4.OpenChannelAndCheck(t, node5)
	time.Sleep(1 * time.Second)

	// Set settlement lines
	node2.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "500")
	node3.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "500")
	node4.CreateAndSetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "100")
	node4.CreateAndSetSettlementLineAndCheck(t, node3, testconfig.Equivalent, "800")
	node5.CreateAndSetSettlementLineAndCheck(t, node4, testconfig.Equivalent, "1000")
	node5.CreateAndSetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "400")

	time.Sleep(2 * time.Second)

	// Perform transactions
	node1.CheckMaxFlow(t, node5, testconfig.Equivalent, "1000")
	node1.CreateTransactionCheckStatus(t, node5, testconfig.Equivalent, "500", vtcp.StatusOK)
	node1.CheckMaxFlow(t, node5, testconfig.Equivalent, "500")
	node5.CreateTransactionCheckStatus(t, node1, testconfig.Equivalent, "500", vtcp.StatusOK)
	node1.CheckMaxFlow(t, node5, testconfig.Equivalent, "1000")
	node1.CreateTransactionCheckStatus(t, node5, testconfig.Equivalent, "700", vtcp.StatusOK)
	node1.CheckMaxFlow(t, node5, testconfig.Equivalent, "300")

	// Check history payments
	jsonRes := node1.HistoryPayments(t)
	if jsonRes["data"].(map[string]interface{})["count"].(float64) != 3 {
		t.Errorf("Expected count 3, got %v", jsonRes["data"].(map[string]interface{})["count"])
	}

	records := jsonRes["data"].(map[string]interface{})["records"].([]interface{})
	if int(records[2].(map[string]interface{})["amount"].(float64)) != 500 {
		t.Errorf("Expected amount 500, got %v", records[2].(map[string]interface{})["amount"])
	}
	if int(records[1].(map[string]interface{})["amount"].(float64)) != 500 {
		t.Errorf("Expected amount 500, got %v", records[1].(map[string]interface{})["amount"])
	}
	if int(records[0].(map[string]interface{})["amount"].(float64)) != 700 {
		t.Errorf("Expected amount 700, got %v", records[0].(map[string]interface{})["amount"])
	}

	// Check balance after operations
	// pay 500 from node_1 to node_5
	if int(records[2].(map[string]interface{})["balance_after_operation"].(float64)) != -500 {
		t.Errorf("Expected balance_after_operation -500, got %v", records[2].(map[string]interface{})["balance_after_operation"])
	}
	// pay 500 from node_5 to node_1
	if int(records[1].(map[string]interface{})["balance_after_operation"].(float64)) != 0 {
		t.Errorf("Expected balance_after_operation 0, got %v", records[1].(map[string]interface{})["balance_after_operation"])
	}
	// pay 700 from node_1 to node_5
	if int(records[0].(map[string]interface{})["balance_after_operation"].(float64)) != -700 {
		t.Errorf("Expected balance_after_operation -700, got %v", records[0].(map[string]interface{})["balance_after_operation"])
	}
}

func TestHistoryPaymentsAllEquivalents(t *testing.T) {
	node1 := vtcp.NewNode(t, getNextIPForHistoryPaymentsTest(), "node_1")
	node2 := vtcp.NewNode(t, getNextIPForHistoryPaymentsTest(), "node_2")
	node3 := vtcp.NewNode(t, getNextIPForHistoryPaymentsTest(), "node_3")
	node4 := vtcp.NewNode(t, getNextIPForHistoryPaymentsTest(), "node_4")
	node5 := vtcp.NewNode(t, getNextIPForHistoryPaymentsTest(), "node_5")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, []*vtcp.Node{node1, node2, node3, node4, node5})

	node1.SetHopsCount(4)

	// Open channels
	node1.OpenChannelAndCheck(t, node2)
	time.Sleep(1 * time.Second)
	node1.CreateSettlementLine(t, node2, "2")
	time.Sleep(1 * time.Second)
	node1.OpenChannelAndCheck(t, node3)
	time.Sleep(1 * time.Second)
	node2.OpenChannelAndCheck(t, node4)
	time.Sleep(1 * time.Second)
	node3.OpenChannelAndCheck(t, node4)
	time.Sleep(1 * time.Second)
	node2.OpenChannelAndCheck(t, node5)
	time.Sleep(1 * time.Second)
	node2.CreateSettlementLine(t, node5, "2")
	time.Sleep(1 * time.Second)
	node4.OpenChannelAndCheck(t, node5)
	time.Sleep(1 * time.Second)

	// Set settlement lines
	node2.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "500")
	node2.CreateAndSetSettlementLineAndCheck(t, node1, "2", "500")
	node3.CreateAndSetSettlementLineAndCheck(t, node1, testconfig.Equivalent, "500")
	node4.CreateAndSetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "100")
	node4.CreateAndSetSettlementLineAndCheck(t, node3, testconfig.Equivalent, "800")
	node5.CreateAndSetSettlementLineAndCheck(t, node4, testconfig.Equivalent, "1000")
	node5.CreateAndSetSettlementLineAndCheck(t, node2, testconfig.Equivalent, "400")
	node5.CreateAndSetSettlementLineAndCheck(t, node2, "2", "300")

	time.Sleep(2 * time.Second)

	node1.CheckMaxFlow(t, node5, "2", "300")

	// Perform transactions
	node1.CreateTransactionCheckStatus(t, node5, testconfig.Equivalent, "500", vtcp.StatusOK)
	time.Sleep(1 * time.Second)
	node1.CreateTransactionCheckStatus(t, node5, "2", "200", vtcp.StatusOK)
	time.Sleep(1 * time.Second)
	node1.CreateTransactionCheckStatus(t, node5, "2", "100", vtcp.StatusOK)
	time.Sleep(1 * time.Second)

	// Check history payments all equivalents
	jsonRes := node1.HistoryPaymentsAllEquivalents(t)
	if jsonRes["data"].(map[string]interface{})["count"].(float64) != 3 {
		t.Errorf("Expected count 3, got %v", jsonRes["data"].(map[string]interface{})["count"])
	}

	records := jsonRes["data"].(map[string]interface{})["records"].([]interface{})
	if int(records[0].(map[string]interface{})["equivalent"].(float64)) != 2 {
		t.Errorf("Expected equivalent 2, got %v", records[0].(map[string]interface{})["equivalent"])
	}
	if int(records[0].(map[string]interface{})["balance_after_operation"].(float64)) != -300 {
		t.Errorf("Expected balance_after_operation -300, got %v", records[0].(map[string]interface{})["balance_after_operation"])
	}

	if int(records[1].(map[string]interface{})["equivalent"].(float64)) != 2 {
		t.Errorf("Expected equivalent 2, got %v", records[1].(map[string]interface{})["equivalent"])
	}
	if int(records[1].(map[string]interface{})["balance_after_operation"].(float64)) != -200 {
		t.Errorf("Expected balance_after_operation -200, got %v", records[1].(map[string]interface{})["balance_after_operation"])
	}

	if int(records[2].(map[string]interface{})["equivalent"].(float64)) != 1 {
		t.Errorf("Expected equivalent 1, got %v", records[2].(map[string]interface{})["equivalent"])
	}
	if int(records[2].(map[string]interface{})["balance_after_operation"].(float64)) != -500 {
		t.Errorf("Expected balance_after_operation -500, got %v", records[2].(map[string]interface{})["balance_after_operation"])
	}
}
