package executors

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/terraconstructs/go-synth/models"
)

func Test_nodeExecutor_Setup(t *testing.T) {
	be := getTestNodeExecutor()
	defer be.Cleanup(context.Background())

	envVars := EnvMap(os.Environ())
	err := be.Setup(context.Background(), models.AppConfig{}, envVars)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	snapshotFs(t, "node_setup", ".", be.fs)
}

func Test_nodeExecutor_Fixtures(t *testing.T) {
	ctx := context.Background()
	be := getTestNodeExecutor()
	defer be.Cleanup(ctx)
	be.logger.Info(be.workingDir)

	envVars := EnvMap(os.Environ())

	fixtureFs := afero.NewBasePathFs(afero.NewOsFs(), "../fixtures")
	// copy in local package
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
		ExecutorOptions: map[string]string{
			// required for pnpm to install cdktf-lib dependencies ...
			"pnpmWorkspace": "packages:\n- \"./fixtures/cdktf-lib\"",
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
	snapshotFs(t, "node_fixtures", "cdktf.out", be.fs)
}

func getTestNodeExecutor() *nodeExecutor {
	be, _ := NewNodeExecutor(getPrettyLogger())
	return be.(*nodeExecutor)
}
