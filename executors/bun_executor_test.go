package executors

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/terraconstructs/go-synth/models"
)

func Test_bunExecutor_BasicExec(t *testing.T) {
	be := getTestBunExecutor()
	defer be.Cleanup(context.Background())

	// Create a test file
	const testFile = "file.txt"
	mainTS := fmt.Sprintf(`await Bun.write(%q, "Lorem ipsum");`, testFile)

	envVars := EnvMap(os.Environ())
	err := be.Exec(context.Background(), mainTS, envVars)
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

	envVars := EnvMap(os.Environ())
	err := be.Setup(context.Background(), models.AppConfig{}, envVars)
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
	envVars := EnvMap(os.Environ())

	fixtureFs := afero.NewBasePathFs(afero.NewOsFs(), "../fixtures")
	if err := be.CopyFrom(ctx, fixtureFs, "cdktf-lib", "./fixtures/cdktf-lib", models.CopyOptions{
		SkipDirs: []string{"node_modules"},
	}); err != nil {
		t.Fatalf("CopyFrom failed: %v", err)
	}

	// copy in local file
	if err := be.CopyFileFrom(ctx, fixtureFs, "testfile", "testfile"); err != nil {
		t.Fatalf("CopyFileFrom failed: %v", err)
	}

	err := be.Setup(ctx, models.AppConfig{
		Dependencies: map[string]string{
			"cdktf-lib": "./fixtures/cdktf-lib",
		},
	}, envVars)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	var mainTs []byte
	if mainTs, err = afero.ReadFile(fixtureFs, "local-package/main.ts"); err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if err := be.Exec(ctx, string(mainTs), envVars); err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	// assert the testfile exists in  the be.fs
	exists, err := afero.Exists(be.fs, "testfile")
	if err != nil {
		t.Fatalf("Failed to check if testfile exists: %v", err)
	}
	if !exists {
		t.Errorf("testfile was not created")
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
	be, _ := NewBunExecutor(getPrettyLogger())
	return be.(*bunExecutor)
}
