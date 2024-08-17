package executors

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/environment-toolkit/go-synth/config"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Test_bunExecutor_BasicExec(t *testing.T) {
	be := getTestBunExecutor()
	defer be.Cleanup(context.Background())

	// Create a test file
	const testFile = "file.txt"
	mainTS := fmt.Sprintf(`await Bun.write(%q, "Lorem ipsum");`, testFile)

	err := be.Exec(context.Background(), mainTS, nil)
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	exists, err := afero.Exists(be.fs, testFile)
	if err != nil {
		t.Fatalf("Failed to check if file exists: %v", err)
	}
	if !exists {
		t.Errorf("main.ts file was not created")
	}
}

func Test_bunExecutor_Setup(t *testing.T) {
	be := getTestBunExecutor()
	defer be.Cleanup(context.Background())

	err := be.Setup(context.Background(), config.App{}, nil)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	snapshotFs(t, "bun_setup", ".", be.fs)
}

func Test_bunExecutor_Fixtures(t *testing.T) {
	ctx := context.Background()
	be := getTestBunExecutor()
	defer be.Cleanup(ctx)
	be.logger.Info(be.workingDir)

	fixtureFs := afero.NewBasePathFs(afero.NewOsFs(), "../fixtures")
	if err := be.CopyFrom(ctx, "cdktf-lib", "./fixtures/cdktf-lib", fixtureFs); err != nil {
		t.Fatalf("CopyFrom failed: %v", err)
	}

	err := be.Setup(ctx, config.App{
		Dependencies: map[string]string{
			"cdktf-lib": "./fixtures/cdktf-lib",
		},
	}, nil)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	var mainTs []byte
	if mainTs, err = afero.ReadFile(fixtureFs, "local-package/main.ts"); err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if err := be.Exec(ctx, string(mainTs), nil); err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	snapshotFs(t, "bun_fixtures", "cdktf.out", be.fs)
}

// takeSnapshot reads the directory contents and returns a map of paths and file contents
func snapshotFs(t *testing.T, name, root string, fs afero.Fs) error {
	snapsConf := snaps.WithConfig(snaps.Filename(name))
	// fsMap := make(map[string]string)
	err := afero.Walk(fs, root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip node_modules
		if info.IsDir() && info.Name() == "node_modules" {
			return filepath.SkipDir
		}
		if info.IsDir() || info.Name() == "bun.lockb" || info.Name() == "pnpm-lock.yaml" {
			return nil
		}
		content, err := afero.ReadFile(fs, path)
		if err != nil {
			return err
		}
		snapsConf.MatchSnapshot(t, path, string(content))
		// fsMap[path] = string(content)
		return nil
	})
	require.NoError(t, err)
	// convert the map to string

	return nil
}

func getTestBunExecutor() *bunExecutor {
	// Create a custom logger configuration
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Optional: colorize the log level
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder        // Optional: use ISO8601 time format
	config.Encoding = "console"
	logger, _ := config.Build()
	be, _ := NewBunExecutor(logger)
	return be.(*bunExecutor)
}
