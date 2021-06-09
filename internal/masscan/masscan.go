package masscan

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"

	"github.com/Buzz2d0/SecTools/internal/util"
)

type Out struct {
	IP       net.IP
	Port     uint
	Protocol string
}

type Masscan struct {
	File    string
	Port    string
	Rate    int
	IPRange []string
	cmd     *exec.Cmd
	out     chan Out
}

func New(file string, port string, ip ...string) *Masscan {
	return &Masscan{
		File:    file,
		Rate:    10000,
		Port:    port,
		IPRange: ip,
	}
}

func (m *Masscan) Start() (<-chan Out, error) {
	args := make([]string, 0)
	args = append(args, "-p", m.Port)
	args = append(args, fmt.Sprintf("--rate=%d", m.Rate))
	args = append(args, m.IPRange...)
	m.cmd = exec.Command(m.File, args...)
	m.out = make(chan Out, 1e3)
	err := util.StartCmd(m.cmd,
		func(line string) {
			if o := parseLine(line); o != nil {
				m.out <- *o
			}
		},
		func(line string) {
			log.Println("[Masscan]", line)
		},
		func() {
			close(m.out)
		},
	)
	if err != nil {
		return nil, err
	}
	return m.out, err
}

func (m *Masscan) Wait() error {
	if m.cmd == nil {
		return errors.New("masscan not started")
	}
	defer func() {
		m.cmd = nil
	}()
	return m.cmd.Wait()
}

func parseLine(line string) *Out {
	// Discovered open port 80/tcp on 192.168.1.1
	if !strings.HasPrefix(line, "Discovered") {
		return nil
	}
	item := strings.Split(line, " ")
	if len(item) < 6 {
		return nil
	}
	o := &Out{}
	o.IP = net.ParseIP(item[5])
	if o.IP == nil {
		return nil
	}

	if ss := strings.SplitN(item[3], "/", 2); len(ss) < 2 {
		return nil
	} else {
		port, err := strconv.ParseUint(ss[0], 10, 64)
		if err != nil || port == 0 {
			return nil
		}
		o.Port = uint(port)
		o.Protocol = ss[1]
	}
	return o
}
