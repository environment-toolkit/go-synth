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

// bunExecutor implements the Executor interface using bun.sh.
type bunExecutor struct {
	fs         afero.Fs
	workingDir string
	templates  *templateStore
	logger     *zap.Logger
}

// NewBunExecutor creates a new instance of BunExecutor.
func NewBunExecutor(logger *zap.Logger) (models.Executor, error) {
	fs, workingDir, e := newTempFs("go-synth-bun")
	if e != nil {
		return nil, fmt.Errorf("error creating Bun Exector temp fs: %w", e)
	}
	return &bunExecutor{
		logger:     logger,
		templates:  initializeTemplates(logger, "resources/bun"),
		fs:         fs,
		workingDir: workingDir,
	}, nil
}

func (be *bunExecutor) Setup(ctx context.Context, conf models.AppConfig, envVars map[string]string) error {
	merged := models.AppConfig{
		Dependencies: map[string]string{
			"cdktf": "^0.20.7",
		},
		DevDependencies: map[string]string{},
		Scopes:          []models.ScopedPackageOptions{},
	}
	maps.Copy(merged.Dependencies, conf.Dependencies)
	maps.Copy(merged.DevDependencies, conf.DevDependencies)

	if err := be.templates.setupFs(ctx, be.fs, merged); err != nil {
		return err
	}
	options := &xexec.Options{
		WorkingDir:           be.workingDir,
		EntryPoint:           "bun",
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

// Exec runs the main.ts script using bun.sh
func (be *bunExecutor) Exec(ctx context.Context, mainTS string, envVars map[string]string) error {
	if err := afero.WriteFile(be.fs, "main.ts", []byte(mainTS), 0775); err != nil {
		return err
	}
	options := &xexec.Options{
		WorkingDir:           be.workingDir,
		EntryPoint:           "bun",
		Args:                 []string{"run", "main.ts"},
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

func (be *bunExecutor) CopyTo(ctx context.Context, srcDir string, dstFs afero.Fs, dstDir string, opts models.CopyOptions) error {
	return copyDir(be.logger, srcDir, dstDir, be.fs, dstFs, opts)
}

func (be *bunExecutor) CopyFrom(ctx context.Context, srcFs afero.Fs, srcDir, dstDir string, opts models.CopyOptions) error {
	return copyDir(be.logger, srcDir, dstDir, srcFs, be.fs, opts)
}

func (be *bunExecutor) Cleanup(ctx context.Context) error {
	be.logger.Debug("Cleaning up Bun Executor")
	if err := be.fs.RemoveAll(be.workingDir); err != nil {
		return err
	}
	return nil
}
