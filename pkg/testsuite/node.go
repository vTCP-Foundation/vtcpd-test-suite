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
			"LISTEN_ADDRESS=0.0.0.0",
			"LISTEN_PORT=2000",
			"EQUIVALENTS_REGISTRY=eth",
			"MAX_HOPS=5",
		},
	}
}
