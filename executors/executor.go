package executors

import (
	"context"

	"github.com/environment-toolkit/go-synth/config"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

// NewFn returns an Executor
type NewFn func(logger *zap.Logger) (Executor, error)

// Executor defines the interface for executing the synthesis process.
type Executor interface {
	// Setup configures the executor based on the provided AppConfig and environment variables.
	Setup(ctx context.Context, config config.App, envVars map[string]string) error

	// Exec executes the provided main.ts script with the given environment variables.
	Exec(ctx context.Context, mainTS string, envVars map[string]string) error

	// CopyTo retrieves the result from the source path and copies it to the destination within the provided filesystem.
	CopyTo(ctx context.Context, srcPath, dstPath string, destFS afero.Fs) error

	// CopyFrom copies the source path to the executor workingDir from the provided filesystem.
	CopyFrom(ctx context.Context, srcPath, dstPath string, srcFS afero.Fs) error

	// Cleanup cleans up the environment.
	Cleanup(ctx context.Context) error
}
