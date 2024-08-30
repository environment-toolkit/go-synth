package executors

import (
	"context"
	"fmt"
	"maps"

	"github.com/environment-toolkit/go-synth/internal/xexec"
	"github.com/environment-toolkit/go-synth/models"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

// nodeExecutor implements the Executor interface using NodeJS and pnpm.
type nodeExecutor struct {
	fs         afero.Fs
	workingDir string
	templates  *templateStore
	logger     *zap.Logger
	entrypoint string
}

// NewNodeExecutor creates a new instance of nodeExecutor.
func NewNodeExecutor(logger *zap.Logger) (models.Executor, error) {
	fs, workingDir, e := newTempFs("go-synth-node")
	if e != nil {
		return nil, fmt.Errorf("error creating Node Exector temp fs: %w", e)
	}
	return &nodeExecutor{
		logger:     logger,
		templates:  initializeTemplates(logger, "resources/node"),
		fs:         fs,
		workingDir: workingDir,
	}, nil
}

func (be *nodeExecutor) Setup(ctx context.Context, conf models.AppConfig, envVars map[string]string) error {
	merged := models.AppConfig{
		Dependencies: map[string]string{
			"cdktf": "^0.20.7",
		},
		DevDependencies: map[string]string{
			"typescript": "5.4.5",
			"ts-node":    "^10.9.2",
			"@swc/core":  "^1.7.6",
		},
		ExecutorOptions: map[string]string{
			"nodeVersion":    ">=18.0.0",
			"packageManager": "pnpm@9.0.2",
			"entrypoint":     "pnpm",
			"synthScript":    "ts-node --swc -P ./tsconfig.json main.ts",
		},
		Scopes: []models.ScopedPackageOptions{},
	}
	maps.Copy(merged.Dependencies, conf.Dependencies)
	maps.Copy(merged.DevDependencies, conf.DevDependencies)
	maps.Copy(merged.ExecutorOptions, conf.ExecutorOptions)
	be.entrypoint = merged.ExecutorOptions["entrypoint"]

	if err := be.templates.setupFs(ctx, be.fs, merged); err != nil {
		return err
	}

	options := &xexec.Options{
		WorkingDir:           be.workingDir,
		EntryPoint:           be.entrypoint,
		Args:                 []string{"install"},
		EnvironmentVariables: envVars,
	}
	stdoutCh := make(chan string)
	stderrCh := make(chan string)
	defer func() {
		close(stdoutCh)
		close(stderrCh)
	}()

	cmd := xexec.NewCommand(options, stdoutCh, stderrCh)
	if err := cmd.Run(ctx); err != nil {
		return fmt.Errorf("error running bun install: %w", err)
	}
	return nil
}

// Exec runs the main.ts script using ts-node.
func (be *nodeExecutor) Exec(ctx context.Context, mainTS string, envVars map[string]string) error {
	if err := afero.WriteFile(be.fs, "main.ts", []byte(mainTS), 0775); err != nil {
		return err
	}

	options := &xexec.Options{
		WorkingDir:           be.workingDir,
		EntryPoint:           be.entrypoint,
		Args:                 []string{"run", "synth"},
		EnvironmentVariables: envVars,
	}
	stdoutCh := make(chan string)
	stderrCh := make(chan string)
	defer func() {
		close(stdoutCh)
		close(stderrCh)
	}()

	cmd := xexec.NewCommand(options, stdoutCh, stderrCh)
	if err := cmd.Run(ctx); err != nil {
		return fmt.Errorf("error running bun install: %w", err)
	}
	return nil
}

func (be *nodeExecutor) CopyTo(ctx context.Context, srcDir string, dstFs afero.Fs, dstDir string, opts models.CopyOptions) error {
	return copyDir(be.logger, srcDir, dstDir, be.fs, dstFs, opts)
}

func (be *nodeExecutor) CopyFrom(ctx context.Context, srcFs afero.Fs, srcDir, dstDir string, opts models.CopyOptions) error {
	return copyDir(be.logger, srcDir, dstDir, srcFs, be.fs, opts)
}

func (be *nodeExecutor) Cleanup(ctx context.Context) error {
	be.logger.Debug("Cleaning up Node Executor")
	if err := be.fs.RemoveAll(be.workingDir); err != nil {
		return err
	}
	return nil
}
