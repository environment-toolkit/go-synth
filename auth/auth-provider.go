package auth

import (
	"context"
	"fmt"
)

type Authenticator interface {
	// Auth returns envVars map with the authentication token set.
	Auth(ctx context.Context, envKey string, envVars map[string]string) (map[string]string, error)
}

type Provider interface {
	Provide(ctx context.Context, registryUrl string) (Authenticator, error)
}

type provider struct {
	// authenticators is a map of registry URL to Authenticator.
	authenticators map[string]Authenticator
}

func NewAuthProvider() Provider {
	return &provider{
		authenticators: make(map[string]Authenticator),
	}
}

func (ap *provider) Provide(ctx context.Context, registryUrl string) (Authenticator, error) {
	if authenticator, ok := ap.authenticators[registryUrl]; ok {
		return authenticator, nil
	}
	authenticator, err := ap.newAuthenticator(ctx, registryUrl)
	if err != nil {
		return nil, err
	}
	ap.authenticators[registryUrl] = authenticator
	return authenticator, nil
}

func (ap *provider) newAuthenticator(ctx context.Context, registryUrl string) (Authenticator, error) {
	if IsCodeArtifactURL(registryUrl) {
		return NewCodeArtifact(ctx, registryUrl)
	}
	return nil, fmt.Errorf("unsupported registry URL: %s", registryUrl)
}
