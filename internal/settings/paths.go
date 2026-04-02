package settings

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alliecatowo/alliecode/internal/config"
	"gopkg.in/yaml.v3"
)

type PathOptions struct {
	ProjectDir string
	Scope      Scope
}

func ResolvePath(opts PathOptions) (string, error) {
	projectDir := strings.TrimSpace(opts.ProjectDir)
	if projectDir == "" {
		var err error
		projectDir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("resolve working directory: %w", err)
		}
	}

	globalPath, projectPath, err := config.ConventionalPaths(projectDir)
	if err != nil {
		return "", err
	}
	globalPath = filepath.Clean(globalPath)
	projectPath = filepath.Clean(projectPath)

	switch opts.Scope {
	case ScopeGlobal:
		return globalPath, nil
	case ScopeProject:
		return projectPath, nil
	default:
		return "", fmt.Errorf("unsupported scope %q", opts.Scope)
	}
}

type ScopePaths struct {
	GlobalPath  string
	ProjectPath string
	Effective   string
	Layered     bool
}

func ResolveScopePaths(opts LoadOptions) (ScopePaths, error) {
	projectDir := strings.TrimSpace(opts.ProjectDir)
	if projectDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return ScopePaths{}, fmt.Errorf("resolve working directory: %w", err)
		}
		projectDir = cwd
	}
	globalPath, projectPath, err := config.ConventionalPaths(projectDir)
	if err != nil {
		return ScopePaths{}, err
	}
	out := ScopePaths{GlobalPath: filepath.Clean(globalPath), ProjectPath: filepath.Clean(projectPath)}
	if opts.Scope == nil {
		out.Layered = true
		out.Effective = out.ProjectPath
		return out, nil
	}
	resolved, err := ResolvePath(PathOptions{ProjectDir: projectDir, Scope: *opts.Scope})
	if err != nil {
		return ScopePaths{}, err
	}
	out.Effective = filepath.Clean(resolved)
	return out, nil
}

func LoadScoped(opts LoadOptions) (*config.Config, error) {
	if opts.Scope == nil {
		return config.LoadLayered(opts.ProjectDir)
	}
	path, err := ResolvePath(PathOptions{ProjectDir: opts.ProjectDir, Scope: *opts.Scope})
	if err != nil {
		return nil, err
	}
	return config.Load(path)
}

func SaveScoped(cfg *config.Config, opts SaveOptions) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	path, err := ResolvePath(PathOptions{ProjectDir: opts.ProjectDir, Scope: opts.Scope})
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
