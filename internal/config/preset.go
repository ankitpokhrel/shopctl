package config

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	keyAlias     = "alias"
	keyKind      = "type"
	keyBkpDir    = "backup_dir"
	keyResources = "resources"

	presetConfigDir = "preset"
)

// PresetItems defines item in a config file.
type PresetItems struct {
	Alias     string
	Kind      string
	BkpDir    string
	Resources []string
	Force     bool
}

// StoreConfig is a Shopify store config.
type PresetConfig struct {
	*config
	data PresetItems
}

// NewPresetConfig builds a new backup config for a given store.
func NewPresetConfig(store string, items PresetItems) *PresetConfig {
	dir := filepath.Join(home(), store, presetConfigDir)

	return &PresetConfig{
		config: newConfig(dir, items.Alias),
		data:   items,
	}
}

// Save saves the config of a store to the file.
func (c *PresetConfig) Save() error {
	if err := ensureConfigFile(c.dir, c.data.Alias, c.data.Force); err != nil {
		return err
	}
	return c.writeAll()
}

func (c *PresetConfig) writeAll() error {
	c.writer.Set(keyAlias, c.data.Alias)
	c.writer.Set(keyKind, c.data.Kind)
	c.writer.Set(keyBkpDir, c.data.BkpDir)
	c.writer.Set(keyResources, c.data.Resources)

	return c.writer.WriteConfig()
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
func ReadAllPreset(store string, file string) (*PresetItems, error) {
	root := filepath.Join(home(), store, presetConfigDir)

	w := makeWriter(root, file)
	if err := w.ReadInConfig(); err != nil {
		return nil, err
	}
	return &PresetItems{
		Alias:     w.GetString(keyAlias),
		Kind:      w.GetString(keyKind),
		BkpDir:    w.GetString(keyBkpDir),
		Resources: w.GetStringSlice(keyResources),
	}, nil
}

// DeletePreset deletes preset for a store if it exist.
func DeletePreset(store string, preset string, force bool) error {
	root := filepath.Join(home(), store, presetConfigDir)
	file := filepath.Join(root, fmt.Sprintf("%s.%s", preset, fileType))

	if !exists(file) {
		return ErrNoConfig
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
