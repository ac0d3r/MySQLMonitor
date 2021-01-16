package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/Buzz2d0/SecTools/pkg"
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
	scanMode  mode
	threadNum int
	timeout   time.Duration
	taskChan  chan target
	wg        *sync.WaitGroup = &sync.WaitGroup{}

	flagHelp     *bool   = flag.Bool("h", false, "Shows usage options.")
	flagIPList   *string = flag.String("t", "", "target ip list")
	flagPortList *string = flag.String("p", "22,80,3306,8080", "target port list")
	flagSyn      *bool   = flag.Bool("syn", false, "set scan mode with \"syn\"")
	flagTimeout  *uint   = flag.Uint("timeout", 2, "set connent timeout")
)

func initOptions(targetNum int) {
	// init concurrent quantity
	if targetNum > 1000 {
		threadNum = 1000
	} else {
		threadNum = targetNum
	}
	taskChan = make(chan target, threadNum)
	timeout = time.Duration(*flagTimeout) * time.Second
	if *flagSyn {
		fmt.Println("\033[91m not support syn")
		os.Exit(0)
		scanMode = syn
		// check run with root
		if !pkg.IsRoot() {
			fmt.Println("\033[91m must run with root!")
			os.Exit(1)
		}
	} else {
		scanMode = full
	}

}

func genTasks(ipList []net.IP, portList []int) {
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
		wg.Add(1)
		go worker()
	}
}

func main() {
	flag.Parse()
	if *flagHelp || *flagIPList == "" {
		fmt.Printf("Usage: PortScanner [options]\n\n")
		flag.PrintDefaults()
		return
	}
	ipList, err := parse.GetIPList(*flagIPList)
	if err != nil {
		log.Fatalln(err)
	}
	portList, err := parse.GetPorts(*flagPortList)
	if err != nil {
		log.Fatalln(err)
	}
	// run
	initOptions(len(ipList))
	wg.Add(1)
	go genTasks(ipList, portList)
	scan(scanMode)
	wg.Wait()
}
