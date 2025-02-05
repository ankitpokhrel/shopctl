package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

const (
	rootDir = "shopctl"

	fileTypeYaml = "yml"
	fileTypeJson = "json"

	modeDir   = 0o755
	modeFile  = 0o644
	modeOwner = 0o700
)

var (
	// ErrConfigExist is thrown if the config file already exist.
	ErrConfigExist = fmt.Errorf("config already exist")
	// ErrNoConfig is thrown if a config file couldn't be found.
	ErrNoConfig = fmt.Errorf("config doesn't exist")
	// ErrActionAborted is thrown if a user cancels an action.
	ErrActionAborted = fmt.Errorf("action aborted")
)

type config struct {
	writer *viper.Viper
	kind   string
	dir    string
	name   string
}

func newConfig(dir, name, kind string) *config {
	return &config{
		writer: makeWriter(dir, name, kind),
		kind:   kind,
		dir:    dir,
		name:   name,
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

func makeWriter(dir, name, kind string) *viper.Viper {
	w := viper.New()
	w.SetConfigType(kind)
	w.AddConfigPath(dir)
	w.SetConfigName(name)

	return w
}

func makeYamlWriter(dir, name string) *viper.Viper {
	return makeWriter(dir, name, fileTypeYaml)
}

func exists(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

func ensureConfigFile(dir, file, kind string, force bool) error {
	cfgFile := filepath.Join(dir, fmt.Sprintf("%s.%s", file, kind))

	// Bail early if config already exists.
	if !force && exists(cfgFile) {
		return ErrConfigExist
	}

	// Create top-level dir.
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, modeOwner); err != nil {
			return err
		}
	}

	// Create config file.
	f, err := os.Create(cfgFile)
	if err != nil {
		return err
	}
	return f.Close()
}
