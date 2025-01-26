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
	store := "store.myshopify.com"
	root := "./testdata/.tmp/shopctl"

	t.Setenv("SHOPIFY_CONFIG_HOME", "./testdata/.tmp/")

	c := NewStoreConfig(store)
	assert.NotNil(t, c)
	assert.NoError(t, c.Save())

	assert.DirExists(t, root)
	assert.FileExists(t, fmt.Sprintf("%s/%s/store.yml", root, store))
	assert.Equal(t, 1, c.data.version)
	assert.Equal(t, store, c.data.store)
	assert.Empty(t, c.data.token)
	assert.Equal(t, 1, c.writer.GetInt("_ver"))
	assert.Equal(t, store, c.writer.GetString("store"))
	assert.Empty(t, GetToken(store))

	c.SetToken("abc123")
	assert.NoError(t, c.Save())
	assert.Equal(t, "abc123", c.data.token)
	assert.Equal(t, "abc123", GetToken(store))

	p := NewPresetConfig(store, PresetItems{
		Alias:     "daily",
		Kind:      "full",
		BkpDir:    "./testdata/",
		BkpPrefix: "test",
		Resources: []string{"product"},
		Force:     true,
	})
	assert.NotNil(t, p)
	assert.NoError(t, p.Save())

	assert.DirExists(t, fmt.Sprintf("%s/%s/preset", root, store))
	assert.FileExists(t, fmt.Sprintf("%s/%s/preset/daily.yml", root, store))
	assert.Equal(t, "daily", p.writer.GetString("alias"))
	assert.Equal(t, "full", p.writer.GetString("type"))
	assert.Equal(t, "./testdata/", p.writer.GetString("backup.dir"))
	assert.Equal(t, "test", p.writer.GetString("backup.prefix"))
	assert.Equal(t, []string{"product"}, p.writer.GetStringSlice("resources"))

	assert.NoError(t, os.RemoveAll("./testdata/.tmp/"))
}
