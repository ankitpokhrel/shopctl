package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

const (
	version  = 1 // Current config version.
	rootDir  = "shopctl"
	fileType = "yml"

	modeOwner = 0o700
)

type config struct {
	writer *viper.Viper
	dir    string
}

func newConfig(dir, name string) *config {
	return &config{
		writer: makeWriter(dir, name),
		dir:    dir,
	}
}

// home returns dir for the config.
func home() string {
	if home := os.Getenv("SHOPIFY_CONFIG_HOME"); home != "" {
		return filepath.Join(home, rootDir)
	}
	if home := os.Getenv("XDG_CONFIG_HOME"); home != "" {
		return filepath.Join(home, rootDir)
	}
	if home := os.Getenv("AppData"); runtime.GOOS == "windows" && home != "" {
		return filepath.Join(home, "ShopCTL")
	}

	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", rootDir)
}

func makeWriter(dir, name string) *viper.Viper {
	w := viper.New()
	w.SetConfigType(fileType)
	w.AddConfigPath(dir)
	w.SetConfigName(name)

	return w
}

func exists(dir string) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return false
	}
	return true
}

func ensureConfigFile(dir, file string) error {
	// Bail early if config already exists.
	if exists(dir) {
		return nil
	}

	// Create top-level dir.
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, modeOwner); err != nil {
			return err
		}
	}
	cfgFile := filepath.Join(dir, fmt.Sprintf("%s.%s", file, fileType))

	// Create config file.
	f, err := os.Create(cfgFile)
	if err != nil {
		return err
	}
	return f.Close()
}
