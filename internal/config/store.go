package config

import (
	"errors"

	"github.com/ankitpokhrel/shopctl"
)

// Config keys.
const (
	keyApiVer = "_apiVer"
	keyStore  = "store"
)

type storeItems struct {
	apiVer string
	store  string
}

// StoreConfig is a Shopify store config.
type StoreConfig struct {
	*config
	data storeItems
}

// NewStoreConfig constructs a new config for a given store.
func NewStoreConfig(store string, alias string) *StoreConfig {
	return &StoreConfig{
		config: newConfig(home(), alias, fileTypeYaml),
		data: storeItems{
			apiVer: shopctl.ShopifyApiVersion,
			store:  store,
		},
	}
}

// Save saves the config of a store to the file.
func (c *StoreConfig) Save() error {
	if err := ensureConfigFile(c.dir, c.name, c.kind, false); err != nil && !errors.Is(err, ErrConfigExist) {
		return err
	}
	return c.writeAll()
}

func (c *StoreConfig) writeAll() error {
	c.writer.Set(keyApiVer, c.data.apiVer)
	c.writer.Set(keyStore, c.data.store)

	return c.writer.WriteConfig()
}
