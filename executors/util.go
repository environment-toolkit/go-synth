package executors

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

// Format environment variables for command execution.
func formatEnvVars(envVars map[string]string) []string {
	formatted := make([]string, 0, len(envVars))
	for k, v := range envVars {
		formatted = append(formatted, fmt.Sprintf("%s=%s", k, v))
	}
	return formatted
}

// streamOutput reads from the provided pipe and logs the output using the provided logger.
func streamOutput(logger *zap.Logger, pipe io.ReadCloser, level zapcore.Level) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		logger.Check(level, scanner.Text()).Write()
	}
	if err := scanner.Err(); err != nil {
		logger.Error("error reading from pipe", zap.Error(err))
	}
}
