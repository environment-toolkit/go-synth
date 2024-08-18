package models

type AppConfig struct {
	DevDependencies map[string]string      // DevDependencies
	Dependencies    map[string]string      // Dependencies
	Scopes          []ScopedPackageOptions // Options for package scopes
	ExecutorOptions map[string]string      // Options for the executor
	PreSetupFn      func(e Executor) error // Function to run before setup
	EnvVars         map[string]string      // Environment variables to set
}

type ScopedPackageOptions struct {
	// Scape of package
	Scope string

	// URL of the registry for the scoped packages
	RegistryURL string

	// Whether the registry requires authentication
	RequiresAuth bool

	// Env var to pass Auth token to bun install
	AuthTokenEnvVar *string //TODO validate auth token env var is unique
}
