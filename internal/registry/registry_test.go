package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProductByID(t *testing.T) {
	pid := "8737843216608"
	path := "./testdata/bkp/"

	reg, err := NewRegistry("invalid")
	assert.NotNil(t, err)
	assert.Nil(t, reg)

	reg, err = NewRegistry(path)
	assert.NoError(t, err)

	product, err := reg.GetProductByID(pid)
	assert.NoError(t, err)
	assert.Equal(t, "gid://shopify/Product/8737843216608", product.ID)

	product, err = reg.GetProductByID("invalid")
	assert.Error(t, err)
	assert.Nil(t, product)

	reg, err = NewRegistry("./testdata/")
	assert.NoError(t, err)

	// This should fail because the file is located above the max depth.
	product, err = reg.GetProductByID(pid)
	assert.Error(t, err)
	assert.Nil(t, product)
}
