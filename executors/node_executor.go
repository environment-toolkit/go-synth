package executors

import (
	"context"
	"fmt"
	"maps"

	"github.com/environment-toolkit/go-synth/config"
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
func NewNodeExecutor(logger *zap.Logger) (Executor, error) {
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

func (be *nodeExecutor) Setup(ctx context.Context, conf config.App, envVars map[string]string) error {
	merged := config.App{
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
		Scopes: []config.ScopedPackageOptions{},
	}
	maps.Copy(merged.Dependencies, conf.Dependencies)
	maps.Copy(merged.DevDependencies, conf.DevDependencies)
	maps.Copy(merged.ExecutorOptions, conf.ExecutorOptions)
	be.entrypoint = merged.ExecutorOptions["entrypoint"]

	if err := be.templates.setupFs(ctx, be.fs, merged); err != nil {
		return err
	}

	options := &runCommandOptions{
		workingDir: be.workingDir,
		entrypoint: be.entrypoint,
		envVars:    envVars,
		logger:     be.logger,
	}
	if err := runCommand(ctx, options, "install"); err != nil {
		return fmt.Errorf("error running %s install: %w", be.entrypoint, err)
	}
	return nil
}

// Exec runs the main.ts script using ts-node.
func (be *nodeExecutor) Exec(ctx context.Context, mainTS string, envVars map[string]string) error {
	if err := afero.WriteFile(be.fs, "main.ts", []byte(mainTS), 0775); err != nil {
		return err
	}
	options := &runCommandOptions{
		workingDir: be.workingDir,
		entrypoint: be.entrypoint,
		envVars:    envVars,
		logger:     be.logger,
	}
	if err := runCommand(ctx, options, "run", "synth"); err != nil {
		return fmt.Errorf("error running synthScript: %w", err)
	}
	return nil
}

func (be *nodeExecutor) CopyTo(ctx context.Context, srcDir, destDir string, dest afero.Fs) error {
	return copyDir(be.logger, srcDir, destDir, be.fs, dest)
}

func (be *nodeExecutor) CopyFrom(ctx context.Context, srcDir, destDir string, src afero.Fs) error {
	return copyDir(be.logger, srcDir, destDir, src, be.fs)
}

func (be *nodeExecutor) Cleanup(ctx context.Context) error {
	be.logger.Debug("Cleaning up Node Executor")
	if err := be.fs.RemoveAll(be.workingDir); err != nil {
		return err
	}
	return nil
}
