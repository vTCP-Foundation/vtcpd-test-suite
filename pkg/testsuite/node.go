package testsuite

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultNodePort    = 2000
	DefaultCLIPort     = 3000
	DefaultCLIPortTest = 3001

	ChannelConfirmed = "1"

	SettlementLineStateInit   = "1"
	SettlementLineStateActive = "2"
	SettlementLineStateClosed = "4"

	SettlementLineKeysPresent = "1"
	SettlementLineKeysAbsent  = "0"

	StatusOK                 = 200
	StatusProtocolError      = 401
	StatusNoConsensusError   = 409
	StatusInsufficientFunds  = 412
	StatusNoPaymentRoutes    = 462
	StatusNotImplemented     = 501 // HTTP 501 Not Implemented
	StatusServiceUnavailable = 503 // HTTP 503 Service Unavailable

	PaymentObservingStateNoInfo   = "0"
	PaymentObservingStateClaimed  = "1"
	PaymentObservingStateRejected = "5"

	WaitingParticipantsVotesSec = 15

	NodePaymentRecoveryAttempts      = 3
	NodePaymentRecoveryTimePeriodSec = 10

	ObservingCntBlocksForClaiming = 20
	ObservingCntSecondsPerBlock   = 20

	// Test Flags for SetTestingSLFlag (derived from Python test_trustlines_open.py)
	TrustLineDebugFlagRejectNewRequestRace           uint32 = 4       // Corresponds to Python's usage of 4 in set_TL_debug_flag for rejecting new requests during trustline operations. The specific meaning (request vs audit) is often contextual to the message type it's paired with.
	TrustLineDebugFlagRejectNewAuditRace             uint32 = 4       // Also uses 4, context via message type. Python used same numeric flag for different semantic meanings.
	TestFlagExceptionOnInitTAModifyingStage          uint32 = 2048    // VTCPD_DEBUG_FLAG_SETTLEMENT_LINE_INITIATOR_TA_MODIFYING_STAGE_EXCEPTION
	TestFlagExceptionOnInitTAResponseProcessingStage uint32 = 4096    // VTCPD_DEBUG_FLAG_SETTLEMENT_LINE_INITIATOR_TA_RESPONSE_PROCESSING_STAGE_EXCEPTION
	TestFlagExceptionOnInitTAResumingStage           uint32 = 8192    // VTCPD_DEBUG_FLAG_SETTLEMENT_LINE_INITIATOR_TA_RESUMING_STAGE_EXCEPTION
	TestFlagExceptionOnContractorTAStage             uint32 = 16384   // VTCPD_DEBUG_FLAG_SETTLEMENT_LINE_CONTRACTOR_TA_STAGE_EXCEPTION
	TestFlagTerminateOnInitTAModifyingStage          uint32 = 2097152 // VTCPD_DEBUG_FLAG_SETTLEMENT_LINE_INITIATOR_TA_MODIFYING_STAGE_TERMINATE
	TestFlagTerminateOnInitTAResponseProcessingStage uint32 = 4194304 // VTCPD_DEBUG_FLAG_SETTLEMENT_LINE_INITIATOR_TA_RESPONSE_PROCESSING_STAGE_TERMINATE
	TestFlagTerminateOnInitTAResumingStage           uint32 = 8388608 // VTCPD_DEBUG_FLAG_SETTLEMENT_LINE_INITIATOR_TA_RESUMING_STAGE_TERMINATE

	// Message types/codes for SetTestingSLFlag firstParam (derived from Python)
	// These typically correspond to transaction types or message identifiers used in debug flags.
	SettlementLineStandardMessageType   = "101" // Corresponds to self.trustLineMessage in Python, also self.targetTransactionType
	SettlementLineResponseMessageType   = "102" // Corresponds to self.trustLineResponseMessage
	SettlementLineSourceTransactionType = "100" // Corresponds to self.sourceTransactionType
	// TargetTransactionType from Python is "101", which is SettlementLineStandardMessageType

	// New constants for SetSettlementLine tests
	SettlementLineSetMessageType              = "106" // Python: self.trustLineMessage for set operations
	SettlementLineSetAuditMessageType         = "107" // Python: self.auditResponseMessage for set operations
	SettlementLineSetInitiatorTransactionType = "102" // Python: self.sourceTransactionType for set (Note: same as SettlementLineResponseMessageType)
	SettlementLineSetTargetTransactionType    = "107" // Python: self.targetTransactionType for set (Note: same as SettlementLineSetAuditMessageType)

	// New constants for Settlement Line Keys Sharing Init tests
	SettlementLinePublicKeyInitMessageType     = "103" // Python: self.publicKeyInitMessage
	SettlementLinePublicKeyMessageType         = "104" // Python: self.publicKeyMessage (also sourceTransactionType for these tests)
	SettlementLinePublicKeyResponseMessageType = "105" // Python: self.publicKeyResponseMessage (also targetTransactionType for these tests)

	// Default values inspired by Python test suite
	DefaultKeysCount                    = 20
	DefaultWaitingResponseTime          = 20 * time.Second
	DefaultMaxMessageSendingAttemptsStr = "3" // As string for SetTestingSLFlag
	DefaultMaxMessageSendingAttemptsInt = 3

	// Log Messages
	LogMessageRecoveringLogMessage = "runVotesRecoveryParentStage"
	DefaultOperationsLogPath       = "/vtcp/vtcpd/logs/operations.log"
)

// Testing flags based on Python test suite debug flags
const (
	// From test_transaction_direct_payment_two_nodes.py and their Go equivalents
	FlagForbidSendInitMessage                        uint32 = 4        // Python: flag_forbid_send_init_message
	FlagForbidSendMessageToCoordinatorReservation    uint32 = 8        // Python: flag_forbid_send_message_to_coordinator_on_reservation
	FlagForbidSendRequestToIntermediateReservation   uint32 = 16       // Python: flag_forbid_send_request_to_intermediate_node_on_reservation
	FlagForbidSendResponseToIntemediateOnReservation uint32 = 32       // Python: flag_forbid_send_response_to_intermediate_node_on_reservation
	FlagForbidSendMessageFinalPathConfig             uint32 = 64       // Python: flag_forbid_send_message_with_final_path_configuration
	FlagForbidSendMessageFinalAmountClarification    uint32 = 128      // Python: flag_forbid_send_message_on_final_amount_clarification
	FlagForbidSendMessageVoteStage                   uint32 = 256      // Python: flag_forbid_send_message_on_vote_stage
	FlagForbidSendMessageVoteConsistency             uint32 = 512      // Python: flag_forbid_send_message_on_vote_consistency
	FlagForbidSendMessageRecoveryStage               uint32 = 1024     // Python: flag_forbid_send_message_on_recovery_stage
	FlagThrowExceptionPreviousNeighborRequest        uint32 = 2048     // Python: flag_throw_exception_on_previous_neighbor_request
	FlagThrowExceptionCoordinatorRequest             uint32 = 4096     // Python: flag_throw_exception_on_coordinator_request_processing
	FlagThrowExceptionNextNeighborResponse           uint32 = 8192     // Python: flag_throw_exception_on_next_neighbor_response_processing
	FlagThrowExceptionVote                           uint32 = 16384    // Python: flag_throw_exception_on_vote
	FlagThrowExceptionVoteConsistency                uint32 = 32768    // Python: flag_throw_exception_on_vote_consistency
	FlagThrowExceptionCoordinatorAfterApprove        uint32 = 65536    // Python: flag_throw_exception_on_coordinator_after_approve_before_send_message
	FlagTerminateProcessPreviousNeighborRequest      uint32 = 2097152  // Python: flag_terminate_process_on_previous_neighbor_request
	FlagTerminateProcessCoordinatorRequest           uint32 = 4194304  // Python: flag_terminate_process_on_coordinator_request_processing
	FlagTerminateProcessNextNeighborResponse         uint32 = 8388608  // Python: flag_terminate_process_on_next_neighbor_response_processing
	FlagTerminateProcessVote                         uint32 = 16777216 // Python: flag_terminate_process_on_vote
	FlagTerminateProcessVoteConsistency              uint32 = 33554432 // Python: flag_terminate_process_on_vote_consistency
	FlagTerminateProcessCoordinatorAfterApprove      uint32 = 67108864 // Python: flag_terminate_process_on_coordinator_after_approve_before_send_message
)

type Node struct {
	ID          string
	Host        string
	NodePort    uint16
	CLIPort     uint16
	CLIPortTest uint16
	IPAddress   string
	ContainerID string
	Alias       string
	Env         []string
}

type ChannelInitResponseData struct {
	ChannelID string `json:"channel_id"`
	CryptoKey string `json:"crypto_key"`
}

// ChannelInfo holds information about a channel between nodes.
type ChannelInfo struct {
	ChannelID        string `json:"channel_id"`
	ChannelConfirmed string `json:"channel_confirmed"`
}

// SettlementLineInfo holds information about a settlement line.
type SettlementLineInfo struct {
	ID                    string `json:"id"`
	ContractorAddress     string `json:"contractor"`
	State                 string `json:"state"`
	OwnKeysPresent        string `json:"own_keys_present"`
	ContractorKeysPresent string `json:"contractor_keys_present"`
	AuditNumber           string `json:"audit_number"`
	MaxNegativeBalance    string `json:"max_negative_balance"`
	MaxPositiveBalance    string `json:"max_positive_balance"`
	Balance               string `json:"balance"`
}

type SettlementLineInfoList struct {
	Count   int                  `json:"count"`
	Records []SettlementLineInfo `json:"settlement_lines"`
}

type MaxFlowItemInfo struct {
	AddressType       string `json:"address_type"`
	ContractorAddress string `json:"contractor_address"`
	MaxAmount         string `json:"max_amount"`
}

type MaxFlowInfo struct {
	Count   int               `json:"count"`
	Records []MaxFlowItemInfo `json:"records"`
}

// MaxFlowBatchResult holds the contractor address and its corresponding max flow amount.
type MaxFlowBatchResult struct {
	ContractorAddress string
	MaxAmount         string
}

// MaxFlowBatchCheck is a helper struct for CheckMaxFlowBatch
type MaxFlowBatchCheck struct {
	Node            *Node
	ExpectedMaxFlow string
}

func NewNode(t *testing.T, ipAddress string, alias string) *Node {
	return &Node{
		ID:          uuid.New().String(),
		Host:        "0.0.0.0",
		NodePort:    DefaultNodePort,
		CLIPort:     DefaultCLIPort,
		CLIPortTest: DefaultCLIPortTest,
		IPAddress:   ipAddress,
		ContainerID: "", // Must be set on container creation.
		Alias:       alias,
		Env: []string{
			fmt.Sprintf("VTCPD_LISTEN_ADDRESS=%s", ipAddress),
			fmt.Sprintf("VTCPD_LISTEN_PORT=%d", DefaultNodePort),
			"VTCPD_EQUIVALENTS_REGISTRY=eth",
			"VTCPD_MAX_HOPS=5",
			"CLI_LISTEN_ADDRESS=0.0.0.0",
			fmt.Sprintf("CLI_LISTEN_PORT=%d", DefaultCLIPort),
			fmt.Sprintf("CLI_LISTEN_PORT_TESTING=%d", DefaultCLIPortTest),
		},
	}
}

func (n *Node) GetIpAndPort() string {
	return fmt.Sprintf("%s:%d", n.IPAddress, n.NodePort)
}

func (n *Node) GetIPAddressForRequests() string {
	return fmt.Sprintf("12-%s:%d", n.IPAddress, n.NodePort)
}

// OpenChannel opens a channel between this node and the target node.
// It uses the init-channel functionality to establish the connection.
// Returns an error if the channel initialization fails, otherwise returns nil.
func (n *Node) OpenChannel(t *testing.T, targetNode *Node) {
	// Step 1: This node initiates the channel
	// Prepare the request body with the target node's address
	// Using IPV4 type code 12
	initURL := fmt.Sprintf("http://%s:%d/api/v1/node/contractors/init-channel/?contractor_address=%s",
		n.IPAddress, n.CLIPort, targetNode.GetIPAddressForRequests())

	// Send the request to initialize the channel
	resp, err := http.Post(initURL, "application/json", nil)
	if err != nil {
		t.Fatalf("failed to send init-channel request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("init-channel request failed with status: %d", resp.StatusCode)
	}

	// Parse the response to get channel_id and crypto_key
	var initResponse struct {
		Data ChannelInitResponseData `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&initResponse); err != nil {
		t.Fatalf("failed to decode init-channel response: %v", err)
	}

	// Step 2: Target node completes the channel initialization
	// Prepare the request body with this node's address, channel_id, and crypto_key
	targetURL := fmt.Sprintf("http://%s:%d/api/v1/node/contractors/init-channel/?contractor_address=%s&contractor_id=%s&crypto_key=%s",
		targetNode.IPAddress, targetNode.CLIPort, n.GetIPAddressForRequests(), initResponse.Data.ChannelID, initResponse.Data.CryptoKey)

	// Send the request to complete channel initialization
	targetResp, err := http.Post(targetURL, "application/json", nil)
	if err != nil {
		t.Fatalf("failed to send target init-channel request: %v", err)
	}
	defer targetResp.Body.Close()

	if targetResp.StatusCode != http.StatusOK {
		t.Fatalf("target init-channel request failed with status: %d", targetResp.StatusCode)
	}
}

// getChannelInfo queries the channel-by-address endpoint to get channel info with another node.
// Returns channel info or an error.
func (n *Node) GetChannelInfo(targetNode *Node) (*ChannelInfo, error) {
	url := fmt.Sprintf("http://%s:%d/api/v1/node/channel-by-address/?contractor_address=%s",
		n.IPAddress, n.CLIPort, targetNode.GetIPAddressForRequests())

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send channel-by-address request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("channel-by-address request failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Data ChannelInfo `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode channel-by-address response: %w", err)
	}
	return &result.Data, nil
}

func (n *Node) OpenChannelAndCheck(t *testing.T, targetNode *Node) {
	n.OpenChannel(t, targetNode)
	channelInfo, err := n.GetChannelInfo(targetNode)
	if err != nil {
		t.Fatalf("failed to get channel info: %v", err)
	}
	if channelInfo.ChannelConfirmed != ChannelConfirmed {
		t.Fatalf("channel is not confirmed")
	}
	channelInfo, err = targetNode.GetChannelInfo(n)
	if err != nil {
		t.Fatalf("failed to get channel info: %v", err)
	}
	if channelInfo.ChannelConfirmed != ChannelConfirmed {
		t.Fatalf("channel is not confirmed")
	}
}

// CreateAndSetSettlementLine creates a settlement line with another node.
// It first calls init-settlement-line and sets the max positive balance.
func (n *Node) CreateAndSetSettlementLine(t *testing.T, targetNode *Node, equivalent string, amount string) {
	n.CreateSettlementLine(t, targetNode, equivalent)
	time.Sleep(2 * time.Second)
	n.SetSettlementLine(t, targetNode, equivalent, amount)
}

// CreateAndSetSettlementLine creates a settlement line with another node.
// It first gets the contractor_id using getChannelInfo
// Returns error if any step fails.
func (n *Node) CreateSettlementLine(t *testing.T, targetNode *Node, equivalent string) {
	// Step 1: Get contractor_id (channel_id) for the target node
	channelInfo, err := n.GetChannelInfo(targetNode)
	if err != nil {
		t.Fatalf("failed to get channel info: %v", err)
	}
	contractorID := channelInfo.ChannelID

	// Step 2: Call init-settlement-line
	initURL := fmt.Sprintf("http://%s:%d/api/v1/node/contractors/%s/init-settlement-line/%s/",
		n.IPAddress, n.CLIPort, contractorID, equivalent)

	initResp, err := http.Post(initURL, "application/json", nil)
	if err != nil {
		t.Fatalf("failed to send init-settlement-line request: %v", err)
	}
	defer initResp.Body.Close()
	if initResp.StatusCode != http.StatusOK {
		t.Fatalf("init-settlement-line request failed with status: %d", initResp.StatusCode)
	}
}

func (n *Node) SetSettlementLine(t *testing.T, targetNode *Node, equivalent string, amount string) {
	// Step 1: Get contractor_id (channel_id) for the target node
	channelInfo, err := n.GetChannelInfo(targetNode)
	if err != nil {
		t.Fatalf("failed to get channel info: %v", err)
	}
	contractorID := channelInfo.ChannelID

	// Step 2: Set max positive balance (PUT)
	setURL := fmt.Sprintf("http://%s:%d/api/v1/node/contractors/%s/settlement-lines/%s/?amount=%s",
		n.IPAddress, n.CLIPort, contractorID, equivalent, amount)
	request, err := http.NewRequest(http.MethodPut, setURL, nil)
	if err != nil {
		t.Fatalf("failed to create set-settlement-line request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	setResp, err := client.Do(request)
	if err != nil {
		t.Fatalf("failed to send set-settlement-line request: %v", err)
	}
	defer setResp.Body.Close()
	if setResp.StatusCode != http.StatusOK {
		t.Fatalf("set-settlement-line request failed with status: %d", setResp.StatusCode)
	}
}

// GetSettlementsLineInfoByAddress fetches settlement line information using the target node's address.
func (n *Node) GetSettlementsLineInfoByAddress(targetNode *Node, equivalent string) (*SettlementLineInfo, int, error) {
	url := fmt.Sprintf("http://%s:%d/api/v1/node/contractors/settlement-line-by-address/%s/?contractor_address=%s",
		n.IPAddress, n.CLIPort, equivalent, targetNode.GetIPAddressForRequests())

	resp, err := http.Get(url)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to send settlement-line-by-address request: %w", err)
	}
	defer resp.Body.Close()

	// Return status code for all responses
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, resp.StatusCode, fmt.Errorf("settlement-line-by-address request failed with status: %d. Body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Data struct {
			SettlementLine SettlementLineInfo `json:"settlement_line"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to decode settlement-line-by-address response: %w", err)
	}

	return &result.Data.SettlementLine, resp.StatusCode, nil
}

func (n *Node) CheckSettlementLine(t *testing.T, targetNode *Node, equivalent, expectedState, expectedMaxPositiveBalance, expectedMaxNegativeBalance,
	expectedBalance, expectedOwnKeysPresent, expectedContractorKeysPresent string, expectedStatusCode int) {
	settlementLineInfo, actualStatusCode, err := n.GetSettlementsLineInfoByAddress(targetNode, equivalent)

	if actualStatusCode != expectedStatusCode {
		t.Fatalf("Node %s to %s for equivalent %s: CheckSettlementLine status code mismatch. Expected: %d, Got: %d. Error (if any): %v", n.Alias, targetNode.Alias, equivalent, expectedStatusCode, actualStatusCode, err)
	}

	// If status code matches and it's not OK, we don't need to check other fields.
	// If an error occurred during GetSettlementsLineInfoByAddress (e.g. network issue) and actualStatusCode is 0 (or some other non-HTTP status),
	// the previous check would have caught it if expectedStatusCode was a valid HTTP status.
	// If expectedStatusCode was also 0 (or something indicating error), then this is fine.
	if expectedStatusCode != StatusOK {
		return
	}

	// If we expected OK, but an error still occurred (e.g. JSON decode issue after 200 OK)
	if err != nil {
		t.Fatalf("Node %s to %s for equivalent %s: failed to get settlement line info even with status OK. Error: %v", n.Alias, targetNode.Alias, equivalent, err)
	}

	if settlementLineInfo == nil {
		t.Fatalf("Node %s to %s for equivalent %s: settlementLineInfo is nil even with status OK and no error.", n.Alias, targetNode.Alias, equivalent)
		return
	}

	if settlementLineInfo.State != expectedState {
		t.Fatalf("Node %s to %s for equivalent %s: settlement line state mismatch. Expected: %s, Got: %s", n.Alias, targetNode.Alias, equivalent, expectedState, settlementLineInfo.State)
	}
	if settlementLineInfo.MaxPositiveBalance != expectedMaxPositiveBalance {
		t.Fatalf("Node %s to %s for equivalent %s: max positive balance mismatch. Expected: %s, Got: %s", n.Alias, targetNode.Alias, equivalent, expectedMaxPositiveBalance, settlementLineInfo.MaxPositiveBalance)
	}
	if settlementLineInfo.MaxNegativeBalance != expectedMaxNegativeBalance {
		t.Fatalf("Node %s to %s for equivalent %s: max negative balance mismatch. Expected: %s, Got: %s", n.Alias, targetNode.Alias, equivalent, expectedMaxNegativeBalance, settlementLineInfo.MaxNegativeBalance)
	}
	if settlementLineInfo.Balance != expectedBalance {
		t.Fatalf("Node %s to %s for equivalent %s: balance mismatch. Expected: %s, Got: %s", n.Alias, targetNode.Alias, equivalent, expectedBalance, settlementLineInfo.Balance)
	}
	if settlementLineInfo.OwnKeysPresent != expectedOwnKeysPresent {
		t.Fatalf("Node %s to %s for equivalent %s: own keys present mismatch. Expected: %s, Got: %s", n.Alias, targetNode.Alias, equivalent, expectedOwnKeysPresent, settlementLineInfo.OwnKeysPresent)
	}
	if settlementLineInfo.ContractorKeysPresent != expectedContractorKeysPresent {
		t.Fatalf("Node %s to %s for equivalent %s: contractor keys present mismatch. Expected: %s, Got: %s", n.Alias, targetNode.Alias, equivalent, expectedContractorKeysPresent, settlementLineInfo.ContractorKeysPresent)
	}
}

func (n *Node) CheckActiveSettlementLine(t *testing.T, targetNode *Node, equivalent, maxPositiveBalance, maxNegativeBalance, balance string) {
	n.CheckSettlementLine(t, targetNode, equivalent, SettlementLineStateActive, maxPositiveBalance, maxNegativeBalance, balance,
		SettlementLineKeysPresent, SettlementLineKeysPresent, StatusOK)
}

func (n *Node) CreateSettlementLineAndCheck(t *testing.T, targetNode *Node, equivalent string, amount string) {
	n.CreateAndSetSettlementLine(t, targetNode, equivalent, amount)

	time.Sleep(500 * time.Millisecond)

	n.CheckActiveSettlementLine(t, targetNode, equivalent, amount, "0", "0")
	targetNode.CheckActiveSettlementLine(t, n, equivalent, "0", amount, "0")
}

func (n *Node) CreateChannelAndSettlementLineAndCheck(t *testing.T, targetNode *Node, equivalent string, amount string) {
	n.OpenChannelAndCheck(t, targetNode)
	n.CreateSettlementLineAndCheck(t, targetNode, equivalent, amount)
}

func (n *Node) GetSettlementLines(equivalent string) ([]SettlementLineInfo, error) {
	url := fmt.Sprintf("http://%s:%d/api/v1/node/contractors/settlement-lines/%s/", n.IPAddress, n.CLIPort, equivalent)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send get settlement-lines request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get settlement-lines request failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Count           int                  `json:"count"`
			SettlementLines []SettlementLineInfo `json:"settlement_lines"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode get settlement-lines response: %w", err)
	}

	return result.Data.SettlementLines, nil
}

func (n *Node) CheckSettlementLineForSync(t *testing.T, targetNode *Node, equivalent string) {
	settlementLineInfo, _, err := n.GetSettlementsLineInfoByAddress(targetNode, equivalent)
	if err != nil {
		t.Fatalf("failed to get settlement line info: %v", err)
	}

	targetNodeSettlementLineInfo, _, err := targetNode.GetSettlementsLineInfoByAddress(n, equivalent)
	if err != nil {
		t.Fatalf("failed to get settlement line info: %v", err)
	}

	if settlementLineInfo.State != targetNodeSettlementLineInfo.State {
		t.Fatalf("settlement line state is not synced")
	}

	if settlementLineInfo.MaxPositiveBalance != targetNodeSettlementLineInfo.MaxNegativeBalance {
		t.Fatalf("settlement line max positive balance is not synced")
	}

	if settlementLineInfo.MaxNegativeBalance != targetNodeSettlementLineInfo.MaxPositiveBalance {
		t.Fatalf("settlement line max negative balance is not synced")
	}

	settlementLineBalance, err := strconv.Atoi(settlementLineInfo.Balance)
	if err != nil {
		t.Fatalf("failed to convert settlement line balance to int: %v", err)
	}
	targetNodeSettlementLineBalance, err := strconv.Atoi(targetNodeSettlementLineInfo.Balance)
	if err != nil {
		t.Fatalf("failed to convert settlement line balance to int: %v", err)
	}
	if settlementLineBalance != -targetNodeSettlementLineBalance {
		t.Fatalf("settlement line balance is not synced")
	}
}

func CheckSettlementLineForSyncBatch(t *testing.T, nodes []*Node, equivalent string, timeToSleepSeconds int) {
	time.Sleep(time.Duration(timeToSleepSeconds) * time.Second)
	for _, node := range nodes {
		println(fmt.Sprintf("check node for batch sync: %s", node.Alias))
		settlementLines, err := node.GetSettlementLines(equivalent)
		if err != nil {
			t.Fatalf("failed to get settlement lines: %v", err)
		}
		println(fmt.Sprintf("settlement lines: %+v", settlementLines))
		for _, settlementLine := range settlementLines {
			var targetNode *Node
			for _, otherNode := range nodes {
				if otherNode.GetIpAndPort() == settlementLine.ContractorAddress {
					targetNode = otherNode
					break
				}
			}
			if targetNode == nil {
				t.Fatalf("target node not found for settlement line: %+v", settlementLine)
			}
			println(fmt.Sprintf("check target node for batch sync: %s", targetNode.Alias))
			node.CheckSettlementLineForSync(t, targetNode, equivalent)
		}
	}
}

// CreateTransaction initiates a transaction to the target node.
// It first gets the contractor_id using getChannelInfo.
func (n *Node) CreateTransactionCheckStatus(t *testing.T, targetNode *Node, equivalent string, amount string, expectedStatus int) (string, error) {
	url := fmt.Sprintf("http://%s:%d/api/v1/node/contractors/transactions/%s/?contractor_address=%s&amount=%s",
		n.IPAddress, n.CLIPort, equivalent, targetNode.GetIPAddressForRequests(), amount)

	// No request body needed, parameters are in the URL query
	resp, err := http.Post(url, "application/json", nil) // Body is nil
	if err != nil {
		t.Fatalf("failed to send create transaction request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus { // Documentation example implies 200 OK
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("create transaction request failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Decode response to get transaction_uuid if needed
	var result struct {
		Data struct {
			TransactionUUID string `json:"transaction_uuid"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode create transaction response: %v", err)
	}

	t.Logf("transaction_uuid: %s", result.Data.TransactionUUID)

	return result.Data.TransactionUUID, nil
}

func (n *Node) GetMaxFlow(t *testing.T, targetNode *Node, equivalent string) (string, error) {
	url := fmt.Sprintf("http://%s:%d/api/v1/node/contractors/transactions/max/%s/?contractor_address=%s",
		n.IPAddress, n.CLIPort, equivalent, targetNode.GetIPAddressForRequests())

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to send max-flow request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("max-flow request failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Data MaxFlowInfo `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode max-flow response: %w", err)
	}
	println(fmt.Sprintf("max-flow response: %+v", result.Data))

	if result.Data.Count != 1 {
		return "", fmt.Errorf("max-flow response has wrong count. expected: 1, got: %d", result.Data.Count)
	}

	return result.Data.Records[0].MaxAmount, nil
}

func (n *Node) CheckMaxFlow(t *testing.T, targetNode *Node, equivalent string, expectedMaxFlow string) {
	maxFlow, err := n.GetMaxFlow(t, targetNode, equivalent)
	if err != nil {
		t.Fatalf("failed to get max-flow: %v", err)
	}
	if maxFlow != expectedMaxFlow {
		t.Fatalf("max-flow is wrong. expected: %s, got: %s", expectedMaxFlow, maxFlow)
	}
}

func (n *Node) GetMaxFlowBatch(t *testing.T, targetNodes []*Node, equivalent string) []MaxFlowBatchResult {
	if len(targetNodes) == 0 {
		return []MaxFlowBatchResult{}
	}

	baseURL := fmt.Sprintf("http://%s:%d/api/v1/node/contractors/transactions/max/%s/",
		n.IPAddress, n.CLIPort, equivalent)

	var queryParams []string
	for _, targetNode := range targetNodes {
		queryParams = append(queryParams, fmt.Sprintf("contractor_address=%s", targetNode.GetIPAddressForRequests()))
	}
	url := baseURL + "?" + queryParams[0]
	if len(queryParams) > 1 {
		for i := 1; i < len(queryParams); i++ {
			url += "&" + queryParams[i]
		}
	}

	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to send max-flow batch request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("max-flow batch request failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Data MaxFlowInfo `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode max-flow batch response: %v", err)
	}

	if result.Data.Count != len(targetNodes) {
		t.Fatalf("max-flow batch response has wrong count. expected: %d, got: %d. Records: %+v", len(targetNodes), result.Data.Count, result.Data.Records)
	}

	if len(result.Data.Records) != len(targetNodes) {
		// This check is important if the API might return fewer records than addresses requested,
		// even if Count matches.
		t.Fatalf("max-flow batch response records count (%d) does not match target nodes count (%d). Records: %+v", len(result.Data.Records), len(targetNodes), result.Data.Records)
	}

	maxFlowResults := make([]MaxFlowBatchResult, len(result.Data.Records))
	for i, record := range result.Data.Records {
		maxFlowResults[i] = MaxFlowBatchResult{
			ContractorAddress: record.ContractorAddress,
			MaxAmount:         record.MaxAmount,
		}
	}

	return maxFlowResults
}

func (n *Node) CheckMaxFlowBatch(t *testing.T, checks []MaxFlowBatchCheck, equivalent string) {
	if len(checks) == 0 {
		return
	}

	targetNodes := make([]*Node, len(checks))
	for i, check := range checks {
		targetNodes[i] = check.Node
	}

	maxFlowResults := n.GetMaxFlowBatch(t, targetNodes, equivalent)

	if len(maxFlowResults) != len(checks) {
		// This might be redundant if GetMaxFlowBatch already fatals, but good for safety.
		t.Fatalf("number of max flows received (%d) does not match number of checks (%d)", len(maxFlowResults), len(checks))
	}

	resultsMap := make(map[string]string)
	for _, res := range maxFlowResults {
		resultsMap[res.ContractorAddress] = res.MaxAmount
	}

	for _, check := range checks {
		nodeAddress := check.Node.GetIpAndPort()
		actualMaxAmount, found := resultsMap[nodeAddress]
		if !found {
			t.Fatalf("max-flow result not found for contractor address: %s. Available results: %+v", nodeAddress, resultsMap)
		}
		if actualMaxAmount != check.ExpectedMaxFlow {
			t.Fatalf("max-flow for node %s (address %s) is wrong. expected: %s, got: %s",
				check.Node.Alias, nodeAddress, check.ExpectedMaxFlow, actualMaxAmount)
		}
	}
}

func (n *Node) SetTestingFlag(flag uint32, appliableNodeAddress string, appliableAmount string) error {
	url := fmt.Sprintf("http://%s:%d/api/v1/node/subsystems-controller/%d/?forbidden_address=%s&forbidden_amount=%s",
		n.IPAddress, n.CLIPortTest, flag, appliableNodeAddress, appliableAmount)

	request, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return fmt.Errorf("failed to send set testing flag request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	setResp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send subsystems-controller request: %w", err)
	}
	defer setResp.Body.Close()
	if setResp.StatusCode != http.StatusOK {
		return fmt.Errorf("subsystems-controller request failed with status: %d", setResp.StatusCode)
	}

	return nil
}

func (n *Node) SetTestingSLFlag(flag uint32, firstParam, secondParam, thirdParam string) error {
	url := fmt.Sprintf("http://%s:%d/api/v1/node/settlement-lines-influence/%d/?first_parameter=%s&second_parameter=%s&third_parameter=%s",
		n.IPAddress, n.CLIPortTest, flag, firstParam, secondParam, thirdParam)

	request, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return fmt.Errorf("failed to send set testing SL flag request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	setResp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send settlement-lines-influence request: %w", err)
	}
	defer setResp.Body.Close()
	if setResp.StatusCode != http.StatusOK {
		return fmt.Errorf("settlement-lines-influence request failed with status: %d", setResp.StatusCode)
	}

	return nil
}

// CheckPaymentTransaction queries the node's SQLite database within its Docker container
// to verify various aspects of payment transactions.
// - transactionState: Optional. If provided, checks the 'observing_state' of the latest payment transaction.
// - paymentTransactionsCount: Expected count of records in 'payment_transactions' table.
// - participantsVotesCount: Expected count of records in 'payment_participants_votes' table.
// - incomingReceiptsCount: Expected count of records in 'incoming_receipt' table.
// - outgoingReceiptsCount: Expected count of records in 'outgoing_receipt' table.
func (n *Node) CheckPaymentTransaction(
	t *testing.T,
	transactionState string,
	paymentTransactionsCount int,
	participantsVotesCount int,
	incomingReceiptsCount int,
	outgoingReceiptsCount int,
) {
	if n.ContainerID == "" {
		t.Fatalf("Node %s: ContainerID is not set, cannot execute database checks.", n.Alias)
	}

	dbPath := "/vtcp/vtcpd/io/storageDB" // As specified by the user

	executeQuery := func(query string) (string, error) {
		cmdArgs := []string{"exec", n.ContainerID, "sqlite3", dbPath, query}
		cmd := exec.Command("docker", cmdArgs...)

		output, err := cmd.CombinedOutput()
		trimmedOutput := strings.TrimSpace(string(output))

		if err != nil {
			return trimmedOutput, fmt.Errorf("docker exec command failed for query ['%s'] on node %s (container: %s): %v. Output: %s", query, n.Alias, n.ContainerID, err, trimmedOutput)
		}
		return trimmedOutput, nil
	}

	// 1. Check transaction_state (if provided)
	if transactionState != "" {
		query := "SELECT observing_state FROM payment_transactions ORDER BY recording_time DESC LIMIT 1"
		actualState, err := executeQuery(query)
		if err != nil {
			// If executeQuery returns an error, it means the command failed (e.g. docker exec error, sqlite error like no such table)
			// If the query executes but returns no rows (e.g. table is empty), sqlite3 CLI typically outputs nothing and exits 0.
			// In that case, actualState would be an empty string.
			if actualState == "" && strings.Contains(err.Error(), "Output: ") { // Check if error message already contains output
				t.Fatalf("Node %s: Error querying transaction state. Query: '%s'. Error: %v", n.Alias, query, err)
			} else {
				t.Fatalf("Node %s: Error querying transaction state. Query: '%s'. Error: %v. Output: %s", n.Alias, query, err, actualState)
			}
		}
		// If query was successful but returned no rows (e.g. payment_transactions table is empty), actualState will be ""
		if actualState != transactionState {
			t.Fatalf("Node %s: Transaction state mismatch. Expected: '%s', Got: '%s'. Query: '%s'", n.Alias, transactionState, actualState, query)
		}
	}

	// Helper for checking counts
	checkCount := func(tableName string, expectedCount int) {
		query := fmt.Sprintf("SELECT count(*) FROM %s", tableName)
		countStr, err := executeQuery(query)
		if err != nil {
			t.Fatalf("Node %s: Error querying count for table '%s'. Query: '%s'. Error: %v. Output: %s", n.Alias, tableName, query, err, countStr)
		}
		actualCount, convErr := strconv.Atoi(countStr)
		if convErr != nil {
			t.Fatalf("Node %s: Error converting count '%s' to int for table '%s'. Query: '%s'. Error: %v", n.Alias, countStr, tableName, query, convErr)
		}
		if actualCount != expectedCount {
			t.Fatalf("Node %s: Count mismatch for table '%s'. Expected: %d, Got: %d. Query: '%s'", n.Alias, tableName, expectedCount, actualCount, query)
		}
	}

	// 2. Check payment_transactions count
	checkCount("payment_transactions", paymentTransactionsCount)

	// 3. Check payment_participants_votes count
	checkCount("payment_participants_votes", participantsVotesCount)

	// 4. Check incoming_receipt count
	checkCount("incoming_receipt", incomingReceiptsCount)

	// 5. Check outgoing_receipt count
	checkCount("outgoing_receipt", outgoingReceiptsCount)
}

// CheckSerializedTransaction queries the node's SQLite database within its Docker container
// to check for the presence and count of records in the 'transactions' table.
// - t: The testing.T instance for logging and failing tests.
// - isTransactionShouldBePresent: A boolean indicating if a transaction record is expected (true means 1, false means 0).
// - timeToSleepSeconds: The number of seconds to sleep before performing the check.
func (n *Node) CheckSerializedTransaction(
	t *testing.T,
	isTransactionShouldBePresent bool,
	timeToSleepSeconds int,
) {
	if timeToSleepSeconds > 0 {
		time.Sleep(time.Duration(timeToSleepSeconds) * time.Second)
	}

	if n.ContainerID == "" {
		t.Fatalf("Node %s: ContainerID is not set, cannot execute database checks.", n.Alias)
	}

	dbPath := "/vtcp/vtcpd/io/storageDB" // As specified by the user

	executeQuery := func(query string) (string, error) {
		cmdArgs := []string{"exec", n.ContainerID, "sqlite3", dbPath, query}
		cmd := exec.Command("docker", cmdArgs...)

		output, err := cmd.CombinedOutput()
		trimmedOutput := strings.TrimSpace(string(output))

		if err != nil {
			return trimmedOutput, fmt.Errorf("docker exec command failed for query ['%s'] on node %s (container: %s): %v. Output: %s", query, n.Alias, n.ContainerID, err, trimmedOutput)
		}
		return trimmedOutput, nil
	}

	query := "SELECT count(*) FROM transactions"
	countStr, err := executeQuery(query)
	if err != nil {
		// If executeQuery returns an error, it implies command failure or sqlite error (e.g., table doesn't exist)
		// countStr might contain stderr output in such cases.
		t.Fatalf("Node %s: Error querying count from 'transactions' table. Query: '%s'. Error: %v. Output: %s", n.Alias, query, err, countStr)
	}

	actualCount, convErr := strconv.Atoi(countStr)
	if convErr != nil {
		t.Fatalf("Node %s: Error converting count '%s' to int for 'transactions' table. Query: '%s'. SQLite output: '%s'. Conversion error: %v", n.Alias, countStr, query, countStr, convErr)
	}

	expectedCount := 0
	if isTransactionShouldBePresent {
		expectedCount = 1
	}

	if actualCount != expectedCount {
		if isTransactionShouldBePresent {
			t.Fatalf("Node %s: Expected 1 serialized transaction in DB, but found %d. Query: '%s'", n.Alias, actualCount, query)
		} else {
			t.Fatalf("Node %s: Expected 0 serialized transactions in DB, but found %d. Query: '%s'", n.Alias, actualCount, query)
		}
	}
}

// CheckValidKeys queries the node's SQLite database to check the count of valid own and contractor keys.
func (n *Node) CheckValidKeys(t *testing.T, expectedOwnValidKeysCount, expectedContractorValidKeysCount int) {
	if n.ContainerID == "" {
		t.Fatalf("Node %s: ContainerID is not set, cannot execute database checks.", n.Alias)
	}

	dbPath := "/vtcp/vtcpd/io/storageDB"

	executeQuery := func(query string) (string, error) {
		cmdArgs := []string{"exec", n.ContainerID, "sqlite3", dbPath, query}
		cmd := exec.Command("docker", cmdArgs...)
		output, err := cmd.CombinedOutput()
		trimmedOutput := strings.TrimSpace(string(output))
		if err != nil {
			return trimmedOutput, fmt.Errorf("docker exec command failed for query ['%s'] on node %s (container: %s): %v. Output: %s", query, n.Alias, n.ContainerID, err, trimmedOutput)
		}
		return trimmedOutput, nil
	}

	// Check own_keys
	queryOwnKeys := "SELECT COUNT(*) FROM own_keys WHERE is_valid = 1"
	ownKeysCountStr, err := executeQuery(queryOwnKeys)
	if err != nil {
		t.Fatalf("Node %s: Error querying own_keys count. Query: '%s'. Error: %v. Output: %s", n.Alias, queryOwnKeys, err, ownKeysCountStr)
	}
	ownKeysCount, convErr := strconv.Atoi(ownKeysCountStr)
	if convErr != nil {
		t.Fatalf("Node %s: Error converting own_keys count '%s' to int. Query: '%s'. Error: %v", n.Alias, ownKeysCountStr, queryOwnKeys, convErr)
	}
	if ownKeysCount != expectedOwnValidKeysCount {
		t.Fatalf("Node %s: Own valid keys count mismatch. Expected: %d, Got: %d. Query: '%s'", n.Alias, expectedOwnValidKeysCount, ownKeysCount, queryOwnKeys)
	}

	// Check contractor_keys
	queryContractorKeys := "SELECT COUNT(*) FROM contractor_keys WHERE is_valid = 1"
	contractorKeysCountStr, err := executeQuery(queryContractorKeys)
	if err != nil {
		t.Fatalf("Node %s: Error querying contractor_keys count. Query: '%s'. Error: %v. Output: %s", n.Alias, queryContractorKeys, err, contractorKeysCountStr)
	}
	contractorKeysCount, convErr := strconv.Atoi(contractorKeysCountStr)
	if convErr != nil {
		t.Fatalf("Node %s: Error converting contractor_keys count '%s' to int. Query: '%s'. Error: %v", n.Alias, contractorKeysCountStr, queryContractorKeys, convErr)
	}
	if contractorKeysCount != expectedContractorValidKeysCount {
		t.Fatalf("Node %s: Contractor valid keys count mismatch. Expected: %d, Got: %d. Query: '%s'", n.Alias, expectedContractorValidKeysCount, contractorKeysCount, queryContractorKeys)
	}
}

// CheckSettlementLineState queries the node's SQLite database to check the state of a trust line.
// equivalent is typically "1" or another string.
func (n *Node) CheckSettlementLineState(t *testing.T, targetNode *Node, equivalent string, expectedState string) {
	if n.ContainerID == "" {
		t.Fatalf("Node %s: ContainerID is not set, cannot execute database checks.", n.Alias)
	}

	// First, get contractor_id (channel_id) for the target node, as this is used in the trust_lines table.
	// This assumes that a channel must exist for a trust line to exist, which is typical.
	// If trust lines can exist without a channel, this part needs adjustment or contractor_id must be passed directly.
	channelInfo, err := n.GetChannelInfo(targetNode)
	if err != nil {
		t.Fatalf("Node %s: Failed to get channel info for target %s to find contractor_id: %v", n.Alias, targetNode.Alias, err)
	}
	contractorID := channelInfo.ChannelID // This is the contractor_id in the context of trust_lines table

	dbPath := "/vtcp/vtcpd/io/storageDB"
	query := fmt.Sprintf("SELECT state FROM trust_lines WHERE contractor_id = '%s' AND equivalent = %s", contractorID, equivalent)

	executeQuery := func(q string) (string, error) {
		cmdArgs := []string{"exec", n.ContainerID, "sqlite3", dbPath, q}
		cmd := exec.Command("docker", cmdArgs...)
		output, errCmd := cmd.CombinedOutput()
		trimmedOutput := strings.TrimSpace(string(output))
		if errCmd != nil {
			return trimmedOutput, fmt.Errorf("docker exec command failed for query ['%s'] on node %s: %v. Output: %s", q, n.Alias, errCmd, trimmedOutput)
		}
		return trimmedOutput, nil
	}

	actualState, err := executeQuery(query)
	if err != nil {
		t.Fatalf("Node %s: Error querying trust_line state. Query: '%s'. Error: %v", n.Alias, query, err)
	}
	if actualState == "" {
		t.Fatalf("Node %s: No trust_line state returned for query: '%s'. The trust line might not exist for contractor_id '%s' and equivalent '%s'.", n.Alias, query, contractorID, equivalent)
	}

	if actualState != expectedState {
		t.Fatalf("Node %s: TrustLine state mismatch for contractor %s (ID: %s), equivalent %s. Expected: %s, Got: %s. Query: '%s'",
			n.Alias, targetNode.Alias, contractorID, equivalent, expectedState, actualState, query)
	}
}

// CheckPaymentRecordWithCommandUUID queries the node's SQLite database
// to check for the presence of a payment record with a specific command_uuid.
func (n *Node) CheckPaymentRecordWithCommandUUID(t *testing.T, commandUUID string, shouldBePresent bool) {
	if n.ContainerID == "" {
		t.Fatalf("Node %s: ContainerID is not set, cannot execute database checks.", n.Alias)
	}

	dbPath := "/vtcp/vtcpd/io/storageDB"
	// The command_uuid in the database is stored as a blob without dashes.
	formattedCommandUUID := strings.ReplaceAll(commandUUID, "-", "")
	query := fmt.Sprintf("SELECT COUNT(*) FROM history WHERE command_uuid = x'%s'", formattedCommandUUID)

	executeQuery := func(q string) (string, error) {
		cmdArgs := []string{"exec", n.ContainerID, "sqlite3", dbPath, q}
		cmd := exec.Command("docker", cmdArgs...)
		output, errCmd := cmd.CombinedOutput()
		trimmedOutput := strings.TrimSpace(string(output))
		if errCmd != nil {
			return trimmedOutput, fmt.Errorf("docker exec command failed for query ['%s'] on node %s: %v. Output: %s", q, n.Alias, errCmd, trimmedOutput)
		}
		return trimmedOutput, nil
	}

	countStr, err := executeQuery(query)
	if err != nil {
		t.Fatalf("Node %s: Error querying history for command_uuid '%s'. Query: '%s'. Error: %v", n.Alias, commandUUID, query, err)
	}

	actualCount, convErr := strconv.Atoi(countStr)
	if convErr != nil {
		t.Fatalf("Node %s: Error converting count '%s' to int for history query. Command UUID: '%s'. Query: '%s'. Error: %v", n.Alias, countStr, commandUUID, query, convErr)
	}

	if shouldBePresent {
		if actualCount == 0 {
			t.Fatalf("Node %s: Expected payment record with command_uuid '%s' to be present, but found 0. Query: '%s'", n.Alias, commandUUID, query)
		}
		if actualCount > 1 {
			// This might be a valid scenario depending on the system, but the Python test implies checking for existence (not necessarily count=1)
			// For now, just log or consider if this should be a t.Fatalf
			t.Logf("Node %s: Warning - found %d records for command_uuid '%s'. Expected at least 1. Query: '%s'", n.Alias, actualCount, commandUUID, query)
		}
	} else {
		if actualCount > 0 {
			t.Fatalf("Node %s: Expected no payment record with command_uuid '%s', but found %d. Query: '%s'", n.Alias, commandUUID, actualCount, query)
		}
	}
}

// CheckCurrentAudit queries the node's SQLite database to check the current audit number for a trust line.
// equivalent is typically "1" or another integer string.
func (n *Node) CheckCurrentAudit(t *testing.T, targetNode *Node, equivalent string, expectedAuditNumber int) {
	if n.ContainerID == "" {
		t.Fatalf("Node %s: ContainerID is not set, cannot execute database checks.", n.Alias)
	}

	channelInfo, err := n.GetChannelInfo(targetNode)
	if err != nil {
		t.Fatalf("Node %s: Failed to get channel info for target %s to find contractor_id for audit check: %v", n.Alias, targetNode.Alias, err)
	}
	contractorID := channelInfo.ChannelID

	dbPath := "/vtcp/vtcpd/io/storageDB"
	// First, get the trust_line_id
	queryTrustLineID := fmt.Sprintf("SELECT id FROM trust_lines WHERE contractor_id = '%s' AND equivalent = %s", contractorID, equivalent)

	executeQuery := func(q string) (string, error) {
		cmdArgs := []string{"exec", n.ContainerID, "sqlite3", dbPath, q}
		cmd := exec.Command("docker", cmdArgs...)
		output, errCmd := cmd.CombinedOutput()
		trimmedOutput := strings.TrimSpace(string(output))
		if errCmd != nil {
			return trimmedOutput, fmt.Errorf("docker exec command failed for query ['%s'] on node %s: %v. Output: %s", q, n.Alias, errCmd, trimmedOutput)
		}
		return trimmedOutput, nil
	}

	trustLineIDStr, err := executeQuery(queryTrustLineID)
	if err != nil {
		t.Fatalf("Node %s: Error querying trust_line id. Query: '%s'. Error: %v", n.Alias, queryTrustLineID, err)
	}
	if trustLineIDStr == "" {
		t.Fatalf("Node %s: No trust_line id returned for query: '%s'. The trust line might not exist for contractor_id '%s' and equivalent '%s'.", n.Alias, queryTrustLineID, contractorID, equivalent)
	}
	// trust_line_id is an integer, but we use it in the next query as a string.
	// No conversion to int is strictly needed for the SQL query itself.

	queryAudit := fmt.Sprintf("SELECT number FROM audit WHERE trust_line_id = %s ORDER BY number DESC LIMIT 1", trustLineIDStr)
	actualAuditStr, err := executeQuery(queryAudit)
	if err != nil {
		t.Fatalf("Node %s: Error querying audit number. Query: '%s'. Error: %v", n.Alias, queryAudit, err)
	}
	if actualAuditStr == "" {
		// This could mean no audit entries exist yet for this trust line.
		// Depending on expectations, this might be a valid state (e.g., for a new trust line).
		// The Python code assumes an audit entry will be found.
		// If expectedAuditNumber is 0 and actualAuditStr is "", this might be a pass condition.
		// For now, strict check:
		t.Fatalf("Node %s: No audit number returned for trust_line_id '%s'. Query: '%s'", n.Alias, trustLineIDStr, queryAudit)
	}

	actualAuditNumber, convErr := strconv.Atoi(actualAuditStr)
	if convErr != nil {
		t.Fatalf("Node %s: Error converting audit number '%s' to int. Query: '%s'. Error: %v", n.Alias, actualAuditStr, queryAudit, convErr)
	}

	if actualAuditNumber != expectedAuditNumber {
		t.Fatalf("Node %s: Current audit number mismatch for contractor %s (TrustLineID: %s), equivalent %s. Expected: %d, Got: %d. Query: '%s'", n.Alias, targetNode.Alias, trustLineIDStr, equivalent, expectedAuditNumber, actualAuditNumber, queryAudit)
	}
}

// CheckNodeForLogMessage checks if a specific message (optionally for a specific transaction UUID) exists in the node's operations.log.
// - transactionUUID: Optional. If provided, the log line must contain this UUID.
// - message: The specific string to search for in the log line.
// - expectedToFind: Boolean indicating whether the message is expected to be found.
func (n *Node) CheckNodeForLogMessage(t *testing.T, transactionUUID string, message string, expectedToFind bool) {
	if n.ContainerID == "" {
		t.Fatalf("Node %s: ContainerID is not set, cannot execute log checks.", n.Alias)
	}

	// Construct the grep command. -q for quiet (no output), exit status indicates match.
	// We will use a more verbose grep and check the output directly to handle cases where the log might be empty
	// or the file doesn't exist, which `grep -q` might not distinguish well for our needs.
	var grepCmd string
	if transactionUUID != "" {
		// Search for lines containing both the transactionUUID and the message.
		// Using .* to allow any characters between UUID and message in any order on the line.
		// This is a simplified approach. For more robust matching, more complex awk/grep could be used.
		// Escaping for grep can be tricky; assuming simple strings for now.
		grepCmd = fmt.Sprintf("grep '%s' %s | grep '%s'", transactionUUID, DefaultOperationsLogPath, message)
	} else {
		grepCmd = fmt.Sprintf("grep '%s' %s", message, DefaultOperationsLogPath)
	}

	fullCmd := []string{"exec", n.ContainerID, "sh", "-c", grepCmd}
	cmd := exec.Command("docker", fullCmd...)

	output, err := cmd.CombinedOutput()
	trimmedOutput := strings.TrimSpace(string(output))

	found := false
	if err == nil && trimmedOutput != "" {
		// `grep` found matches and output them.
		found = true
	} else if exitError, ok := err.(*exec.ExitError); ok {
		// `grep` exited with non-zero status.
		// Exit status 1 means no lines were selected.
		// Other exit statuses (e.g., 2) indicate an error with grep itself (e.g., file not found).
		if exitError.ExitCode() == 1 {
			found = false // No match
		} else {
			// Grep error (e.g. file not found by grep within container)
			t.Fatalf("Node %s: Error executing grep command for log check. Command: '%s'. Error: %v. Output: %s", n.Alias, strings.Join(fullCmd, " "), err, trimmedOutput)
			return
		}
	} else if err != nil {
		// Other execution error (e.g., docker command itself failed to start)
		t.Fatalf("Node %s: Failed to execute docker command for log check. Command: '%s'. Error: %v. Output: %s", n.Alias, strings.Join(fullCmd, " "), err, trimmedOutput)
		return
	}

	if expectedToFind && !found {
		t.Fatalf("Node %s: Expected to find log message '%s' (UUID: '%s'), but did not. Grep output: %s", n.Alias, message, transactionUUID, trimmedOutput)
	} else if !expectedToFind && found {
		t.Fatalf("Node %s: Expected NOT to find log message '%s' (UUID: '%s'), but did. Grep output: %s", n.Alias, message, transactionUUID, trimmedOutput)
	}
}
