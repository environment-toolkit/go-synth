package synth

import (
	"context"

	"github.com/environment-toolkit/go-synth/auth"
	"github.com/environment-toolkit/go-synth/config"
	"github.com/environment-toolkit/go-synth/executors"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

// App defines the interface for managing the synthesis process.
type App interface {
	// Configure is a one time set up for the App environment reused by each Eval call.
	Configure(ctx context.Context, config config.App) error
	// Eval runs the provided main.ts script in the App environment. Each call to Eval is independent.
	Eval(ctx context.Context, mainTs, src, dest string, fs afero.Fs) error
}

type app struct {
	config        config.App
	newExecutorFn executors.NewFn
	authProvider  auth.Provider
	envVars       map[string]string
	logger        *zap.Logger
}

func NewApp(newFn executors.NewFn, logger *zap.Logger) App {
	return &app{
		newExecutorFn: newFn,
		authProvider:  auth.NewAuthProvider(),
		envVars:       make(map[string]string),
		logger:        logger,
	}
}

func (a *app) Configure(ctx context.Context, config config.App) error {
	a.config = config
	for _, scopedPackage := range a.config.ScopedPackages {
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

func (a *app) Eval(ctx context.Context, mainTs, src, dest string, fs afero.Fs) error {
	e, err := a.newExecutorFn(a.logger)
	if err != nil {
		return err
	}
	defer e.Cleanup(ctx)
	if err := e.Setup(ctx, a.config, a.envVars); err != nil {
		return err
	}
	if err := e.Exec(ctx, mainTs, a.envVars); err != nil {
		return err
	}
	if err := e.CopyTo(ctx, src, dest, fs); err != nil {
		return err
	}
	return nil
}
