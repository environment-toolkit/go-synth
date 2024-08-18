package models

import (
	"context"

	"github.com/spf13/afero"
	"go.uber.org/zap"
)

// NewExecutorFn returns an Executor
type NewExecutorFn func(logger *zap.Logger) (Executor, error)

// Executor defines the interface for executing the synthesis process.
type Executor interface {
	// Setup configures the executor based on the provided AppConfig and environment variables.
	Setup(ctx context.Context, config AppConfig, envVars map[string]string) error

	// Exec executes the provided main.ts script with the given environment variables.
	Exec(ctx context.Context, mainTS string, envVars map[string]string) error

	// CopyTo retrieves the result from the source path and copies it to the destination within the provided filesystem.
	CopyTo(ctx context.Context, srcDir string, dstFS afero.Fs, dstDir string, options CopyOptions) error

	// CopyFrom copies the source path to the executor workingDir from the provided filesystem.
	CopyFrom(ctx context.Context, srcFS afero.Fs, srcDir, dstDir string, options CopyOptions) error

	// Cleanup cleans up the environment.
	Cleanup(ctx context.Context) error
}

type CopyOptions struct {
	// SkipDirs is a list of directories to skip.
	//
	// If a directory is skipped, all its contents will be skipped as well including AllowPatterns.
	SkipDirs []string
	// AllowPatterns is a list of patterns to allow, SkipDirs will still be respected.
	AllowPatterns []string
	// IgnorePatterns is a list of patterns to ignore. Unless they were allowed by AllowPatterns.
	//
	// Note that golang does not support `**` for recursive matching.
	//
	// See: https://github.com/golang/go/issues/11862
	IgnorePatterns []string
}
