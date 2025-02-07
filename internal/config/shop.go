package config

import (
	"fmt"
	"path/filepath"

	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"

	"github.com/ankitpokhrel/shopctl"
)

const (
	shopConfig = ".shopconfig"
)

// StoreContext stores shopify store .
type StoreContext struct {
	Alias string  `koanf:"alias" yaml:"alias"`
	Store string  `koanf:"store" yaml:"store"`
	Token *string `koanf:"token" yaml:"token,omitempty"`
}

type shopItems struct {
	Version    string         `koanf:"ver" yaml:"ver"`
	Contexts   []StoreContext `koanf:"contexts" yaml:"contexts"`
	CurrentCtx string         `koanf:"currentContext" yaml:"currentContext"`
}

// ShopConfig is a Shopify store config.
type ShopConfig struct {
	*config
	data shopItems
}

// NewShopConfig constructs a new config for a given store.
func NewShopConfig() (*ShopConfig, error) {
	cfg, err := newConfig(home(), shopConfig, fileTypeYaml)
	if err != nil {
		return nil, err
	}

	// Load the existing config if it exists.
	var item shopItems
	if err := cfg.writer.Unmarshal("", &item); err != nil {
		return nil, err
	}

	ver := shopctl.AppConfigVersion
	if item.Version == "" {
		item.Version = ver
	}

	shopCfg := ShopConfig{
		config: cfg,
		data:   item,
	}
	return &shopCfg, nil
}

// SetStoreContext adds a store context to the shop config.
// It will update the context if it already exist.
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

// SetCurrentContext updates current active context.
func (c *ShopConfig) SetCurrentContext(ctx string) error {
	found := false
	for _, x := range c.data.Contexts {
		if x.Alias == ctx {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("no context exists with the name: %q", ctx)
	}

	c.data.CurrentCtx = ctx
	return nil
}

// Save saves the config of a store to the file.
func (c *ShopConfig) Save() error {
	k := koanf.New(".")

	if err := k.Load(structs.Provider(c.data, "yaml"), nil); err != nil {
		return err
	}
	if err := c.writer.Merge(k); err != nil {
		return err
	}
	return writeConfig(c.path, c.data)
}

// GetToken retrieves token of a store from the config.
func GetToken(alias string) string {
	w, err := loadConfig(filepath.Join(home(), fmt.Sprintf("%s.yml", alias)))
	if err != nil {
		return ""
	}

	var item shopItems
	if err := w.Unmarshal("", &item); err != nil {
		return ""
	}

	for _, c := range item.Contexts {
		if c.Alias == alias {
			return *c.Token
		}
	}
	return ""
}
