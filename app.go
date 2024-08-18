package synth

import (
	"context"
	"os"

	"github.com/environment-toolkit/go-synth/auth"
	"github.com/environment-toolkit/go-synth/executors"
	"github.com/environment-toolkit/go-synth/models"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

// App defines the interface for managing the synthesis process.
type App interface {
	// Configure is a one time set up for the App environment reused by each Eval call.
	//
	// Configure is meant to handle Auth configuration and other setup that is shared across multiple Eval calls.
	Configure(ctx context.Context, config models.AppConfig) error
	// Eval runs the provided main.ts script in the App environment.
	//
	// Once the script has run, the contents of the src directory are
	// copied to the dest directory into the provided fs.
	//
	// Each call to Eval is independent.
	Eval(ctx context.Context, fs afero.Fs, mainTs, src, dest string) error
}

type app struct {
	config        models.AppConfig
	newExecutorFn models.NewExecutorFn
	authProvider  auth.Provider
	envVars       map[string]string
	logger        *zap.Logger
}

func NewApp(newFn models.NewExecutorFn, logger *zap.Logger) App {
	return &app{
		newExecutorFn: newFn,
		authProvider:  auth.NewAuthProvider(),
		logger:        logger,
	}
}

func (a *app) Configure(ctx context.Context, config models.AppConfig) error {
	envVars := config.EnvVars
	if envVars == nil {
		a.logger.Debug("using os env vars")
		envVars = executors.EnvMap(os.Environ())
	}
	a.envVars = envVars
	a.config = config
	for _, scopedPackage := range a.config.Scopes {
		if !scopedPackage.RequiresAuth {
			continue
		}
		authenticator, err := a.authProvider.Provide(ctx, scopedPackage.RegistryURL)
		if err != nil {
			return err
		}
		a.envVars, err = authenticator.Auth(ctx, *scopedPackage.AuthTokenEnvVar, a.envVars)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *app) Eval(ctx context.Context, dstFs afero.Fs, mainTs, src, dstPath string) error {
	e, err := a.newExecutorFn(a.logger)
	if err != nil {
		return err
	}
	defer e.Cleanup(ctx)
	if a.config.PreSetupFn != nil {
		if err := a.config.PreSetupFn(e); err != nil {
			return err
		}
	}
	if err := e.Setup(ctx, a.config, a.envVars); err != nil {
		return err
	}
	if err := e.Exec(ctx, mainTs, a.envVars); err != nil {
		return err
	}
	if err := e.CopyTo(ctx, src, dstFs, dstPath, models.CopyOptions{}); err != nil {
		return err
	}
	return nil
}
