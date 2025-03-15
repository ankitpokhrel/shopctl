package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	yamlv3 "gopkg.in/yaml.v3"
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
	writer *koanf.Koanf
	kind   string
	dir    string
	name   string
	path   string
}

//nolint:unparam
func newConfig(dir, name, kind string) (*config, error) {
	cfgFile := filepath.Join(dir, fmt.Sprintf("%s.%s", name, kind))

	if err := ensureConfigFile(dir, cfgFile, false); err != nil && !errors.Is(err, ErrConfigExist) {
		return nil, err
	}

	cfg := config{
		kind: kind,
		dir:  dir,
		name: name,
		path: cfgFile,
	}

	w, err := loadConfig(cfgFile)
	if err != nil {
		return nil, err
	}
	cfg.writer = w

	return &cfg, nil
}

// Path is a config file path.
func (c config) Path() string {
	return c.path
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

func loadConfig(cfgFile string) (*koanf.Koanf, error) {
	k := koanf.New(".")
	f := file.Provider(cfgFile)

	if err := k.Load(f, yaml.Parser()); err != nil {
		return nil, err
	}
	return k, nil
}

func writeYAML(cfgFile string, data any) error {
	bytes, err := yamlv3.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(cfgFile, bytes, modeFile)
}

func writeJSON(cfgFile string, data any) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(cfgFile, bytes, modeFile)
}

func exists(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

func ensureConfigFile(dir string, cfgFile string, force bool) error {
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
