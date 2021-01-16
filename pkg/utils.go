package pkg

import "os"

func IsRoot() bool {
	return os.Geteuid() == 0
}
