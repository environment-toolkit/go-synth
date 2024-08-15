package executors

import (
	"context"
	"fmt"
	"maps"
	"os/exec"

	"github.com/environment-toolkit/go-synth/config"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

// bunExecutor implements the Executor interface using bun.sh.
type bunExecutor struct {
	fs         afero.Fs
	workingDir string
	templates  *templateStore
	logger     *zap.Logger
}

// NewBunExecutor creates a new instance of BunExecutor.
func NewBunExecutor(logger *zap.Logger) (Executor, error) {
	fs, workingDir, e := newTempFs("go-synth-bun")
	if e != nil {
		return nil, fmt.Errorf("error creating Bun Exector temp fs: %w", e)
	}
	return &bunExecutor{
		logger:     logger,
		templates:  initializeTemplates(logger),
		fs:         fs,
		workingDir: workingDir,
	}, nil
}

func (be *bunExecutor) Setup(ctx context.Context, conf config.App, envVars map[string]string) error {
	merged := config.App{
		Dependencies: map[string]string{
			"cdktf": "^0.20.7",
		},
		DevDependencies: map[string]string{
			// "@types/bun": "^1.1.3",
			// "typescript": "5.4.5",
		},
		ScopedPackages: []config.ScopedPackageOptions{},
	}
	maps.Copy(merged.Dependencies, conf.Dependencies)
	maps.Copy(merged.DevDependencies, conf.DevDependencies)

	if err := be.templates.setupFs(ctx, be.fs, merged); err != nil {
		return err
	}
	if err := be.runBunCommand(ctx, envVars, "install"); err != nil {
		return fmt.Errorf("error running bun install: %w", err)
	}
	return nil
}

// Exec runs the main.ts script using bun.sh
func (be *bunExecutor) Exec(ctx context.Context, mainTS string, envVars map[string]string) error {
	if err := afero.WriteFile(be.fs, "main.ts", []byte(mainTS), 0775); err != nil {
		return err
	}
	if err := be.runBunCommand(ctx, envVars, "run", "main.ts"); err != nil {
		return fmt.Errorf("error running bun install: %w", err)
	}
	return nil
}
func (be *bunExecutor) CopyTo(ctx context.Context, srcDir, destDir string, dest afero.Fs) error {
	return copyDir(be.logger, srcDir, destDir, be.fs, dest)
}

func (be *bunExecutor) CopyFrom(ctx context.Context, srcDir, destDir string, src afero.Fs) error {
	return copyDir(be.logger, srcDir, destDir, src, be.fs)
}

func (be *bunExecutor) Cleanup(ctx context.Context) error {
	// clean up temp directory
	if err := be.fs.RemoveAll(be.workingDir); err != nil {
		return err
	}
	return nil
}

// runBunCommand runs the specified bun.sh command with the provided environment variables.
func (be *bunExecutor) runBunCommand(ctx context.Context, envVars map[string]string, args ...string) error {
	cmd := exec.CommandContext(ctx, "bun", args...)
	cmd.Dir = be.workingDir
	cmd.Env = formatEnvVars(envVars)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %w", err)
	}

	go streamOutput(be.logger, stdoutPipe, zap.InfoLevel)
	go streamOutput(be.logger, stderrPipe, zap.WarnLevel)

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("error running bun %s: %w", args, err)
	}

	return nil
}
