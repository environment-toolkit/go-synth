package executors

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/afero"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// removeExtension removes the extension from the provided path.
func removeExtension(path string) string {
	return strings.TrimSuffix(path, filepath.Ext(path))
}

// ensurePath creates the directory structure for the provided path.
func ensurePath(dest afero.Fs, path string) error {
	dir, _ := filepath.Split(path)
	ospath := filepath.FromSlash(dir)
	err := dest.MkdirAll(ospath, 0775)
	if err != nil {
		return fmt.Errorf("couldn't create %s directory, %w", dir, err)
	}
	return nil
}

func copyDir(logger *zap.Logger, srcDir, destDir string, src, dest afero.Fs) error {
	srcDirInfo, err := src.Stat(srcDir)
	if err != nil {
		return err
	}
	if !srcDirInfo.IsDir() {
		return fmt.Errorf("source path is not a directory: %s", srcDir)
	}

	afero.Walk(src, srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip node_modules
		if info.IsDir() && info.Name() == "node_modules" {
			return filepath.SkipDir
		}
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(destDir, relPath)
		if info.IsDir() {
			if err := dest.MkdirAll(destPath, 0755); err != nil {
				return err
			}
			return nil
		}
		logger.Info("copying file", zap.String("src", path), zap.String("dest", destPath))
		if err := copyFile(src, dest, path, destPath); err != nil {
			return err
		}
		return nil
	})
	return nil
}

// copyFile copies a file from the source filesystem to the destination filesystem.
func copyFile(src, dest afero.Fs, srcPath, destPath string) error {
	srcFile, err := src.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := dest.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return err
	}
	return nil
}

// newTempFs creates a new temporary filesystem with the provided pattern.
func newTempFs(pattern string) (afero.Fs, string, error) {
	d, err := os.MkdirTemp("", pattern)
	if err != nil {
		return nil, "", err
	}
	fs := afero.NewBasePathFs(afero.NewOsFs(), d)
	return fs, d, nil
}

type runCommandOptions struct {
	workingDir string
	entrypoint string
	envVars    map[string]string
	logger     *zap.Logger
}

// runCommand runs the specified entrypoint with the provided environment variables.
func runCommand(ctx context.Context, options *runCommandOptions, args ...string) error {
	envVars := options.envVars
	if envVars == nil {
		envVars = envMap(os.Environ())
	}

	// check for early cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	cmd := exec.CommandContext(ctx, options.entrypoint, args...)
	cmd.Dir = options.workingDir
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

	var wg sync.WaitGroup
	wg.Add(1)
	go streamOutput(options.logger, &wg, stdoutPipe, zap.InfoLevel)
	wg.Add(1)
	go streamOutput(options.logger, &wg, stderrPipe, zap.WarnLevel)

	// Reads from pipes must be completed before calling
	// cmd.Wait() to prevent race condition
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		// dump env for debug
		for k, v := range envVars {
			options.logger.Debug("environment", zap.String(k, v))
		}
		return fmt.Errorf("error running %s %s: %w", options.entrypoint, args, err)
	}

	return nil
}

// streamOutput reads from the provided pipe and logs the output using the provided logger.
func streamOutput(logger *zap.Logger, wg *sync.WaitGroup, pipe io.ReadCloser, level zapcore.Level) {
	defer wg.Done()
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		logger.Check(level, scanner.Text()).Write()
	}
	if err := scanner.Err(); err != nil {
		logger.Error("error reading from pipe", zap.Error(err))
	}
}

// Format environment variables for command execution.
func formatEnvVars(envVars map[string]string) []string {
	formatted := make([]string, 0, len(envVars))
	for k, v := range envVars {
		formatted = append(formatted, fmt.Sprintf("%s=%s", k, v))
	}
	return formatted
}

// envMap converts os.environ output to a map.
func envMap(environ []string) map[string]string {
	env := map[string]string{}
	for _, ev := range environ {
		parts := strings.SplitN(ev, "=", 2)
		if len(parts) == 0 {
			continue
		}
		k := parts[0]
		v := ""
		if len(parts) == 2 {
			v = parts[1]
		}
		env[k] = v
	}
	return env
}
