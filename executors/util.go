package executors

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/environment-toolkit/go-synth/models"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

// removeExtension removes the extension from the provided path.
func removeExtension(path string) string {
	return strings.TrimSuffix(path, filepath.Ext(path))
}

// copyDir copies a directory from the source filesystem to the destination filesystem.
func copyDir(logger *zap.Logger, srcDir, destDir string, src, dest afero.Fs, options models.CopyOptions) error {
	srcDirInfo, err := src.Stat(srcDir)
	if err != nil {
		return err
	}
	if !srcDirInfo.IsDir() {
		return fmt.Errorf("source path is not a directory: %s", srcDir)
	}

	err = afero.Walk(src, srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			for _, skipDir := range options.SkipDirs {
				if relPath == skipDir {
					logger.Debug("skipping directory", zap.String("path", relPath))
					return filepath.SkipDir
				}
			}
			return nil
		}
		shouldIgnore, err := shouldIgnore(relPath, options.AllowPatterns, options.IgnorePatterns)
		if err != nil {
			return err
		}
		if shouldIgnore {
			logger.Debug("ignoring file", zap.String("path", path))
			return nil
		}
		destPath := filepath.Join(destDir, relPath)
		logger.Debug("copying file", zap.String("src", path), zap.String("dest", destPath))
		if err := copyFile(src, dest, path, destPath); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// shouldIgnore checks if the provided path should be ignored based on the allow and ignore patterns.
func shouldIgnore(path string, allowPatterns []string, ignorePatterns []string) (bool, error) {
	for _, pattern := range allowPatterns {
		match, err := filepath.Match(pattern, path)
		if err != nil {
			return false, err
		}
		if match {
			return false, nil
		}
	}

	for _, pattern := range ignorePatterns {
		match, err := filepath.Match(pattern, path)
		if err != nil {
			return false, err
		}
		if match {
			return true, nil
		}
	}

	return false, nil
}

// copyFile copies a file from the source filesystem to the destination filesystem.
func copyFile(src, dest afero.Fs, srcPath, destPath string) error {
	srcFile, err := src.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	if err := ensurePath(dest, destPath); err != nil {
		return err
	}
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

// newTempFs creates a new temporary filesystem with the provided pattern.
func newTempFs(pattern string) (afero.Fs, string, error) {
	d, err := os.MkdirTemp("", pattern)
	if err != nil {
		return nil, "", err
	}
	fs := afero.NewBasePathFs(afero.NewOsFs(), d)
	return fs, d, nil
}

// EnvMap converts os.environ output to a map.
func EnvMap(environ []string) map[string]string {
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
