package tests

import (
	"context"
	"fmt"
	"sync"
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
	nodeA := vtcp.NewNode(t, "172.18.0.2")
	nodeB := vtcp.NewNode(t, "172.18.0.3")

	ctx := context.Background()
	cluster, err := vtcp.NewCluster(ctx, t, clusterSettings)
	if err != nil {
		t.Fatalf("failed to create cluster: %v", err)
	}

	wg := sync.WaitGroup{}
	{
		err = cluster.RunNode(ctx, t, &wg, nodeA)
		if err != nil {
			t.Fatalf("failed to run nodeA: %v", err)
		}

		err = cluster.RunNode(ctx, t, &wg, nodeB)
		if err != nil {
			t.Fatalf("failed to run nodeB: %v", err)
		}
	}
	wg.Wait()

	println(nodeA.IPAddress)
	println(nodeB.IPAddress)

	println(nodeA.ContainerID)
	println(nodeB.ContainerID)

	time.Sleep(2 * time.Second)

	err = nodeA.OpenChannel(nodeB)
	if err != nil {
		println(err.Error())
	} else {
		println("Channel opened successfully")
	}

	channelInfo, err := nodeA.GetChannelInfo(nodeB)
	if err != nil {
		println(err.Error())
	} else {
		println(fmt.Sprintf("Channel on A side: id:%s, isConfirmed:%s", channelInfo.ChannelID, channelInfo.ChannelConfirmed))
	}

	channelInfo, err = nodeB.GetChannelInfo(nodeA)
	if err != nil {
		println(err.Error())
	} else {
		println(fmt.Sprintf("Channel on B side: id:%s, isConfirmed:%s", channelInfo.ChannelID, channelInfo.ChannelConfirmed))
	}

	equivalent := "1002"

	err = nodeA.CreateSettlementLine(nodeB, equivalent, "1000000000000000000")
	if err != nil {
		println(err.Error())
	} else {
		println(fmt.Sprintf("Settlement line in %s equivalent created successfully", equivalent))
	}

	settlmentLineInfo, err := nodeA.GetSettlementsLineInfoByAddress(nodeB, equivalent)
	if err != nil {
		println(err.Error())
	} else {
		println(fmt.Sprintf("Settlement line info on A side: %+v", settlmentLineInfo))
	}

	settlmentLineInfo, err = nodeB.GetSettlementsLineInfoByAddress(nodeA, equivalent)
	if err != nil {
		println(err.Error())
	} else {
		println(fmt.Sprintf("Settlement line info on B side: %+v", settlmentLineInfo))
	}

	transactionUUID, err := nodeB.CreateTransaction(nodeA, equivalent, "10000")
	if err != nil {
		println(err.Error())
	} else {
		println(fmt.Sprintf("Transaction created successfully: %s", transactionUUID))
	}

	settlmentLineInfo, err = nodeA.GetSettlementsLineInfoByAddress(nodeB, equivalent)
	if err != nil {
		println(err.Error())
	} else {
		println(fmt.Sprintf("Settlement line info on A side: %+v", settlmentLineInfo))
	}

	settlmentLineInfo, err = nodeB.GetSettlementsLineInfoByAddress(nodeA, equivalent)
	if err != nil {
		println(err.Error())
	} else {
		println(fmt.Sprintf("Settlement line info on B side: %+v", settlmentLineInfo))
	}
}
