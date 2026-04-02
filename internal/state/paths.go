package state

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	stateDirName                   = ".alliecode"
	globalHistoryFileName          = "history.jsonl"
	projectsStateFileName          = "projects-state.json"
	settingsCacheFileName          = "settings-cache.json"
	authStateFileName              = "auth-state.json"
	sessionMetadataFileName        = "session-metadata.json"
	runtimeStateFileName           = "runtime-state.json"
	configDiagnosticsFileName      = "config-diagnostics.json"
	migrationsLedgerFileName       = "migrations-ledger.jsonl"
	sessionsDirName                = "sessions"
	projectOnboardingStateFileName = "project-onboarding-state.json"

	stateDirPerm  = 0o700
	stateFilePerm = 0o600
)

type Paths struct {
	HomeDir                    string
	ProjectDir                 string
	GlobalHistoryFile          string
	SessionsDir                string
	ProjectsStateFile          string
	SettingsCacheFile          string
	AuthStateFile              string
	SessionMetadataFile        string
	RuntimeStateFile           string
	ConfigDiagnosticsFile      string
	MigrationsLedgerFile       string
	ProjectOnboardingStateFile string
}

func ResolvePaths(homeBaseDir, projectBaseDir string) (Paths, error) {
	homeRoot, err := resolveRootDir(homeBaseDir, true)
	if err != nil {
		return Paths{}, fmt.Errorf("resolve home root: %w", err)
	}
	projectRoot, err := resolveRootDir(projectBaseDir, false)
	if err != nil {
		return Paths{}, fmt.Errorf("resolve project root: %w", err)
	}

	homeDir := filepath.Join(homeRoot, stateDirName)
	projectDir := filepath.Join(projectRoot, stateDirName)

	return Paths{
		HomeDir:                    homeDir,
		ProjectDir:                 projectDir,
		GlobalHistoryFile:          filepath.Join(homeDir, globalHistoryFileName),
		SessionsDir:                filepath.Join(homeDir, sessionsDirName),
		ProjectsStateFile:          filepath.Join(homeDir, projectsStateFileName),
		SettingsCacheFile:          filepath.Join(homeDir, settingsCacheFileName),
		AuthStateFile:              filepath.Join(homeDir, authStateFileName),
		SessionMetadataFile:        filepath.Join(homeDir, sessionMetadataFileName),
		RuntimeStateFile:           filepath.Join(homeDir, runtimeStateFileName),
		ConfigDiagnosticsFile:      filepath.Join(homeDir, configDiagnosticsFileName),
		MigrationsLedgerFile:       filepath.Join(homeDir, migrationsLedgerFileName),
		ProjectOnboardingStateFile: filepath.Join(projectDir, projectOnboardingStateFileName),
	}, nil
}

func EnsureHomeState(paths Paths) error {
	if err := os.MkdirAll(paths.HomeDir, stateDirPerm); err != nil {
		return fmt.Errorf("create state home directory: %w", err)
	}
	if err := os.MkdirAll(paths.SessionsDir, stateDirPerm); err != nil {
		return fmt.Errorf("create sessions directory: %w", err)
	}
	if err := ensureFile(paths.GlobalHistoryFile, nil); err != nil {
		return fmt.Errorf("ensure global history file: %w", err)
	}
	if err := ensureFile(paths.ProjectsStateFile, []byte("{}\n")); err != nil {
		return fmt.Errorf("ensure projects state file: %w", err)
	}
	if err := ensureFile(paths.SettingsCacheFile, []byte("{}\n")); err != nil {
		return fmt.Errorf("ensure settings cache file: %w", err)
	}
	if err := ensureFile(paths.AuthStateFile, []byte("{}\n")); err != nil {
		return fmt.Errorf("ensure auth state file: %w", err)
	}
	if err := ensureFile(paths.SessionMetadataFile, []byte("{}\n")); err != nil {
		return fmt.Errorf("ensure session metadata file: %w", err)
	}
	if err := ensureFile(paths.RuntimeStateFile, []byte("{}\n")); err != nil {
		return fmt.Errorf("ensure runtime state file: %w", err)
	}
	if err := ensureFile(paths.ConfigDiagnosticsFile, []byte("{}\n")); err != nil {
		return fmt.Errorf("ensure config diagnostics file: %w", err)
	}
	if err := ensureFile(paths.MigrationsLedgerFile, nil); err != nil {
		return fmt.Errorf("ensure migrations ledger file: %w", err)
	}
	return nil
}

func EnsureProjectState(paths Paths) error {
	if err := os.MkdirAll(paths.ProjectDir, stateDirPerm); err != nil {
		return fmt.Errorf("create project state directory: %w", err)
	}
	if err := ensureFile(paths.ProjectOnboardingStateFile, []byte("{}\n")); err != nil {
		return fmt.Errorf("ensure project onboarding state file: %w", err)
	}
	return nil
}

func EnsureState(paths Paths) error {
	if err := EnsureHomeState(paths); err != nil {
		return err
	}
	if err := EnsureProjectState(paths); err != nil {
		return err
	}
	return nil
}

func resolveRootDir(path string, home bool) (string, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		if home {
			h, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			trimmed = h
		} else {
			cwd, err := os.Getwd()
			if err != nil {
				return "", err
			}
			trimmed = cwd
		}
	}
	abs, err := filepath.Abs(trimmed)
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return filepath.Clean(resolved), nil
	}
	return filepath.Clean(abs), nil
}

func ensureFile(path string, initial []byte) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), stateDirPerm); err != nil {
		return err
	}
	content := initial
	if content == nil {
		content = []byte{}
	}
	return os.WriteFile(path, content, stateFilePerm)
}
