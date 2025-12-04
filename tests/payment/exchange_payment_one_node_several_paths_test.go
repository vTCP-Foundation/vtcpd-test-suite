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
	// exchangePaymentOneNodeSeveralPathsNextNodeIndex is used to assign unique IP addresses to nodes across different test functions in this file.
	exchangePaymentOneNodeSeveralPathsNextNodeIndex = 1
)

func getNextIPForExchangePaymentOneNodeSeveralPathsTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForExchangePaymentOneNodeSeveralPaths, exchangePaymentOneNodeSeveralPathsNextNodeIndex)
	exchangePaymentOneNodeSeveralPathsNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForExchangePaymentOneNodeSeveralPathsTest(t *testing.T) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, 6)
	for i := range 6 {
		nodes[i] = vtcp.NewNode(t, getNextIPForExchangePaymentOneNodeSeveralPathsTest(), fmt.Sprintf("node%d", i+1))
	}

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, &testconfig.GSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	cluster.RunNodes(ctx, t, nodes, false)

	// Setup topology according to Python version
	nodes[1].OpenChannelAndCheck(t, nodes[0])
	nodes[2].OpenChannelAndCheck(t, nodes[1])
	nodes[3].OpenChannelAndCheck(t, nodes[0])
	nodes[1].OpenChannelAndCheck(t, nodes[3])
	nodes[4].OpenChannelAndCheck(t, nodes[0])
	nodes[1].OpenChannelAndCheck(t, nodes[4])

	nodes[5].OpenChannelAndCheck(t, nodes[2])

	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.ExchangeEquivalent, "3000")
	nodes[2].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.ExchangeEquivalent, "10000")
	nodes[3].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.ExchangeEquivalent, "2000")
	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[3], testconfig.ExchangeEquivalent, "4000")
	nodes[4].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.ExchangeEquivalent, "5000")
	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[4], testconfig.ExchangeEquivalent, "3000")

	nodes[5].CreateAndSetSettlementLineAndCheck(t, nodes[2], testconfig.Equivalent, "3500")

	return nodes, cluster
}

// use 2 paths for payment
func Test1ExchangePaymentOneNodeSeveralPaths(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentOneNodeSeveralPathsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "2000", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-3000")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "10000", "-4000")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[3], testconfig.ExchangeEquivalent, "0", "2000", "0")
	nodes[3].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "4000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[3], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[4], testconfig.ExchangeEquivalent, "0", "5000", "-1010")
	nodes[4].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[4].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-1010")
	nodes[1].CheckSettlementLineForSync(t, nodes[4], testconfig.ExchangeEquivalent)

	nodes[2].CheckActiveSettlementLine(t, nodes[5], testconfig.Equivalent, "0", "3500", "-2000")
	nodes[5].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "1500")
}

// use 3 paths for payment
func Test1aExchangePaymentOneNodeSeveralPaths(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentOneNodeSeveralPathsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "3000", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-3000")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "10000", "-6000")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[3], testconfig.ExchangeEquivalent, "0", "2000", "-10")
	nodes[3].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "4000", "-10")
	nodes[1].CheckSettlementLineForSync(t, nodes[3], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[4], testconfig.ExchangeEquivalent, "0", "5000", "-3000")
	nodes[4].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[4].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-3000")
	nodes[1].CheckSettlementLineForSync(t, nodes[4], testconfig.ExchangeEquivalent)

	nodes[2].CheckActiveSettlementLine(t, nodes[5], testconfig.Equivalent, "0", "3500", "-3000")
	nodes[5].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "500")
}

// payment amount more than max allowable amount
func Test1bExchangePaymentOneNodeSeveralPaths(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentOneNodeSeveralPathsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "2000", testconfig.ExchangeEquivalent, "4000", vtcp.StatusMoreThanMaxAllowableAmount)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "10000", "0")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[3], testconfig.ExchangeEquivalent, "0", "2000", "0")
	nodes[3].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "4000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[3], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[4], testconfig.ExchangeEquivalent, "0", "5000", "0")
	nodes[4].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[4].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[4], testconfig.ExchangeEquivalent)

	nodes[2].CheckActiveSettlementLine(t, nodes[5], testconfig.Equivalent, "0", "3500", "0")
	nodes[5].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "3500")
}

// increase commissions during payment using 2 paths
func Test2ExchangePaymentOneNodeSeveralPaths(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentOneNodeSeveralPathsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "3500")

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 20},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "2000", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-3000")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "10000", "-4000")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[3], testconfig.ExchangeEquivalent, "0", "2000", "0")
	nodes[3].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "4000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[3], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[4], testconfig.ExchangeEquivalent, "0", "5000", "-1020")
	nodes[4].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[4].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-1020")
	nodes[1].CheckSettlementLineForSync(t, nodes[4], testconfig.ExchangeEquivalent)

	nodes[2].CheckActiveSettlementLine(t, nodes[5], testconfig.Equivalent, "0", "3500", "-2000")
	nodes[5].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "1500")
}

// increase commissions during payment using 3 paths
func Test2aExchangePaymentOneNodeSeveralPaths(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentOneNodeSeveralPathsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "3500")

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 20},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "3000", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-3000")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "10000", "-6000")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[3], testconfig.ExchangeEquivalent, "0", "2000", "-20")
	nodes[3].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "4000", "-20")
	nodes[1].CheckSettlementLineForSync(t, nodes[3], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[4], testconfig.ExchangeEquivalent, "0", "5000", "-3000")
	nodes[4].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[4].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-3000")
	nodes[1].CheckSettlementLineForSync(t, nodes[4], testconfig.ExchangeEquivalent)

	nodes[2].CheckActiveSettlementLine(t, nodes[5], testconfig.Equivalent, "0", "3500", "-3000")
	nodes[5].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "500")
}

// increase commissions during payment which lead to more than max allowable amount
func Test2bExchangePaymentOneNodeSeveralPaths(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentOneNodeSeveralPathsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "3500")

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 20},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "2000", testconfig.ExchangeEquivalent, "4015", vtcp.StatusMoreThanMaxAllowableAmount)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "10000", "0")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[3], testconfig.ExchangeEquivalent, "0", "2000", "0")
	nodes[3].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "4000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[3], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[4], testconfig.ExchangeEquivalent, "0", "5000", "0")
	nodes[4].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[4].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[4], testconfig.ExchangeEquivalent)

	nodes[2].CheckActiveSettlementLine(t, nodes[5], testconfig.Equivalent, "0", "3500", "0")
	nodes[5].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "3500")
}

// decrease commissions during payment using 2 paths
func Test3ExchangePaymentOneNodeSeveralPaths(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentOneNodeSeveralPathsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "3500")

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 5},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "2000", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-2995")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "10000", "-4000")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[3], testconfig.ExchangeEquivalent, "0", "2000", "0")
	nodes[3].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "4000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[3], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[4], testconfig.ExchangeEquivalent, "0", "5000", "-1010")
	nodes[4].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[4].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-1010")
	nodes[1].CheckSettlementLineForSync(t, nodes[4], testconfig.ExchangeEquivalent)

	nodes[2].CheckActiveSettlementLine(t, nodes[5], testconfig.Equivalent, "0", "3500", "-2000")
	nodes[5].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "1500")
}

// decrease commissions during payment using 3 paths
func Test3aExchangePaymentOneNodeSeveralPaths(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentOneNodeSeveralPathsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "3500")

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 5},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "3000", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-2995")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "10000", "-6000")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[3], testconfig.ExchangeEquivalent, "0", "2000", "-10")
	nodes[3].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "4000", "-10")
	nodes[1].CheckSettlementLineForSync(t, nodes[3], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[4], testconfig.ExchangeEquivalent, "0", "5000", "-3000")
	nodes[4].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[4].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-3000")
	nodes[1].CheckSettlementLineForSync(t, nodes[4], testconfig.ExchangeEquivalent)

	nodes[2].CheckActiveSettlementLine(t, nodes[5], testconfig.Equivalent, "0", "3500", "-3000")
	nodes[5].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "500")
}

// decrease commissions during payment with max allowable amount
func Test3bExchangePaymentOneNodeSeveralPaths(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentOneNodeSeveralPathsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "3500")

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 5},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "2000", testconfig.ExchangeEquivalent, "4020", vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-2995")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "10000", "-4000")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[3], testconfig.ExchangeEquivalent, "0", "2000", "0")
	nodes[3].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "4000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[3], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[4], testconfig.ExchangeEquivalent, "0", "5000", "-1010")
	nodes[4].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[4].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-1010")
	nodes[1].CheckSettlementLineForSync(t, nodes[4], testconfig.ExchangeEquivalent)

	nodes[2].CheckActiveSettlementLine(t, nodes[5], testconfig.Equivalent, "0", "3500", "-2000")
	nodes[5].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "1500")
}

// disappear commissions during payment using 2 paths
func Test4ExchangePaymentOneNodeSeveralPaths(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentOneNodeSeveralPathsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "3500")

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 0},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "2000", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-2990")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "10000", "-4000")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[3], testconfig.ExchangeEquivalent, "0", "2000", "0")
	nodes[3].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "4000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[3], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[4], testconfig.ExchangeEquivalent, "0", "5000", "-1010")
	nodes[4].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[4].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-1010")
	nodes[1].CheckSettlementLineForSync(t, nodes[4], testconfig.ExchangeEquivalent)

	nodes[2].CheckActiveSettlementLine(t, nodes[5], testconfig.Equivalent, "0", "3500", "-2000")
	nodes[5].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "1500")
}

// disappear commissions during payment using 3 paths
func Test4aExchangePaymentOneNodeSeveralPaths(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentOneNodeSeveralPathsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "3500")

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 0},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "3000", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-2990")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "10000", "-6000")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[3], testconfig.ExchangeEquivalent, "0", "2000", "-10")
	nodes[3].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "4000", "-10")
	nodes[1].CheckSettlementLineForSync(t, nodes[3], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[4], testconfig.ExchangeEquivalent, "0", "5000", "-3000")
	nodes[4].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[4].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-3000")
	nodes[1].CheckSettlementLineForSync(t, nodes[4], testconfig.ExchangeEquivalent)

	nodes[2].CheckActiveSettlementLine(t, nodes[5], testconfig.Equivalent, "0", "3500", "-3000")
	nodes[5].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "500")
}

// appear commissions during payment using 2 paths
func Test5ExchangePaymentOneNodeSeveralPaths(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentOneNodeSeveralPathsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "3500")

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "2000", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-3000")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "10000", "-4000")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[3], testconfig.ExchangeEquivalent, "0", "2000", "0")
	nodes[3].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "4000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[3], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[4], testconfig.ExchangeEquivalent, "0", "5000", "-1010")
	nodes[4].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[4].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-1010")
	nodes[1].CheckSettlementLineForSync(t, nodes[4], testconfig.ExchangeEquivalent)

	nodes[2].CheckActiveSettlementLine(t, nodes[5], testconfig.Equivalent, "0", "3500", "-2000")
	nodes[5].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "1500")
}

// appear commissions during payment using 3 paths
func Test5aExchangePaymentOneNodeSeveralPaths(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentOneNodeSeveralPathsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "3500")

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "3000", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-3000")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "10000", "-6000")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[3], testconfig.ExchangeEquivalent, "0", "2000", "-10")
	nodes[3].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "4000", "-10")
	nodes[1].CheckSettlementLineForSync(t, nodes[3], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[4], testconfig.ExchangeEquivalent, "0", "5000", "-3000")
	nodes[4].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[4].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-3000")
	nodes[1].CheckSettlementLineForSync(t, nodes[4], testconfig.ExchangeEquivalent)

	nodes[2].CheckActiveSettlementLine(t, nodes[5], testconfig.Equivalent, "0", "3500", "-3000")
	nodes[5].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "500")
}

// appear commissions during payment which lead to more than max allowable amount
func Test5bExchangePaymentOneNodeSeveralPaths(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentOneNodeSeveralPathsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "3500")

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "2000", testconfig.ExchangeEquivalent, "4005", vtcp.StatusMoreThanMaxAllowableAmount)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "10000", "0")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[3], testconfig.ExchangeEquivalent, "0", "2000", "0")
	nodes[3].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "4000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[3], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[4], testconfig.ExchangeEquivalent, "0", "5000", "0")
	nodes[4].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[4].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[4], testconfig.ExchangeEquivalent)

	nodes[2].CheckActiveSettlementLine(t, nodes[5], testconfig.Equivalent, "0", "3500", "0")
	nodes[5].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "3500")
}

// increase exchange rate during payment using 2 paths
func Test6ExchangePaymentOneNodeSeveralPaths(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentOneNodeSeveralPathsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "3500")

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "6", -1, nil, nil, vtcp.StatusOK)

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "2000", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-2502")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "10000", "-3334")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[3], testconfig.ExchangeEquivalent, "0", "2000", "0")
	nodes[3].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "4000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[3], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[4], testconfig.ExchangeEquivalent, "0", "5000", "-842")
	nodes[4].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[4].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-842")
	nodes[1].CheckSettlementLineForSync(t, nodes[4], testconfig.ExchangeEquivalent)

	nodes[2].CheckActiveSettlementLine(t, nodes[5], testconfig.Equivalent, "0", "3500", "-2000")
	nodes[5].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "1500")
}

// increase exchange rate during payment using 3 paths
func Test6aExchangePaymentOneNodeSeveralPaths(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentOneNodeSeveralPathsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "3500")

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "6", -1, nil, nil, vtcp.StatusOK)

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "3400", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-2502")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "10000", "-5667")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[3], testconfig.ExchangeEquivalent, "0", "2000", "-665")
	nodes[3].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "4000", "-665")
	nodes[1].CheckSettlementLineForSync(t, nodes[3], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[4], testconfig.ExchangeEquivalent, "0", "5000", "-2510")
	nodes[4].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[4].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-2510")
	nodes[1].CheckSettlementLineForSync(t, nodes[4], testconfig.ExchangeEquivalent)

	nodes[2].CheckActiveSettlementLine(t, nodes[5], testconfig.Equivalent, "0", "3500", "-3400")
	nodes[5].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "99")
}

// decrease exchange rate during payment using 2 paths
func Test7ExchangePaymentOneNodeSeveralPaths(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentOneNodeSeveralPathsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "3500")

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "4", -1, nil, nil, vtcp.StatusOK)

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "2000", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-3000")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "10000", "-5000")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[3], testconfig.ExchangeEquivalent, "0", "2000", "0")
	nodes[3].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "4000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[3], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[4], testconfig.ExchangeEquivalent, "0", "5000", "-2010")
	nodes[4].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[4].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-2010")
	nodes[1].CheckSettlementLineForSync(t, nodes[4], testconfig.ExchangeEquivalent)

	nodes[2].CheckActiveSettlementLine(t, nodes[5], testconfig.Equivalent, "0", "3500", "-2000")
	nodes[5].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "1192")
}

// decrease exchange rate during payment using 3 paths
// does not work. in practice receive 412 status code. need to investigate
func Test7aExchangePaymentOneNodeSeveralPaths(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentOneNodeSeveralPathsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "3500")

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "4", -1, nil, nil, vtcp.StatusOK)

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "3000", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-3000")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "10000", "-7500")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[3], testconfig.ExchangeEquivalent, "0", "2000", "-1510")
	nodes[3].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "4000", "-1510")
	nodes[1].CheckSettlementLineForSync(t, nodes[3], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[4], testconfig.ExchangeEquivalent, "0", "5000", "-3000")
	nodes[4].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[4].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-3000")
	nodes[1].CheckSettlementLineForSync(t, nodes[4], testconfig.ExchangeEquivalent)

	nodes[2].CheckActiveSettlementLine(t, nodes[5], testconfig.Equivalent, "0", "3500", "-3000")
	nodes[5].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "192")
}

// decrease exchange rate which lead to more than max allowable amount
func Test7bExchangePaymentOneNodeSeveralPaths(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentOneNodeSeveralPathsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "3500")

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "4", -1, nil, nil, vtcp.StatusOK)

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "2000", testconfig.ExchangeEquivalent, "4005", vtcp.StatusMoreThanMaxAllowableAmount)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "10000", "0")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[3], testconfig.ExchangeEquivalent, "0", "2000", "0")
	nodes[3].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "4000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[3], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[4], testconfig.ExchangeEquivalent, "0", "5000", "0")
	nodes[4].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[4].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[4], testconfig.ExchangeEquivalent)

	nodes[2].CheckActiveSettlementLine(t, nodes[5], testconfig.Equivalent, "0", "3500", "0")
	nodes[5].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "3196")
}

// disappear exchange rate during payment
func Test8ExchangePaymentOneNodeSeveralPaths(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentOneNodeSeveralPathsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "3500")

	nodes[2].ClearExchangeRates(t)

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "2000", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "10000", "0")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[3], testconfig.ExchangeEquivalent, "0", "2000", "0")
	nodes[3].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "4000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[3], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[4], testconfig.ExchangeEquivalent, "0", "5000", "0")
	nodes[4].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[4].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[4], testconfig.ExchangeEquivalent)

	nodes[2].CheckActiveSettlementLine(t, nodes[5], testconfig.Equivalent, "0", "3500", "0")
	nodes[5].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "0")
}

// increase exchange rate and decrease commissions during payment using 2 paths
func Test9ExchangePaymentOneNodeSeveralPaths(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentOneNodeSeveralPathsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "3500")

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 5},
	})

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "6", -1, nil, nil, vtcp.StatusOK)

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "2000", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-2497")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "10000", "-3334")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[3], testconfig.ExchangeEquivalent, "0", "2000", "0")
	nodes[3].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "4000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[3], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[4], testconfig.ExchangeEquivalent, "0", "5000", "-842")
	nodes[4].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[4].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-842")
	nodes[1].CheckSettlementLineForSync(t, nodes[4], testconfig.ExchangeEquivalent)

	nodes[2].CheckActiveSettlementLine(t, nodes[5], testconfig.Equivalent, "0", "3500", "-2000")
	nodes[5].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "1500")
}

// increase exchange rate and decrease commissions during payment using 3 paths
func Test9aExchangePaymentOneNodeSeveralPaths(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentOneNodeSeveralPathsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -1, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "3500")

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 5},
	})

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "6", -1, nil, nil, vtcp.StatusOK)

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[5], testconfig.Equivalent, "3000", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-2497")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "10000", "-5001")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[3], testconfig.ExchangeEquivalent, "0", "2000", "-4")
	nodes[3].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "4000", "-4")
	nodes[1].CheckSettlementLineForSync(t, nodes[3], testconfig.ExchangeEquivalent)
	nodes[0].CheckActiveSettlementLine(t, nodes[4], testconfig.ExchangeEquivalent, "0", "5000", "-2505")
	nodes[4].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[4].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-2505")
	nodes[1].CheckSettlementLineForSync(t, nodes[4], testconfig.ExchangeEquivalent)

	nodes[2].CheckActiveSettlementLine(t, nodes[5], testconfig.Equivalent, "0", "3500", "-3000")
	nodes[5].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)

	// 499 is wrong, it should be 500
	nodes[0].CheckExchangeMaxFlow(t, nodes[5], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "499")
}
