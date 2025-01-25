package config

import (
	"path/filepath"
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
	if err := ensureConfigFile(c.dir, c.data.Alias); err != nil {
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
