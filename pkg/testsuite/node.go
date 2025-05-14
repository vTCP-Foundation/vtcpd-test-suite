package testsuite

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	StatusOK = 200
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
	State                 string `json:"state"`
	OwnKeysPresent        string `json:"own_keys_present"`
	ContractorKeysPresent string `json:"contractor_keys_present"`
	// AuditNumber           string `json:"audit_number"`
	MaxNegativeBalance string `json:"max_negative_balance"`
	MaxPositiveBalance string `json:"max_positive_balance"`
	Balance            string `json:"balance"`
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

// CreateSettlementLine creates a settlement line with another node.
// It first gets the contractor_id using getChannelInfo, then calls init-settlement-line and sets the max positive balance.
// Returns error if any step fails.
func (n *Node) CreateSettlementLine(t *testing.T, targetNode *Node, equivalent string, amount string) {
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

	time.Sleep(1 * time.Second)

	// Step 3: Set max positive balance (PUT)
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
func (n *Node) GetSettlementsLineInfoByAddress(targetNode *Node, equivalent string) (*SettlementLineInfo, error) {
	url := fmt.Sprintf("http://%s:%d/api/v1/node/contractors/settlement-line-by-address/%s/?contractor_address=%s",
		n.IPAddress, n.CLIPort, equivalent, targetNode.GetIPAddressForRequests())

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send settlement-line-by-address request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("settlement-line-by-address request failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			SettlementLine SettlementLineInfo `json:"settlement_line"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode settlement-line-by-address response: %w", err)
	}

	return &result.Data.SettlementLine, nil
}

func (n *Node) CheckSettlementLine(t *testing.T, targetNode *Node, equivalent, state, maxPositiveBalance, maxNegativeBalance,
	balance, ownKeysPresent, contractorKeysPresent string) {
	settlementLineInfo, err := n.GetSettlementsLineInfoByAddress(targetNode, equivalent)
	if err != nil {
		t.Fatalf("failed to get settlement line info: %v", err)
	}
	if settlementLineInfo.State != state {
		t.Fatalf("settlement line is not active")
	}
	if settlementLineInfo.MaxPositiveBalance != maxPositiveBalance {
		t.Fatalf("max positive balance is not set")
	}
	if settlementLineInfo.MaxNegativeBalance != maxNegativeBalance {
		t.Fatalf("max negative balance is not set")
	}
	if settlementLineInfo.Balance != balance {
		t.Fatalf("balance is not set. expected: %s, got: %s", balance, settlementLineInfo.Balance)
	}
	if settlementLineInfo.OwnKeysPresent != ownKeysPresent {
		t.Fatalf("own keys are not present")
	}
	if settlementLineInfo.ContractorKeysPresent != contractorKeysPresent {
		t.Fatalf("contractor keys are not present")
	}
}

func (n *Node) CheckActiveSettlementLine(t *testing.T, targetNode *Node, equivalent, maxPositiveBalance, maxNegativeBalance, balance string) {
	n.CheckSettlementLine(t, targetNode, equivalent, SettlementLineStateActive, maxPositiveBalance, maxNegativeBalance, balance,
		SettlementLineKeysPresent, SettlementLineKeysPresent)
}

func (n *Node) CreateSettlementLineAndCheck(t *testing.T, targetNode *Node, equivalent string, amount string) {
	n.CreateSettlementLine(t, targetNode, equivalent, amount)

	time.Sleep(500 * time.Millisecond)

	n.CheckActiveSettlementLine(t, targetNode, equivalent, amount, "0", "0")
	targetNode.CheckActiveSettlementLine(t, n, equivalent, "0", amount, "0")
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
		nodeAddress := check.Node.GetIPAddressForRequests()
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
	url := fmt.Sprintf("http://%s:%d/api/v1/node/subsystems-controller/%d/?d&forbidden_address=%s&forbidden_amount=%s",
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
	url := fmt.Sprintf("http://%s:%d/api/v1/node/settlement-lines-influence/%d/?d&first_parameter=%s&second_parameter=%s&third_parameter=%s",
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
