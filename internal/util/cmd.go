package util

import (
	"bufio"
	"os"
	"os/exec"
)

// ExecCMD 运行 cmd，返回结果
func ExecCMD(cmd string) (string, error) {
	var (
		out []byte
		err error
	)
	if out, err = exec.Command("bash", "-c", cmd).Output(); err != nil {
		return "", err
	}
	return string(out), err
}

// RunExeInOrphanProcess 在孤儿进程中运行可执行文件
func RunExeInOrphanProcess(exefile string) {
	var (
		attr *os.ProcAttr
		args []string
		err  error
	)
	attr = &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Env:   os.Environ(),
	}
	args = []string{"/bin/bash", "-c", exefile}
	if _, err = os.StartProcess("/bin/bash", args, attr); err != nil {
		os.Stderr.Write([]byte(err.Error()))
	}
}

type HandlerFunc func(string)

func StartCmd(cmd *exec.Cmd, outFunc, errFunc HandlerFunc, deferFunc func()) error {
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	bufOut := bufio.NewReader(outPipe)
	bufErr := bufio.NewReader(errPipe)
	go func() {
		defer deferFunc()
		for {
			line, _, err := bufOut.ReadLine()
			if err != nil {
				break
			}
			outFunc(string(line))
		}
	}()
	go func() {
		for {
			line, _, err := bufErr.ReadLine()
			if err != nil {
				break
			}
			errFunc(string(line))
		}
	}()
	return nil
}
