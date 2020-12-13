package pkg

import (
	"os"
	"os/signal"
)

// RegisterSignal 注册系统信号
func RegisterSignal(sigs ...os.Signal) <-chan os.Signal {
	var (
		sig chan os.Signal
	)
	sig = make(chan os.Signal, 1)
	signal.Notify(sig, sigs...)
	return sig
}
