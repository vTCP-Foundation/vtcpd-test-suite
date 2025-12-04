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
	// exchangePaymentFiveNodesWithCommissionsNextNodeIndex is used to assign unique IP addresses to nodes across different test functions in this file.
	exchangePaymentFiveNodesWithCommissionsNextNodeIndex = 1
)

func getNextIPForExchangePaymentFiveNodesWithCommissionsTest() string {
	ip := fmt.Sprintf("%s%d", testconfig.StaticContainerIPPartForExchangePaymentFiveNodesWithCommissions, exchangePaymentFiveNodesWithCommissionsNextNodeIndex)
	exchangePaymentFiveNodesWithCommissionsNextNodeIndex++
	return ip
}

// Helper to create and run nodes for a test
func setupNodesForExchangePaymentFiveNodesWithCommissionsTest(t *testing.T) ([]*vtcp.Node, *vtcp.Cluster) {
	nodes := make([]*vtcp.Node, 5)
	for i := range 5 {
		nodes[i] = vtcp.NewNode(t, getNextIPForExchangePaymentFiveNodesWithCommissionsTest(), fmt.Sprintf("node%d", i+1))
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
	nodes[3].OpenChannelAndCheck(t, nodes[2])
	nodes[4].OpenChannelAndCheck(t, nodes[3])

	nodes[1].CreateAndSetSettlementLineAndCheck(t, nodes[0], testconfig.ExchangeEquivalent, "3000")
	nodes[2].CreateAndSetSettlementLineAndCheck(t, nodes[1], testconfig.ExchangeEquivalent, "5000")

	nodes[3].CreateAndSetSettlementLineAndCheck(t, nodes[2], testconfig.Equivalent, "250")
	nodes[4].CreateAndSetSettlementLineAndCheck(t, nodes[3], testconfig.Equivalent, "500")

	return nodes, cluster
}

func Test1ExchangePaymentFiveNodesWithCommissions(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentFiveNodesWithCommissionsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -2, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})
	nodes[3].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 3},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[4], testconfig.Equivalent, "100", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-2070")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "5000", "-2060")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[2].CheckActiveSettlementLine(t, nodes[3], testconfig.Equivalent, "0", "250", "-103")
	nodes[3].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[4], testconfig.Equivalent, "0", "500", "-100")
	nodes[4].CheckSettlementLineForSync(t, nodes[3], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "43")
}

func Test2ExchangePaymentFiveNodesWithCommissions(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentFiveNodesWithCommissionsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", 1, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})
	nodes[3].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 3},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "247")

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[4], testconfig.Equivalent, "147", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-13")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "5000", "-3")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[2].CheckActiveSettlementLine(t, nodes[3], testconfig.Equivalent, "0", "250", "-150")
	nodes[3].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[4], testconfig.Equivalent, "0", "500", "-147")
	nodes[4].CheckSettlementLineForSync(t, nodes[3], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "97")
}

// increase exchange rate during payment
func Test3ExchangePaymentFiveNodesWithCommissionsChangeExchangeRate(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentFiveNodesWithCommissionsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -2, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})
	nodes[3].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 3},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "146")

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "6", -2, nil, nil, vtcp.StatusOK)

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[4], testconfig.Equivalent, "100", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-1727")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "5000", "-1717")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[2].CheckActiveSettlementLine(t, nodes[3], testconfig.Equivalent, "0", "250", "-103")
	nodes[3].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[4], testconfig.Equivalent, "0", "500", "-100")
	nodes[4].CheckSettlementLineForSync(t, nodes[3], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "72")
}

// decrease exchange rate during payment
func Test3aExchangePaymentFiveNodesWithCommissionsChangeExchangeRate(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentFiveNodesWithCommissionsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -2, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})
	nodes[3].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 3},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "146")

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "4", -2, nil, nil, vtcp.StatusOK)

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[4], testconfig.Equivalent, "100", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-2585")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "5000", "-2575")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[2].CheckActiveSettlementLine(t, nodes[3], testconfig.Equivalent, "0", "250", "-103")
	nodes[3].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[4], testconfig.Equivalent, "0", "500", "-100")
	nodes[4].CheckSettlementLineForSync(t, nodes[3], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "13")
}

// disappear exchange rate during payment
func Test4ExchangePaymentFiveNodesWithCommissionsDisappearExchangeRate(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentFiveNodesWithCommissionsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -2, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})
	nodes[3].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 3},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "146")

	nodes[2].ClearExchangeRates(t)

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[4], testconfig.Equivalent, "100", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusInsufficientFunds)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "0")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "5000", "0")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[2].CheckActiveSettlementLine(t, nodes[3], testconfig.Equivalent, "0", "250", "0")
	nodes[3].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[4], testconfig.Equivalent, "0", "500", "0")
	nodes[4].CheckSettlementLineForSync(t, nodes[3], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "0")
}

// increase first commission during payment
func Test5ExchangePaymentFiveNodesWithCommissionsChangeFirstCommission(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentFiveNodesWithCommissionsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -2, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})
	nodes[3].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 3},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "146")

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 20},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[4], testconfig.Equivalent, "100", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-2080")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "5000", "-2060")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[2].CheckActiveSettlementLine(t, nodes[3], testconfig.Equivalent, "0", "250", "-103")
	nodes[3].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[4], testconfig.Equivalent, "0", "500", "-100")
	nodes[4].CheckSettlementLineForSync(t, nodes[3], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "42")
}

// decrease first commission during payment
func Test5aExchangePaymentFiveNodesWithCommissionsChangeFirstCommission(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentFiveNodesWithCommissionsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -2, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})
	nodes[3].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 3},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "146")

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 6},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[4], testconfig.Equivalent, "100", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-2066")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "5000", "-2060")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[2].CheckActiveSettlementLine(t, nodes[3], testconfig.Equivalent, "0", "250", "-103")
	nodes[3].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[4], testconfig.Equivalent, "0", "500", "-100")
	nodes[4].CheckSettlementLineForSync(t, nodes[3], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "43")
}

// increase second commission during payment
func Test6ExchangePaymentFiveNodesWithCommissionsChangeSecondCommission(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentFiveNodesWithCommissionsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -2, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})
	nodes[3].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 3},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "146")

	nodes[3].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 5},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[4], testconfig.Equivalent, "100", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-2110")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "5000", "-2100")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[2].CheckActiveSettlementLine(t, nodes[3], testconfig.Equivalent, "0", "250", "-105")
	nodes[3].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[4], testconfig.Equivalent, "0", "500", "-100")
	nodes[4].CheckSettlementLineForSync(t, nodes[3], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "39")
}

// decrease second commission during payment
func Test6aExchangePaymentFiveNodesWithCommissionsChangeSecondCommission(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentFiveNodesWithCommissionsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -2, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})
	nodes[3].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 3},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "146")

	nodes[3].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 1},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[4], testconfig.Equivalent, "100", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-2030")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "5000", "-2020")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[2].CheckActiveSettlementLine(t, nodes[3], testconfig.Equivalent, "0", "250", "-101")
	nodes[3].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[4], testconfig.Equivalent, "0", "500", "-100")
	nodes[4].CheckSettlementLineForSync(t, nodes[3], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "47")
}

// change first and second commission during payment
func Test7ExchangePaymentFiveNodesWithCommissionsChangeBothCommissions(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentFiveNodesWithCommissionsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -2, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})
	nodes[3].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 3},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "146")

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 20},
	})

	nodes[3].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 5},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[4], testconfig.Equivalent, "100", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-2120")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "5000", "-2100")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[2].CheckActiveSettlementLine(t, nodes[3], testconfig.Equivalent, "0", "250", "-105")
	nodes[3].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[4], testconfig.Equivalent, "0", "500", "-100")
	nodes[4].CheckSettlementLineForSync(t, nodes[3], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "38")
}

// appear commission during payment
func Test8ExchangePaymentFiveNodesWithCommissionsAppearCommission(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentFiveNodesWithCommissionsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -2, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "149")

	nodes[3].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 5},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[4], testconfig.Equivalent, "100", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-2110")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "5000", "-2100")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[2].CheckActiveSettlementLine(t, nodes[3], testconfig.Equivalent, "0", "250", "-105")
	nodes[3].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[4], testconfig.Equivalent, "0", "500", "-100")
	nodes[4].CheckSettlementLineForSync(t, nodes[3], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "39")
}

// disappear commission during payment
func Test9ExchangePaymentFiveNodesWithCommissionsDisappearCommission(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentFiveNodesWithCommissionsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -2, nil, nil, vtcp.StatusOK)

	nodes[1].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.ExchangeEquivalent, Amount: 10},
	})
	nodes[3].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 3},
	})

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "146")

	nodes[3].SetCommissions([]vtcp.CommissionPair{
		{Equivalent: testconfig.Equivalent, Amount: 0},
	})

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[4], testconfig.Equivalent, "100", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-2010")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "5000", "-2000")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[2].CheckActiveSettlementLine(t, nodes[3], testconfig.Equivalent, "0", "250", "-100")
	nodes[3].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[4], testconfig.Equivalent, "0", "500", "-100")
	nodes[4].CheckSettlementLineForSync(t, nodes[3], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "49")
}

func Test10ExchangePaymentFiveNodesWithCommissionsChangeExchangeRate(t *testing.T) {
	nodes, _ := setupNodesForExchangePaymentFiveNodesWithCommissionsTest(t)

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "5", -2, nil, nil, vtcp.StatusOK)

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "150")

	nodes[2].SetExchangeRateNative(t, testconfig.ExchangeEquivalent, testconfig.Equivalent, "6", -2, nil, nil, vtcp.StatusOK)

	nodes[0].CreateExchangeTransactionCheckStatus(t, nodes[4], testconfig.Equivalent, "100", testconfig.ExchangeEquivalent, vtcp.NoMaxAllowablePaymentAmount, vtcp.StatusOK)

	time.Sleep(2 * time.Second)
	nodes[0].CheckActiveSettlementLine(t, nodes[1], testconfig.ExchangeEquivalent, "0", "3000", "-1667")
	nodes[1].CheckSettlementLineForSync(t, nodes[0], testconfig.ExchangeEquivalent)
	nodes[1].CheckActiveSettlementLine(t, nodes[2], testconfig.ExchangeEquivalent, "0", "5000", "-1667")
	nodes[2].CheckSettlementLineForSync(t, nodes[1], testconfig.ExchangeEquivalent)
	nodes[2].CheckActiveSettlementLine(t, nodes[3], testconfig.Equivalent, "0", "250", "-100")
	nodes[3].CheckSettlementLineForSync(t, nodes[2], testconfig.Equivalent)
	nodes[3].CheckActiveSettlementLine(t, nodes[4], testconfig.Equivalent, "0", "500", "-100")
	nodes[4].CheckSettlementLineForSync(t, nodes[3], testconfig.Equivalent)

	nodes[0].CheckExchangeMaxFlow(t, nodes[4], testconfig.Equivalent, []string{testconfig.ExchangeEquivalent}, "79")
}
