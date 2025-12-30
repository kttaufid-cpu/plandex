package fs

import (
	"os"
	"os/exec"
	"path/filepath"
	"plandex-cli/term"
)

var Cwd string
var PlandexDir string
var ProjectRoot string
var HomePlandexDir string
var CacheDir string

var HomeDir string
var HomeAuthPath string
var HomeAccountsPath string

// getXDGConfigHome returns XDG_CONFIG_HOME or ~/.config as default
func getXDGConfigHome(home string) string {
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return xdgConfig
	}
	return filepath.Join(home, ".config")
}

// getXDGCacheHome returns XDG_CACHE_HOME or ~/.cache as default
func getXDGCacheHome(home string) string {
	if xdgCache := os.Getenv("XDG_CACHE_HOME"); xdgCache != "" {
		return xdgCache
	}
	return filepath.Join(home, ".cache")
}

// getLegacyHomePlandexDir returns the old ~/.plandex-home-v2 path for migration
func getLegacyHomePlandexDir(home string) string {
	if os.Getenv("PLANDEX_ENV") == "development" {
		return filepath.Join(home, ".plandex-home-dev-v2")
	}
	return filepath.Join(home, ".plandex-home-v2")
}

// migrateLegacyConfig migrates config files from legacy path to XDG path
func migrateLegacyConfig(legacyDir, newConfigDir, newCacheDir string) {
	// Check if legacy directory exists
	if _, err := os.Stat(legacyDir); os.IsNotExist(err) {
		return
	}

	// Check if new config directory already has files (don't overwrite)
	if _, err := os.Stat(filepath.Join(newConfigDir, "auth.json")); err == nil {
		return
	}

	// Migrate auth.json
	legacyAuth := filepath.Join(legacyDir, "auth.json")
	if _, err := os.Stat(legacyAuth); err == nil {
		if data, err := os.ReadFile(legacyAuth); err == nil {
			os.WriteFile(filepath.Join(newConfigDir, "auth.json"), data, 0600)
		}
	}

	// Migrate accounts.json
	legacyAccounts := filepath.Join(legacyDir, "accounts.json")
	if _, err := os.Stat(legacyAccounts); err == nil {
		if data, err := os.ReadFile(legacyAccounts); err == nil {
			os.WriteFile(filepath.Join(newConfigDir, "accounts.json"), data, 0600)
		}
	}

	// Migrate cache directory contents
	legacyCacheDir := filepath.Join(legacyDir, "cache")
	if _, err := os.Stat(legacyCacheDir); err == nil {
		// Migrate tiktoken cache
		legacyTiktoken := filepath.Join(legacyCacheDir, "tiktoken")
		if entries, err := os.ReadDir(legacyTiktoken); err == nil {
			newTiktoken := filepath.Join(newCacheDir, "tiktoken")
			os.MkdirAll(newTiktoken, os.ModePerm)
			for _, entry := range entries {
				if !entry.IsDir() {
					if data, err := os.ReadFile(filepath.Join(legacyTiktoken, entry.Name())); err == nil {
						os.WriteFile(filepath.Join(newTiktoken, entry.Name()), data, 0644)
					}
				}
			}
		}
	}

	// Migrate project-specific files (current-plans-v2.json, settings-v2.json)
	entries, err := os.ReadDir(legacyDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "cache" {
			// This is likely a project directory
			projectDir := filepath.Join(legacyDir, entry.Name())
			newProjectDir := filepath.Join(newConfigDir, entry.Name())
			os.MkdirAll(newProjectDir, os.ModePerm)

			// Migrate current-plans-v2.json
			currentPlans := filepath.Join(projectDir, "current-plans-v2.json")
			if data, err := os.ReadFile(currentPlans); err == nil {
				os.WriteFile(filepath.Join(newProjectDir, "current-plans-v2.json"), data, 0644)
			}

			// Migrate plan subdirectories
			projectEntries, err := os.ReadDir(projectDir)
			if err != nil {
				continue
			}
			for _, pe := range projectEntries {
				if pe.IsDir() {
					// Plan directory
					planDir := filepath.Join(projectDir, pe.Name())
					newPlanDir := filepath.Join(newProjectDir, pe.Name())
					os.MkdirAll(newPlanDir, os.ModePerm)

					// Migrate settings-v2.json
					settings := filepath.Join(planDir, "settings-v2.json")
					if data, err := os.ReadFile(settings); err == nil {
						os.WriteFile(filepath.Join(newPlanDir, "settings-v2.json"), data, 0644)
					}
				}
			}
		}
	}
}

func init() {
	var err error
	Cwd, err = os.Getwd()
	if err != nil {
		term.OutputErrorAndExit("Error getting current working directory: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		term.OutputErrorAndExit("Couldn't find home dir: %v", err.Error())
	}
	HomeDir = home

	// Get XDG base directories
	xdgConfigHome := getXDGConfigHome(home)
	xdgCacheHome := getXDGCacheHome(home)

	// Set up plandex directories following XDG spec
	if os.Getenv("PLANDEX_ENV") == "development" {
		HomePlandexDir = filepath.Join(xdgConfigHome, "plandex", "dev-v2")
		CacheDir = filepath.Join(xdgCacheHome, "plandex", "dev-v2")
	} else {
		HomePlandexDir = filepath.Join(xdgConfigHome, "plandex", "v2")
		CacheDir = filepath.Join(xdgCacheHome, "plandex", "v2")
	}

	// Create the directories if they don't exist
	err = os.MkdirAll(HomePlandexDir, os.ModePerm)
	if err != nil {
		term.OutputErrorAndExit(err.Error())
	}

	err = os.MkdirAll(CacheDir, os.ModePerm)
	if err != nil {
		term.OutputErrorAndExit(err.Error())
	}

	// Migrate from legacy location if needed
	legacyDir := getLegacyHomePlandexDir(home)
	migrateLegacyConfig(legacyDir, HomePlandexDir, CacheDir)

	HomeAuthPath = filepath.Join(HomePlandexDir, "auth.json")
	HomeAccountsPath = filepath.Join(HomePlandexDir, "accounts.json")

	err = os.MkdirAll(filepath.Join(CacheDir, "tiktoken"), os.ModePerm)
	if err != nil {
		term.OutputErrorAndExit(err.Error())
	}
	err = os.Setenv("TIKTOKEN_CACHE_DIR", CacheDir)
	if err != nil {
		term.OutputErrorAndExit(err.Error())
	}

	FindPlandexDir()
	if PlandexDir != "" {
		ProjectRoot = Cwd
	}
}

func FindOrCreatePlandex() (string, bool, error) {
	FindPlandexDir()
	if PlandexDir != "" {
		ProjectRoot = Cwd
		return PlandexDir, false, nil
	}

	// Determine the directory path
	var dir string
	if os.Getenv("PLANDEX_ENV") == "development" {
		dir = filepath.Join(Cwd, ".plandex-dev-v2")
	} else {
		dir = filepath.Join(Cwd, ".plandex-v2")
	}

	err := os.Mkdir(dir, os.ModePerm)
	if err != nil {
		return "", false, err
	}
	PlandexDir = dir
	ProjectRoot = Cwd

	return dir, true, nil
}

func ProjectRootIsGitRepo() bool {
	if ProjectRoot == "" {
		return false
	}

	return IsGitRepo(ProjectRoot)
}

func IsGitRepo(dir string) bool {
	isGitRepo := false

	if isCommandAvailable("git") {
		// check whether we're in a git repo
		cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")

		cmd.Dir = dir

		err := cmd.Run()

		if err == nil {
			isGitRepo = true
		}
	}

	return isGitRepo
}

func FindPlandexDir() {
	PlandexDir = findPlandex(Cwd)
}

func findPlandex(baseDir string) string {
	var dir string
	if os.Getenv("PLANDEX_ENV") == "development" {
		dir = filepath.Join(baseDir, ".plandex-dev-v2")
	} else {
		dir = filepath.Join(baseDir, ".plandex-v2")
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		return dir
	}

	return ""
}

func isCommandAvailable(name string) bool {
	cmd := exec.Command(name, "--version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
