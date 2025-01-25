package config

import (
	"errors"
	"path/filepath"
)

// Config keys.
const (
	keyVersion = "_ver"
	keyStore   = "store"
	keyToken   = "token"

	storeConfigFile = "store"
)

type storeItems struct {
	version int
	store   string
	token   string
}

// StoreConfig is a Shopify store config.
type StoreConfig struct {
	*config
	data storeItems
}

// NewStoreConfig constructs a new config for a given store.
func NewStoreConfig(store string) *StoreConfig {
	dir := filepath.Join(home(), store)

	return &StoreConfig{
		config: newConfig(dir, storeConfigFile),
		data: storeItems{
			version: version,
			store:   store,
		},
	}
}

// SetToken sets the token.
func (c *StoreConfig) SetToken(token string) {
	c.data.token = token
}

// Save saves the config of a store to the file.
func (c *StoreConfig) Save() error {
	if err := ensureConfigFile(c.dir, storeConfigFile, false); err != nil && !errors.Is(err, ErrConfigExist) {
		return err
	}
	return c.writeAll()
}

func (c *StoreConfig) writeAll() error {
	c.writer.Set(keyVersion, c.data.version)
	c.writer.Set(keyStore, c.data.store)

	if c.data.token != "" {
		c.writer.Set(keyToken, c.data.token)
	}

	return c.writer.WriteConfig()
}

// GetToken retrieves token of a store from the config.
func GetToken(store string) string {
	root := filepath.Join(home(), store)

	w := makeWriter(root, storeConfigFile)
	if err := w.ReadInConfig(); err != nil {
		return ""
	}
	return w.GetString(keyToken)
}
