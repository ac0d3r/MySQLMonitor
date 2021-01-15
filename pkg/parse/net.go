package parse

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/malfunkt/iprange"
)

// AnalysePort 解析 port (0-65535)
func AnalysePort(portStr string) (int, error) {
	var (
		port int
		err  error
	)
	port, err = strconv.Atoi(portStr)
	if err != nil || port < 0 || port > 65535 {
		return 0, fmt.Errorf("Invalid port number: '%s'", portStr)
	}
	return port, nil
}

// GetPorts doc
// 多端口处理：支持 `,` 分割端口以及 `-` 表示的端口范围(如："22-80")。
func GetPorts(portsStr string) ([]int, error) {
	var (
		ports         []int
		ranges, parts []string
	)
	if portsStr == "" {
		return ports, nil
	}
	ranges = strings.Split(portsStr, ",")
	for _, r := range ranges {
		r = strings.TrimSpace(r)
		if strings.Contains(r, "-") {
			parts = strings.Split(r, "-")
			if len(parts) != 2 {
				return nil, fmt.Errorf("Invalid port selection segment: '%s'", r)
			}
			prePort, err := AnalysePort(parts[0])
			if err != nil {
				return nil, err
			}
			sufPort, err := AnalysePort(parts[1])
			if err != nil {
				return nil, err
			}
			if prePort > sufPort {
				return nil, fmt.Errorf("Invalid port range: %d-%d", prePort, sufPort)
			}
			for i := prePort; i <= sufPort; i++ {
				ports = append(ports, i)
			}
		} else {
			port, err := strconv.Atoi(r)
			if err != nil || port < 0 || port > 65535 {
				return nil, fmt.Errorf("Invalid port number: '%s'", r)
			}
			ports = append(ports, port)
		}
	}
	return ports, nil
}

// GetIPList 解析 iprange
func GetIPList(iprangeStr string) ([]net.IP, error) {
	var (
		addrRan iprange.AddressRangeList
		err     error
	)
	addrRan, err = iprange.ParseList(iprangeStr)
	if err != nil {
		return nil, err
	}
	return addrRan.Expand(), nil
}
