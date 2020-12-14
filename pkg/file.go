package pkg

import (
	"fmt"
	"os"
)

// PathExist 判断文件是否存在
func PathExist(_path string) bool {
	var err error
	if _, err = os.Stat(_path); err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

// CopyFile 硬链接方式拷贝文件
func CopyFile(oldname, newname string) error {
	if !PathExist(oldname) {
		return fmt.Errorf("Not Found %s", oldname)
	}
	if PathExist(newname) {
		return fmt.Errorf("Already exists %s", newname)
	}
	return os.Link(oldname, newname)
}
