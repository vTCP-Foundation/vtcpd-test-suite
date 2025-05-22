package testconfig

import (
	"github.com/vTCP-Foundation/vtcpd-test-suite/internal/conf"
	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
)

var (
	GSettings                                                 vtcp.ClusterSettings
	StaticContainerIPPartForDirectPaymentTwoNodes             = "172.18.1."
	StaticContainerIPPartForDirectPaymentSevenNodes           = "172.18.2."
	StaticContainerIPPartForOpenSettlementLineTest            = "172.18.3."
	StaticContainerIPPartForSetSettlementLineTest             = "172.18.4."
	StaticContainerIPPartForKeysSharingInitSettlementLineTest = "172.18.5."
	StaticContainerIPPartForMaxFlowBatchTest                  = "172.18.6."
	StaticContainerIPPartForMaxFlowTest                       = "172.18.7."
	Equivalent                                                = "1"
	OperationsLogPathInContainer                              = "/vtcp/vtcpd/operations.log"
)

func init() {
	configFromInternalConf := conf.GetConfig()
	GSettings = vtcp.ClusterSettings{
		NodeImageName: configFromInternalConf.NodeImageName,
		NetworkName:   configFromInternalConf.NetworkName,
	}
}
