package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStoreConfig_Save(t *testing.T) {
	assert.NoError(t, os.Setenv("SHOPIFY_CONFIG_HOME", "./testdata/.tmp/"))
	defer func() {
		assert.NoError(t, os.Unsetenv("SHOPIFY_CONFIG_HOME"))
	}()

	c := NewStoreConfig("store.myshopify.com")
	assert.NotNil(t, c)
	assert.NoError(t, c.Save())

	assert.DirExists(t, "./testdata/.tmp/shopctl")
	assert.FileExists(t, "./testdata/.tmp/shopctl/store.myshopify.com.yml")
	assert.Equal(t, 1, c.data.version)
	assert.Equal(t, "store.myshopify.com", c.data.store)
	assert.Empty(t, c.data.token)
	assert.Equal(t, 1, c.writer.GetInt("_ver"))
	assert.Equal(t, "store.myshopify.com", c.writer.GetString("store"))
	assert.Empty(t, GetToken("store.myshopify.com"))

	c.SetToken("abc123")
	assert.NoError(t, c.Save())
	assert.Equal(t, "abc123", c.data.token)
	assert.Equal(t, "abc123", GetToken("store.myshopify.com"))

	assert.NoError(t, c.Write("store", "temp"))
	assert.Equal(t, "temp", c.writer.GetString("store"))

	assert.NoError(t, os.RemoveAll("./testdata/.tmp/"))
}
