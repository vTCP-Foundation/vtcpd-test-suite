package testconfig

import (
	"github.com/vTCP-Foundation/vtcpd-test-suite/internal/conf"
	vtcp "github.com/vTCP-Foundation/vtcpd-test-suite/pkg/testsuite"
)

var (
	GSettings                                                                     vtcp.ClusterSettings
	StaticContainerIPPartForOpenChannelTest                                       = "172.18.1."
	StaticContainerIPPartForDirectPaymentTwoNodes                                 = "172.18.2."
	StaticContainerIPPartForDirectPaymentSevenNodes                               = "172.18.3."
	StaticContainerIPPartForSeveralPathPaymentCoordinatorBranchingTest            = "172.18.4."
	StaticContainerIPPartForSeveralPathPaymentIntermidiateBranchingTest           = "172.18.5."
	StaticContainerIPPartForSeveralPathPaymentReceiverBranchingTest               = "172.18.6."
	StaticContainerIPPartForSeveralPaymentsAtTheSameTimeTest                      = "172.18.7."
	StaticContainerIPPartForSeveralPathsPaymentSearchingNewPathTest               = "172.18.8."
	StaticContainerIPPartForPaymentTimeoutsPartOneTest                            = "172.18.9."
	StaticContainerIPPartForPaymentTimeoutsPartTwoTest                            = "172.18.10."
	StaticContainerIPPartForPaymentTimeoutsPartThreeTest                          = "172.18.11."
	StaticContainerIPPartForPaymentHopsCountTest                                  = "172.18.12."
	StaticContainerIPPartForOpenSettlementLineTest                                = "172.18.13."
	StaticContainerIPPartForOpenSettlementLineBadInternetTest                     = "172.18.14."
	StaticContainerIPPartForSetSettlementLineTest                                 = "172.18.15."
	StaticContainerIPPartForSetSettlementLineBadInternetTest                      = "172.18.16."
	StaticContainerIPPartForSettlementLineCloseMaxNegativeBalanceTest             = "172.18.17."
	StaticContainerIPPartForKeysSharingInitSettlementLineTest                     = "172.18.18."
	StaticContainerIPPartForKeysSharingNextSettlementLineTest                     = "172.18.19."
	StaticContainerIPPartForKeysSharingNextCntPaymentsAuditRuleSettlementLineTest = "172.18.20."
	StaticContainerIPPartForSettlementLineKeysSharingBadInternetTest              = "172.18.21."
	StaticContainerIPPartForSettlementLineAuditRuleOverflowedTest                 = "172.18.22."
	StaticContainerIPPartForSettlementLineArchivedTest                            = "172.18.23."
	StaticContainerIPPartForSettlementLineAuditVsPaymentTest                      = "172.18.24."
	StaticContainerIPPartForMaxFlowBatchTest                                      = "172.18.25."
	StaticContainerIPPartForMaxFlowTest                                           = "172.18.26."
	StaticContainerIPPartForHistoryPaymentsTest                                   = "172.18.27."
	Equivalent                                                                    = "1"
	OperationsLogPathInContainer                                                  = "/vtcp/vtcpd/operations.log"
	ConfigFilePathInContainer                                                     = "/vtcp/vtcpd/conf.json"
)

func init() {
	configFromInternalConf := conf.GetConfig()
	GSettings = vtcp.ClusterSettings{
		NodeImageName: configFromInternalConf.NodeImageName,
		NetworkName:   configFromInternalConf.NetworkName,
		SudoPassword:  configFromInternalConf.SudoPassword,
	}
}
