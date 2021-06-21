package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func CatchCtrlC(cancel context.CancelFunc) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals,
		syscall.SIGTERM,
		syscall.SIGINT,
	)

	go func() {
		<-signals
		signal.Stop(signals)
		cancel()
	}()
}
