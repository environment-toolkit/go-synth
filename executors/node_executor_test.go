package executors

import (
	"context"
	"testing"

	"github.com/environment-toolkit/go-synth/config"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Test_nodeExecutor_Setup(t *testing.T) {
	be := getTestNodeExecutor()
	defer be.Cleanup(context.Background())

	err := be.Setup(context.Background(), config.App{}, nil)
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

	fixtureFs := afero.NewBasePathFs(afero.NewOsFs(), "../fixtures")
	// copy in local package
	if err := be.CopyFrom(ctx, "cdktf-lib", "./fixtures/cdktf-lib", fixtureFs); err != nil {
		t.Fatalf("CopyFrom failed: %v", err)
	}
	err := be.Setup(ctx, config.App{
		Dependencies: map[string]string{
			"cdktf-lib": "./fixtures/cdktf-lib",
		},
		ExecutorOptions: map[string]string{
			// required for pnpm to install cdktf-lib dependencies ...
			"pnpmWorkspace": "packages:\n- \"./fixtures/cdktf-lib\"",
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

	snapshotFs(t, "node_fixtures", "cdktf.out", be.fs)
}

func getTestNodeExecutor() *nodeExecutor {
	// Create a custom logger configuration
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Optional: colorize the log level
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder        // Optional: use ISO8601 time format
	config.Encoding = "console"
	logger, _ := config.Build()
	be, _ := NewNodeExecutor(logger)
	return be.(*nodeExecutor)
}
