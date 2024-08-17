package executors

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"text/template"

	"github.com/environment-toolkit/go-synth/config"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

var (
	//go:embed all:resources/*
	embeddedFiles embed.FS
)

type templateStore struct {
	logger    *zap.Logger
	templates map[string]*template.Template
	basePath  string
}

func initializeTemplates(logger *zap.Logger, basePath string) *templateStore {
	tmpls := make(map[string]*template.Template)
	err := fs.WalkDir(embeddedFiles, basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("unable to walk embedded files: %w", err)
		}
		if d.IsDir() {
			return nil // skip dirs
		}

		extension := filepath.Ext(path)
		if extension == ".tmpl" {
			content, err := fs.ReadFile(embeddedFiles, path)
			if err != nil {
				return fmt.Errorf("unable to read template %s: %w", path, err)
			}
			tmpl, err := template.New(filepath.Base(path)).Funcs(funcMap).Parse(string(content))
			if err != nil {
				return fmt.Errorf("unable to parse template %s: %w", path, err)
			}
			tmpls[path] = tmpl
		}
		return nil
	})

	// we control the embedded templates, so we can panic here
	if err != nil {
		logger.Fatal("unable to walk embedded resources", zap.Error(err), zap.String("basePath", basePath))
	}
	return &templateStore{
		logger:    logger,
		templates: tmpls,
		basePath:  basePath,
	}
}

func (t *templateStore) setupFs(ctx context.Context, dest afero.Fs, config config.App) error {
	err := fs.WalkDir(embeddedFiles, t.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("unable to setup fs: %w", err)
		}

		// Check if the context is done
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if d.IsDir() {
			return nil // skip dirs
		}
		if err := ensurePath(dest, path); err != nil {
			return err
		}

		// strip basepath from path
		target := path[len(t.basePath):]
		extension := filepath.Ext(path)
		if extension == ".tmpl" {
			target = removeExtension(target)
			writer, err := dest.OpenFile(target, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return fmt.Errorf("unable to open file %s, %w", path, err)
			}
			tpl, ok := t.templates[path]
			if !ok {
				t.logger.Error("template not found", zap.String("path", path))
			}
			if err := tpl.Execute(writer, config); err != nil {
				return fmt.Errorf("unable to execute template %s, %w", path, err)
			}

			t.logger.Info("templated", zap.String("template", path), zap.String("target", target))
			return nil
		}
		return copyFile(afero.FromIOFS{FS: embeddedFiles}, dest, path, target)
	})

	return err
}
