package config

import (
	"os"

	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"

	"github.com/ankitpokhrel/shopctl"
)

type BackupStrategy struct {
	Name      string   `koanf:"name" yaml:"name"`
	Kind      string   `koanf:"type" yaml:"type"`
	BkpDir    string   `koanf:"dir" yaml:"dir"`
	BkpPrefix string   `koanf:"prefix" yaml:"prefix"`
	Resources []string `koanf:"resources" yaml:"resources"`
}

type storeItems struct {
	ApiVer     string           `koanf:"apiVer" yaml:"apiVer"`
	Store      string           `koanf:"store" yaml:"store"`
	Strategies []BackupStrategy `koanf:"strategies" yaml:"strategies"`
}

// StoreConfig is a Shopify store config.
type StoreConfig struct {
	*config
	data storeItems
}

// NewStoreConfig constructs a new config for a given store.
func NewStoreConfig(store string, alias string) (*StoreConfig, error) {
	cfg, err := newConfig(home(), alias, fileTypeYaml)
	if err != nil {
		return nil, err
	}

	// Load the existing config if it exists.
	var item storeItems
	if err := cfg.writer.Unmarshal("", &item); err != nil {
		return nil, err
	}

	ver := shopctl.ShopifyApiVersion
	if item.ApiVer == "" {
		item.ApiVer = ver
	}
	if item.Store == "" {
		item.Store = store
	}

	storeCfg := StoreConfig{
		config: cfg,
		data:   item,
	}
	return &storeCfg, nil
}

// SetStoreBackupStrategy adds a store context to the shop config.
// It will update the context if it already exist.
func (c *StoreConfig) SetStoreBackupStrategy(bst *BackupStrategy) {
	for i, s := range c.data.Strategies {
		if s.Name != bst.Name {
			continue
		}
		c.data.Strategies[i] = *bst
		return
	}
	c.data.Strategies = append(c.data.Strategies, *bst)
}

// Save saves the config of a store to the file.
func (c *StoreConfig) Save() error {
	k := koanf.New(".")

	if err := k.Load(structs.Provider(c.data, "yaml"), nil); err != nil {
		return err
	}
	if err := c.writer.Merge(k); err != nil {
		return err
	}
	return writeConfig(c.path, c.data)
}

// Delete removes the config file.
func (c *StoreConfig) Delete() error {
	return os.Remove(c.Path())
}
