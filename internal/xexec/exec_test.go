package xexec

import (
	"context"
	"fmt"
	"testing"
)

func Test_It(t *testing.T) {
	opts := &Options{
		EntryPoint: "bun",
		WorkingDir: "./buntest",
		Args:       []string{"install"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stdoutCh := make(chan string)
	stderrCh := make(chan string)

	// do logs
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case stdout := <-stdoutCh:
				fmt.Printf("stdout: %s\n", stdout)
			case stderr := <-stderrCh:
				fmt.Printf("stderr: %s\n", stderr)
			}
		}
	}()

	cmd := NewCommand(opts, stdoutCh, stderrCh)
	if err := cmd.Run(ctx); err != nil {
		t.Error(err)
		return
	}
}
