package testsuite

import (
	"bytes"
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
)

type Node struct {
	ID          string
	Host        string
	NodePort    uint16
	CLIPort     uint16
	CLIPortTest uint16
	IPAddress   string
	ContainerID string
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

func NewNode(t *testing.T, ipAddress string) *Node {
	return &Node{
		ID:          uuid.New().String(),
		Host:        "0.0.0.0",
		NodePort:    DefaultNodePort,
		CLIPort:     DefaultCLIPort,
		CLIPortTest: DefaultCLIPortTest,
		IPAddress:   ipAddress,
		ContainerID: "", // Must be set on container creation.
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
func (n *Node) OpenChannel(targetNode *Node) error {
	// Step 1: This node initiates the channel
	// Prepare the request body with the target node's address
	// Using IPV4 type code 12
	initURL := fmt.Sprintf("http://%s:%d/api/v1/node/contractors/init-channel/?contractor_address=%s",
		n.IPAddress, n.CLIPort, targetNode.GetIPAddressForRequests())

	// Send the request to initialize the channel
	resp, err := http.Post(initURL, "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to send init-channel request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("init-channel request failed with status: %d", resp.StatusCode)
	}

	// Parse the response to get channel_id and crypto_key
	var initResponse struct {
		Data ChannelInitResponseData `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&initResponse); err != nil {
		return fmt.Errorf("failed to decode init-channel response: %w", err)
	}

	// Step 2: Target node completes the channel initialization
	// Prepare the request body with this node's address, channel_id, and crypto_key
	targetURL := fmt.Sprintf("http://%s:%d/api/v1/node/contractors/init-channel/?contractor_address=%s&contractor_id=%s&crypto_key=%s",
		targetNode.IPAddress, targetNode.CLIPort, n.GetIPAddressForRequests(), initResponse.Data.ChannelID, initResponse.Data.CryptoKey)

	// Send the request to complete channel initialization
	targetResp, err := http.Post(targetURL, "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to send target init-channel request: %w", err)
	}
	defer targetResp.Body.Close()

	if targetResp.StatusCode != http.StatusOK {
		return fmt.Errorf("target init-channel request failed with status: %d", targetResp.StatusCode)
	}

	return nil
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

// CreateSettlementLine creates a settlement line with another node.
// It first gets the contractor_id using getChannelInfo, then calls init-settlement-line and sets the max positive balance.
// Returns error if any step fails.
func (n *Node) CreateSettlementLine(targetNode *Node, equivalent string, amount string) error {
	// Step 1: Get contractor_id (channel_id) for the target node
	channelInfo, err := n.GetChannelInfo(targetNode)
	if err != nil {
		return fmt.Errorf("failed to get channel info: %w", err)
	}
	contractorID := channelInfo.ChannelID

	// Step 2: Call init-settlement-line
	initURL := fmt.Sprintf("http://%s:%d/api/v1/node/contractors/%s/init-settlement-line/%s/",
		n.IPAddress, n.CLIPort, contractorID, equivalent)
	initBody := map[string]string{
		"max_positive_balance": amount,
	}
	initJson, err := json.Marshal(initBody)
	if err != nil {
		return fmt.Errorf("failed to marshal init-settlement-line body: %w", err)
	}
	initResp, err := http.Post(initURL, "application/json", bytes.NewBuffer(initJson))
	if err != nil {
		return fmt.Errorf("failed to send init-settlement-line request: %w", err)
	}
	defer initResp.Body.Close()
	if initResp.StatusCode != http.StatusOK {
		return fmt.Errorf("init-settlement-line request failed with status: %d", initResp.StatusCode)
	}

	time.Sleep(1 * time.Second)

	// Step 3: Set max positive balance (PUT)
	setURL := fmt.Sprintf("http://%s:%d/api/v1/node/contractors/%s/settlement-lines/%s/?amount=%s",
		n.IPAddress, n.CLIPort, contractorID, equivalent, amount)
	request, err := http.NewRequest(http.MethodPut, setURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create set-settlement-line request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	setResp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send set-settlement-line request: %w", err)
	}
	defer setResp.Body.Close()
	if setResp.StatusCode != http.StatusOK {
		return fmt.Errorf("set-settlement-line request failed with status: %d", setResp.StatusCode)
	}

	return nil
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

// CreateTransaction initiates a transaction to the target node.
// It first gets the contractor_id using getChannelInfo.
func (n *Node) CreateTransaction(targetNode *Node, equivalent string, amount string) (string, error) {
	url := fmt.Sprintf("http://%s:%d/api/v1/node/contractors/transactions/%s/?contractor_address=%s&amount=%s",
		n.IPAddress, n.CLIPort, equivalent, targetNode.GetIPAddressForRequests(), amount)
	fmt.Println("Creating transaction to", url)

	// No request body needed, parameters are in the URL query
	resp, err := http.Post(url, "application/json", nil) // Body is nil
	if err != nil {
		return "", fmt.Errorf("failed to send create transaction request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK { // Documentation example implies 200 OK
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create transaction request failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Decode response to get transaction_uuid if needed
	var result struct {
		Data struct {
			TransactionUUID string `json:"transaction_uuid"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode create transaction response: %w", err)
	}

	return result.Data.TransactionUUID, nil
}

func (n *Node) SetTestingFlag(flag uint32, appliableNodeAddress string, appliableAmount string) error {
	url := fmt.Sprintf("http://%s:%d/api/v1/node/subsystems-controller/%d/?d&forbidden_address=%s&forbidden_amount=%s",
		n.IPAddress, n.CLIPort, flag, appliableNodeAddress, appliableAmount)

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
		n.IPAddress, n.CLIPort, flag, firstParam, secondParam, thirdParam)

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
