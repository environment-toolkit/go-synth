package executors

import (
	"context"
	"os"
	"testing"

	"github.com/environment-toolkit/go-synth/models"
	"github.com/spf13/afero"
)

func Test_nodeExecutor_Setup(t *testing.T) {
	be := getTestNodeExecutor()
	defer be.Cleanup(context.Background())

	err := be.Setup(context.Background(), models.AppConfig{}, nil)
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

	snapshotFs(t, "node_fixtures", "cdktf.out", be.fs)
}

func getTestNodeExecutor() *nodeExecutor {
	be, _ := NewNodeExecutor(getPrettyLogger())
	return be.(*nodeExecutor)
}
