package executors

import (
	"testing"

	"github.com/environment-toolkit/go-synth/models"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Test_copyDir(t *testing.T) {
	logger := getPrettyLogger()
	// create test filesystem
	files := []string{
		"foo/test.txt",
		"foo/subdir1/test.txt",
		"foo/subdir2/test.txt",
	}
	fs := afero.NewMemMapFs()
	for _, path := range files {
		err := afero.WriteFile(fs, path, []byte("Hello, World!"), 0644)
		require.NoError(t, err)
	}

	testCases := []struct {
		name          string
		fromDir       string
		toDir         string
		options       models.CopyOptions
		expectedFiles map[string]bool
	}{
		{
			name:    "Default should copy everything",
			fromDir: "foo",
			toDir:   "bar",
			options: models.CopyOptions{},
			expectedFiles: map[string]bool{
				"bar/test.txt":         true,
				"bar/subdir1/test.txt": true,
				"bar/subdir2/test.txt": true,
			},
		},
		{
			name:    "ignore all files except test.txt",
			fromDir: "foo",
			toDir:   "bar",
			options: models.CopyOptions{
				IgnorePatterns: []string{"*/*"},
				AllowPatterns:  []string{"test.txt"},
			},
			expectedFiles: map[string]bool{
				"bar/test.txt":         true,
				"bar/subdir1/test.txt": false,
				"bar/subdir2/test.txt": false,
			},
		},
		{
			name:    "skip directory",
			fromDir: "foo",
			toDir:   "bar",
			options: models.CopyOptions{
				SkipDirs: []string{"subdir1"},
			},
			expectedFiles: map[string]bool{
				"bar/test.txt":         true,
				"bar/subdir1/test.txt": false,
				"bar/subdir2/test.txt": true,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			destFs := afero.NewMemMapFs()
			err := copyDir(logger, tc.fromDir, tc.toDir, fs, destFs, tc.options)
			require.NoError(t, err)
			failed := false
			for path, shouldExist := range tc.expectedFiles {
				exists := fileExists(destFs, path)
				if exists != shouldExist {
					t.Logf("expected file %s: got %t - expected %t", path, exists, shouldExist)
					failed = true
				} else {
					t.Logf("expected file %s: got %t", path, exists)
				}
			}
			if failed {
				t.Fail()
			}
		})
	}
}

func fileExists(fs afero.Fs, path string) bool {
	_, err := fs.Stat(path)
	return err == nil
}

func getPrettyLogger() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Optional: colorize the log level
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder        // Optional: use ISO8601 time format
	config.Encoding = "console"
	logger, _ := config.Build()
	return logger
}
