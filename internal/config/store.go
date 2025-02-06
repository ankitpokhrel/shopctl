package config

import (
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"

	"github.com/ankitpokhrel/shopctl"
)

type storeItems struct {
	ApiVer string `koanf:"apiVer" yaml:"apiVer"`
	Store  string `koanf:"store" yaml:"store"`
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

	return &StoreConfig{
		config: cfg,
		data: storeItems{
			ApiVer: shopctl.ShopifyApiVersion,
			Store:  store,
		},
	}, nil
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
