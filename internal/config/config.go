package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

// Config keys.
const (
	KeyVersion = "_ver"
	KeyStore   = "store"
	KeyToken   = "token"
)

const (
	version   = 1 // Current config version.
	configDir = "shopctl"
	fileType  = "yml"
	modeOwner = 0o700
)

type items struct {
	version int
	store   string
	token   string
}

// StoreConfig is a Shopify store config.
type StoreConfig struct {
	writer *viper.Viper
	file   string
	data   items
}

// NewStoreConfig constructs a new config for a given store.
func NewStoreConfig(store string) *StoreConfig {
	path := home()
	file := filepath.Join(path, fmt.Sprintf("%s.%s", store, fileType))

	w := viper.New()
	w.SetConfigType(fileType)
	w.AddConfigPath(path)
	w.SetConfigName(store)

	return &StoreConfig{
		writer: w,
		file:   file,
		data: items{
			version: version,
			store:   store,
		},
	}
}

// SetToken sets the token.
func (c *StoreConfig) SetToken(token string) {
	c.data.token = token
}

// Write updates value for the key.
func (c *StoreConfig) Write(key, val string) error {
	if err := c.ensureConfigFile(); err != nil {
		return err
	}
	c.writer.Set(key, val)
	return c.writer.WriteConfig()
}

// Save saves the config of a store to the file.
func (c *StoreConfig) Save() error {
	if err := c.ensureConfigFile(); err != nil {
		return err
	}
	return c.writeAll()
}

func (c *StoreConfig) exists() bool {
	if _, err := os.Stat(c.file); os.IsNotExist(err) {
		return false
	}
	return true
}

func (c *StoreConfig) ensureConfigFile() error {
	// Bail early if config already exists.
	if c.exists() {
		return nil
	}
	root := filepath.Dir(c.file)

	// Create top-level dir.
	if _, err := os.Stat(root); os.IsNotExist(err) {
		if err := os.MkdirAll(root, modeOwner); err != nil {
			return err
		}
	}

	// Create config file.
	f, err := os.Create(c.file)
	if err != nil {
		return err
	}
	return f.Close()
}

func (c *StoreConfig) writeAll() error {
	c.writer.Set(KeyVersion, c.data.version)
	c.writer.Set(KeyStore, c.data.store)

	if c.data.token != "" {
		c.writer.Set(KeyToken, c.data.token)
	}

	return c.writer.WriteConfig()
}

// home returns dir for the config.
func home() string {
	if home := os.Getenv("SHOPIFY_CONFIG_HOME"); home != "" {
		return filepath.Join(home, configDir)
	}
	if home := os.Getenv("XDG_CONFIG_HOME"); home != "" {
		return filepath.Join(home, configDir)
	}
	if home := os.Getenv("AppData"); runtime.GOOS == "windows" && home != "" {
		return filepath.Join(home, "ShopCTL")
	}

	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", configDir)
}

// GetToken retrieves token of a store from the config.
func GetToken(store string) string {
	path := home()

	w := viper.New()
	w.SetConfigType(fileType)
	w.AddConfigPath(path)
	w.SetConfigName(store)

	if err := w.ReadInConfig(); err != nil {
		return ""
	}
	return w.GetString(KeyToken)
}
