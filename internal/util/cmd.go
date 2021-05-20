package util

import (
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
