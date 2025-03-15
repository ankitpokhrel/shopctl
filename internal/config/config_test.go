package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigHome(t *testing.T) {
	assert.NoError(t, os.Unsetenv("SHOPIFY_CONFIG_HOME"))
	assert.NoError(t, os.Unsetenv("XDG_CONFIG_HOME"))

	userHome, err := os.UserHomeDir()
	assert.NoError(t, err)
	assert.Equal(t, userHome+"/.config/shopctl", home())

	t.Setenv("XDG_CONFIG_HOME", "./testdata")
	assert.Equal(t, "testdata/shopctl", home())
}

func TestConfigSave(t *testing.T) {
	alias := "teststore"
	store := "store.myshopify.com"
	root := "./testdata/.tmp/shopctl"

	t.Setenv("SHOPIFY_CONFIG_HOME", "./testdata/.tmp/")

	c, _ := NewShopConfig()
	assert.NotNil(t, c)
	assert.NoError(t, c.Save())

	assert.DirExists(t, root)
	assert.FileExists(t, fmt.Sprintf("%s/.shopconfig.yml", root))
	assert.Equal(t, "v0", c.data.Version)
	assert.Equal(t, "", c.data.CurrentCtx)
	assert.Equal(t, "v0", c.writer.Get("ver"))
	assert.Equal(t, "", c.writer.Get("currentContext"))
	assert.Empty(t, GetToken(alias))

	sc, _ := NewStoreConfig(store, alias)
	assert.NoError(t, sc.Save())
	assert.FileExists(t, fmt.Sprintf("%s/%s.yml", root, alias))
	assert.Equal(t, "2025-01", sc.writer.Get("apiVer"))
	assert.Equal(t, store, sc.writer.Get("store"))

	assert.NoError(t, os.RemoveAll("./testdata/.tmp/"))
}
