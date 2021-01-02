package main

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	timeout = 5 * time.Second
)

func CheckSsh(addr string, username, password string) (result bool, err error) {
	var (
		config  *ssh.ClientConfig
		client  *ssh.Client
		session *ssh.Session
	)
	config = &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		Timeout: timeout,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	if client, err = ssh.Dial("tcp", addr, config); err != nil {
		return false, err
	}
	defer client.Close()
	session, err = client.NewSession()
	errRet := session.Run("echo xsec")
	if err == nil && errRet == nil {
		defer session.Close()
		result = true
	}

	return result, err
}

func main() {
	result, err := CheckSsh("192.168.1.188:22", "r00t", "r00t")
	fmt.Printf("%t, %v\n", result, err)
}
