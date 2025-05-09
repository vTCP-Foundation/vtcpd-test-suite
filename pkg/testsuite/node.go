package testsuite

import (
	"testing"

	"github.com/google/uuid"
)

type Node struct {
	ID          string
	Host        string
	Port        uint16
	IPAddress   string
	ContainerID string
	Env         []string
}

func NewNode(t *testing.T) *Node {
	return &Node{
		ID:          uuid.New().String(),
		Host:        "0.0.0.0",
		Port:        2000,
		IPAddress:   "", // Must be set on container creation.
		ContainerID: "", // Must be set on container creation.
		Env: []string{
			"VTCPD_LISTEN_ADDRESS=0.0.0.0",
			"VTCPD_LISTEN_PORT=2000",
			"VTCPD_EQUIVALENTS_REGISTRY=eth",
			"VTCPD_MAX_HOPS=5",
			"CLI_LISTEN_ADDRESS=0.0.0.0",
			"CLI_LISTEN_PORT=3000",
			"CLI_LISTEN_PORT_TESTING=3001",
		},
	}
}
