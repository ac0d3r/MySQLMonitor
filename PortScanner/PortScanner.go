package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/Buzz2d0/SecTools/pkg/parse"
	"github.com/Buzz2d0/SecTools/pkg/tcp"
)

type mode int

type target struct {
	ip   string
	port int
}

const (
	full mode = iota
	syn
)

var (
	threadNum int
	timeout   time.Duration = 2 * time.Second
	taskChan  chan target
	wg        *sync.WaitGroup = &sync.WaitGroup{}
)

func tcpSynConnect(ip string, port int) {

}

func genTasks(ipList []net.IP, portList []int) {
	wg.Add(1)
	defer func() {
		wg.Done()
		close(taskChan)
	}()

	for _, ip := range ipList {
		for _, port := range portList {
			taskChan <- target{ip.String(), port}
		}
	}
}

func scan(mod mode) {
	worker := func() {
		wg.Add(1)
		defer wg.Done()

		for t := range taskChan {
			if mod == full {
				if tcp.FullConnectTest(t.ip, t.port, timeout) {
					fmt.Printf("\033[92m%s:%d alived\n", t.ip, t.port)
				}
			}
		}
	}
	for i := 0; i < threadNum; i++ {
		go worker()
	}
}

func main() {
	if len(os.Args) != 3 {
		log.Fatalln("Input Ips and Ports")
	}
	ipList, err := parse.GetIPList(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}
	portList, err := parse.GetPorts(os.Args[2])
	if err != nil {
		log.Fatalln(err)
	}
	// init concurrent quantity
	if len(ipList) > 1000 {
		threadNum = 1000
	} else {
		threadNum = len(ipList) / 2
	}
	taskChan = make(chan target, threadNum)
	// run
	go genTasks(ipList, portList)
	scan(full)
	wg.Wait()
}
