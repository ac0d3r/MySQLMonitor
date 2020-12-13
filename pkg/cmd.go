package pkg

import (
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
