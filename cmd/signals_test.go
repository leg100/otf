package cmd

import (
	"context"
	"sync"
	"syscall"
	"testing"
)

func TestCatchCtrlC(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())
	CatchCtrlC(cancel)

	go func() {
		<-ctx.Done()
		wg.Done()
	}()

	syscall.Kill(syscall.Getpid(), syscall.SIGINT)

	wg.Wait()
}
