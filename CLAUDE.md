# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go-synth is a Go library that dynamically synthesizes CDKTF (Cloud Development Kit for Terraform) constructs. It executes TypeScript CDKTF applications from Go code, handling setup, execution, and result extraction. This solves the limitation where JSII cross-compilation isn't sufficient for dynamic CDKTF construct loading in Go environments.

## Development Commands

### Building
- Build the main library: `go build`
- Build CLI example: `go build -o synth ./cmd`
- Embed Bun binary (Linux only): `make synth/bun`

### Testing
- Run all tests: `make test` or `go test ./...`
- Update test snapshots: `UPDATE_SNAPS=true go test ./...`
- The project uses `go-snaps` for snapshot testing and regression detection

### CLI Example Usage
```bash
./synth -file example/network.ts -deps "terraconstructs:0.0.8" -src "cdktf.out/stacks/network-stack" -out result/network
```

## Core Architecture

### Main Components

**App Interface** (`app.go`)
- Primary orchestrator with `Configure()` and `Eval()` methods
- Manages executor lifecycle and authentication
- Supports debug mode for development

**Executor Interface** (`models/executor.go`)
- Common interface for different runtime environments
- Key methods: Setup, Exec, CopyTo/CopyFrom, Cleanup
- Two implementations: BunExecutor (fast) and NodeExecutor (configurable)

**Executors** (`executors/`)
- **BunExecutor**: Uses Bun.sh for fast TypeScript execution
- **NodeExecutor**: Uses Node.js + pnpm with ts-node compilation
- Both create temporary workspaces with proper package.json and tsconfig.json setup

**Authentication** (`auth/`)
- Pluggable authentication system for private registries
- Currently supports AWS CodeArtifact
- Handles scoped package registry authentication

### Key Models (`models/`)
- `AppConfig`: Dependencies, dev dependencies, scoped packages, executor options
- `ScopedPackageOptions`: Registry configuration with authentication
- `CopyOptions`: File copying behavior with pattern matching

## Runtime Requirements

**For BunExecutor**: Bun binary on PATH (tested with v1.3.5)
**For NodeExecutor**: Node.js (>=18.0.0) and pnpm (v9.0.2) on PATH

Currently tested on Linux (WSL). Platform support for other OSes is planned.

## Usage Patterns

### BunExecutor (Recommended for Speed)
```go
app := synth.NewApp(executors.NewBunExecutor, logger)
app.Configure(ctx, models.AppConfig{
    Dependencies: map[string]string{"cdktf-pkg": "1.0.0"},
})
app.Eval(ctx, fs, mainTsScript, srcDir, outDir)
```

### NodeExecutor (More Configurable)
```go
app := synth.NewApp(executors.NewNodeExecutor, logger)
app.Configure(ctx, models.AppConfig{
    Dependencies: map[string]string{"cdktf-pkg": "1.0.0"},
    ExecutorOptions: map[string]string{
        "entrypoint": "pnpm",
        "synthScript": "ts-node --swc -P ./tsconfig.json main.ts",
    },
})
```

## File Structure Patterns

- `/cmd/` - CLI application examples
- `/executors/` - Runtime implementations with comprehensive tests
- `/models/` - Core interfaces and data structures
- `/auth/` - Authentication providers
- `/fixtures/` - Test fixtures and example projects
- `/executors/resources/` - Template files for generated package.json and tsconfig.json
- `/executors/__snapshots__/` - Snapshot test data for regression testing

## Testing Guidelines

- All executors have comprehensive test coverage in `*_test.go` files
- Uses snapshot testing with `go-snaps` for regression detection
- Test fixtures are in `/fixtures/` directory
- Update snapshots with `UPDATE_SNAPS=true go test ./...`
- Both unit tests and integration tests are included

## Key Dependencies

- `github.com/spf13/afero` - File system abstraction
- `go.uber.org/zap` - Structured logging
- AWS SDK v2 - CodeArtifact authentication
- `github.com/gkampitakis/go-snaps` - Snapshot testing
- `github.com/stretchr/testify` - Test assertions

## Current Development Status

The project is mature with core functionality complete. Recent work includes:
- PreSetup hook functionality
- File copying improvements (current branch: feat-copyFile)
- Comprehensive authentication support

Future planned features include CI/CD, additional executors (Go, go-typescript), and cross-platform support.