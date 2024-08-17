package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/environment-toolkit/go-synth"
	"github.com/environment-toolkit/go-synth/executors"
	"github.com/spf13/afero"

	"github.com/environment-toolkit/go-synth/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Define flags
	mainTsPath := flag.String("file", "", "Path to the main.ts file")
	dependencies := flag.String("deps", "", "Comma-separated list of dependencies in format 'pkg@version'")
	devDependencies := flag.String("devdeps", "", "Comma-separated list of devDependencies in format 'pkg@version'")
	srcDir := flag.String("src", "cdktf.out", "Source directory for synthesized files")
	outDir := flag.String("out", ".", "Output directory for synthesized files")

	// Parse flags
	flag.Parse()

	logConfig := zap.NewDevelopmentConfig()
	logConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Optional: colorize the log level
	logConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder        // Optional: use ISO8601 time format
	logConfig.Encoding = "console"
	logger, _ := logConfig.Build()

	// Validate required flags
	if *mainTsPath == "" {
		logger.Fatal("The -file flag is required.")
	}

	depsMap := parseDependencies(*dependencies)
	devDepsMap := parseDependencies(*devDependencies)

	if _, err := os.Stat(*mainTsPath); os.IsNotExist(err) {
		logger.Fatal("The specified main.ts file does not exist", zap.String("main.ts", *mainTsPath))
	}

	mainTs, err := os.ReadFile(*mainTsPath)
	if err != nil {
		logger.Fatal("Failed to read main.ts file", zap.Error(err))
	}

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals to gracefully shut down
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		logger.Info("\nReceived interrupt signal, shutting down...")
		cancel()
	}()

	app := synth.NewApp(executors.NewBunExecutor, logger)
	app.Configure(ctx, config.App{
		Dependencies:    depsMap,
		DevDependencies: devDepsMap,
	})
	// prepare afero fs
	destFs := afero.NewOsFs()

	// Execute the main.ts script
	app.Eval(ctx, destFs, string(mainTs), *srcDir, *outDir)
}

// parseDependencies parses a comma-separated list of dependencies into a map.
func parseDependencies(deps string) map[string]string {
	depsMap := make(map[string]string)
	if deps != "" {
		depList := strings.Split(deps, ",")
		for _, dep := range depList {
			parts := strings.Split(dep, ":")
			if len(parts) == 2 {
				depsMap[parts[0]] = parts[1]
			} else {
				log.Printf("Skipping invalid dependency format: %s", dep)
			}
		}
	}
	return depsMap
}
