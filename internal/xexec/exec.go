package xexec

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
)

type Options struct {
	// WorkingDir is the working directory for the command.
	WorkingDir string

	// EntryPoint is the entry point for the command.
	EntryPoint string

	// Args are the arguments for the command.
	Args []string

	// EnvironmentVariables are the environment variables for the command.
	EnvironmentVariables map[string]string
}

type Command interface {
	Run(ctx context.Context) error
}

type command struct {
	stdoutCh chan string
	stderrCh chan string

	opts *Options
}

func (c *command) stdout(pipe io.Reader) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		c.stdoutCh <- scanner.Text()
	}
}

func (c *command) stderr(pipe io.Reader) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		c.stderrCh <- scanner.Text()
	}
}

func (c *command) Run(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, c.opts.EntryPoint, c.opts.Args...)
	cmd.Dir = c.opts.WorkingDir
	cmd.Env = formatEnvVars(c.opts.EnvironmentVariables)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating stderr pipe: %w", err)
	}

	go c.stdout(stdoutPipe)
	go c.stderr(stderrPipe)
	return cmd.Run()
}

func NewCommand(opts *Options, stdoutCh chan string, stderrCh chan string) Command {
	return &command{
		stdoutCh: stdoutCh,
		stderrCh: stderrCh,
		opts:     opts,
	}
}
