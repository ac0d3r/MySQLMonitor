package tcp

import (
	"testing"
	"time"
)

func TestGetLocalIPPort(t *testing.T) {
	// ip, port, err := getLocalIPPort("192.168.31.1")
	// t.Log(ip, port, err)
}

func TestFullConnectTest(t *testing.T) {
	t.Log(FullConnectTest("192.168.31.1", 80, 3*time.Second))
}

func TestSynConnectTest(t *testing.T) {
	// t.Log(SynConnectTest("192.168.31.1", 80, 3*time.Second))
}
