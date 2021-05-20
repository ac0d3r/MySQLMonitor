package parse

import (
	"testing"
)

func TestGetPorts(t *testing.T) {
	ports, err := GetPorts("80")
	t.Logf("%#v, %v", ports, err)

	ports, err = GetPorts("22,23,25-30")
	t.Logf("%#v, %v", ports, err)

	ports, err = GetPorts("1080-30")
	t.Logf("%#v, %v", ports, err)

	ports, err = GetPorts("1i-30")
	t.Logf("%#v, %v", ports, err)

	ports, err = GetPorts("22,23,25-30,2o")
	t.Logf("%#v, %v", ports, err)
}

func TestParsePort(t *testing.T) {
	port, err := AnalysePort("80")
	t.Logf("%d, %v", port, err)

	port, err = AnalysePort("-1")
	t.Logf("%d, %v", port, err)

	port, err = AnalysePort("65536")
	t.Logf("%d, %v", port, err)
}

func TestGetIPList(t *testing.T) {
	ips, err := GetIPList("10.0.0.1")
	t.Logf("%v, %v", ips, err)

	ips, err = GetIPList("10.0.0.1, 10.0.0.5-10")
	t.Logf("%v, %v", ips, err)

	ips, err = GetIPList("192.168.1.*")
	t.Logf("%v, %v", ips, err)

	ips, err = GetIPList("192.168.10.0/24")
	t.Logf("%v, %v", ips, err)
}
