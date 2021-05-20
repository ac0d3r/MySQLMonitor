package tcp

import (
	"fmt"
	"net"
	"time"
)

// FullConnectTest tcp 完全连接测试
func FullConnectTest(ip string, port int, timeout time.Duration) bool {
	var (
		conn net.Conn
		err  error
	)
	if conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), timeout); err != nil {
		return false
	}
	if conn != nil {
		_ = conn.Close()
		return true
	}
	return false
}
