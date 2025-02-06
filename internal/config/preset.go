package config

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
)

const (
	presetConfigDir = "preset"
)

// PresetItems defines item in a config file.
type PresetItems struct {
	Alias     string   `koanf:"alias"`
	Kind      string   `koanf:"type"`
	BkpDir    string   `koanf:"backup.dir"`
	BkpPrefix string   `koanf:"backup.prefix"`
	Resources []string `koanf:"resources"`

	Force bool
}

// StoreConfig is a Shopify store config.
type PresetConfig struct {
	*config
	data PresetItems
}

// NewPresetConfig builds a new backup config for a given store.
func NewPresetConfig(store string, items PresetItems) (*PresetConfig, error) {
	dir := filepath.Join(home(), store, presetConfigDir)

	cfg, err := newConfig(dir, items.Alias, fileTypeYaml)
	if err != nil {
		return nil, err
	}

	return &PresetConfig{
		config: cfg,
		data:   items,
	}, nil
}

// Save saves the config of a store to the file.
func (c *PresetConfig) Save() error {
	return c.writeAll()
}

func (c *PresetConfig) writeAll() error {
	data, err := yaml.Parser().Marshal(c.config.writer.All())
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, modeFile)
}

// GetPresetLoc returns the location of the preset if it exist.
func GetPresetLoc(store string, preset string) (string, error) {
	root := filepath.Join(home(), store, presetConfigDir)
	file := filepath.Join(root, fmt.Sprintf("%s.%s", preset, fileTypeYaml))

	if !exists(file) {
		return "", ErrNoConfig
	}
	return file, nil
}

// ListPresets returns available presets for the store.
func ListPresets(store string) ([]string, error) {
	var out []string

	root := filepath.Join(home(), store, presetConfigDir)
	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		out = append(out, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, err
}

// ReadAllPreset reads preset config for a store.
func ReadAllPreset(store string, preset string) (*PresetItems, error) {
	root := filepath.Join(home(), store, presetConfigDir)

	w, err := loadConfig(filepath.Join(root, fmt.Sprintf("%s.%s", preset, fileTypeYaml)))
	if err != nil {
		return nil, err
	}

	var item PresetItems
	if err := w.Unmarshal("", &item); err != nil {
		return nil, err
	}
	return &item, nil
}

// DeletePreset deletes preset for a store if it exist.
func DeletePreset(store string, preset string, force bool) error {
	file, err := GetPresetLoc(store, preset)
	if err != nil {
		return err
	}
	if force {
		return os.Remove(file)
	}

	fmt.Printf("You are about to delete preset '%s' for store '%s'. Are you sure? (y/N): ", preset, store)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "y" || input == "yes" {
		return os.Remove(file)
	}
	return ErrActionAborted
}
