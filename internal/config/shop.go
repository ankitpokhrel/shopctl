package config

import (
	"errors"

	"github.com/ankitpokhrel/shopctl"
)

const (
	shopConfig = ".shopconfig"

	keyVersion    = "_ver"
	keyToken      = "token"
	keyContexts   = "contexts"
	keyCurrentCtx = "currentContext"
)

// StoreContext stores shopify store .
type StoreContext struct {
	Alias string  `mapstructure:"alias"`
	Store string  `mapstructure:"store"`
	Token *string `mapstructure:"token"`
}

// MarshalYAML is a custom YAML Marshaler for StoreContext.
func (sc StoreContext) MarshalYAML() (interface{}, error) {
	m := map[string]string{
		"alias": sc.Alias,
		"store": sc.Store,
	}
	if sc.Token != nil && *sc.Token != "" {
		m["token"] = *sc.Token
	}
	return m, nil
}

type shopItems struct {
	Version    string         `mapstructure:"_ver"`
	Contexts   []StoreContext `mapstructure:"contexts"`
	CurrentCtx string         `mapstructure:"currentContext"`
}

// ShopConfig is a Shopify store config.
type ShopConfig struct {
	*config
	data shopItems
}

// NewShopConfig constructs a new config for a given store.
func NewShopConfig() (*ShopConfig, error) {
	config := newConfig(home(), shopConfig, fileTypeYaml)

	shopCfg := ShopConfig{
		config: config,
		data: shopItems{
			Version: shopctl.AppConfigVersion,
		},
	}

	// Load the existing config if it exists.
	if err := config.writer.ReadInConfig(); err == nil {
		var items shopItems
		if err := config.writer.Unmarshal(&items); err != nil {
			return nil, err
		}
		shopCfg.data = items
	}

	return &shopCfg, nil
}

// SetStoreContext adds a store context to the shop config
// It will update the config if the context already exist.
func (c *ShopConfig) SetStoreContext(ctx *StoreContext) {
	for i, x := range c.data.Contexts {
		if x.Store != ctx.Store {
			continue
		}
		c.data.Contexts[i].Alias = ctx.Alias
		if ctx.Token != nil {
			c.data.Contexts[i].Token = ctx.Token
		}
		return
	}
	c.data.Contexts = append(c.data.Contexts, *ctx)
}

// Save saves the config of a store to the file.
func (c *ShopConfig) Save() error {
	if err := ensureConfigFile(c.dir, c.name, c.kind, false); err != nil && !errors.Is(err, ErrConfigExist) {
		return err
	}
	return c.writeAll()
}

func (c *ShopConfig) writeAll() error {
	c.writer.Set(keyVersion, c.data.Version)
	c.writer.Set(keyContexts, c.data.Contexts)
	c.writer.Set(keyCurrentCtx, c.data.CurrentCtx)

	return c.writer.WriteConfig()
}

// GetToken retrieves token of a store from the config.
func GetToken(alias string) string {
	w := makeYamlWriter(home(), alias)
	if err := w.ReadInConfig(); err != nil {
		return ""
	}

	var item shopItems
	if err := w.Unmarshal(&item); err != nil {
		return ""
	}

	for _, c := range item.Contexts {
		if c.Alias == alias {
			return *c.Token
		}
	}
	return ""
}
