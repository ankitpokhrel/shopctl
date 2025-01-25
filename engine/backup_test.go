package engine

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockHandler struct {
	dataFile string
}

func (m *mockHandler) Handle() (any, error) {
	content, err := os.ReadFile(m.dataFile)
	if err != nil {
		return nil, err
	}

	var jsonContent map[string]any
	err = json.Unmarshal(content, &jsonContent)
	if err != nil {
		return nil, err
	}

	return jsonContent, nil
}

func TestBackup_Do(t *testing.T) {
	path := "./testdata/.tmp/"

	bkpEngine := New(NewBackup(WithBackupDir(path + "test")))
	bkpEngine.Register(Product)

	jobs := []ResourceCollection{
		{
			NewResource(
				Product,
				"2024/11/6d/8737843216608",
				&mockHandler{dataFile: "./testdata/product.json"},
			),
			NewResource(
				ProductVariant,
				"2024/11/6d/8737843216608",
				&mockHandler{dataFile: "./testdata/variants.json"},
			),
			NewResource(
				ProductMedia,
				"2024/11/6d/8737843216608",
				&mockHandler{dataFile: "./testdata/media.json"},
			),
		},
		{
			NewResource(
				Product,
				"2024/11/6d/8737843347680",
				&mockHandler{dataFile: "./testdata/empty.json"},
			),
			NewResource(
				ProductVariant,
				"2024/11/6d/8737843347680",
				&mockHandler{dataFile: "./testdata/empty.json"},
			),
			NewResource(
				ProductMedia,
				"2024/11/6d/8737843347680",
				&mockHandler{dataFile: "./testdata/empty.json"},
			),
		},
		{
			NewResource(
				ProductMedia,
				"2024/12/ae/8773308023008",
				&mockHandler{dataFile: "./testdata/empty.json"},
			),
		},
	}

	go func() {
		defer bkpEngine.Done(Product)

		for _, j := range jobs {
			bkpEngine.Add(Product, j)
		}
	}()

	for res := range bkpEngine.Run(Product) {
		assert.NoError(t, res.Err)
	}

	// Assert that folder and files were created.
	assert.DirExists(t, path+"test")
	assert.DirExists(t, path+"test/2024/11")
	assert.DirExists(t, path+"test/2024/11/6d")
	assert.DirExists(t, path+"test/2024/11/6d/8737843216608")
	assert.DirExists(t, path+"test/2024/11/6d/8737843347680")
	assert.DirExists(t, path+"test/2024/12")
	assert.DirExists(t, path+"test/2024/12/ae")
	assert.DirExists(t, path+"test/2024/12/ae/8773308023008")

	assert.FileExists(t, path+"test/2024/11/6d/8737843216608/product.json")
	assert.FileExists(t, path+"test/2024/11/6d/8737843216608/variants.json")
	assert.FileExists(t, path+"test/2024/11/6d/8737843216608/media.json")

	assert.FileExists(t, path+"test/2024/11/6d/8737843347680/product.json")
	assert.FileExists(t, path+"test/2024/11/6d/8737843347680/variants.json")
	assert.FileExists(t, path+"test/2024/11/6d/8737843347680/media.json")

	assert.FileExists(t, path+"test/2024/12/ae/8773308023008/media.json")

	// Assert file contents.
	content, err := os.ReadFile(path + "test/2024/11/6d/8737843216608/product.json")
	assert.NoError(t, err)
	assert.Equal(
		t,
		`{"createdAt":"2024-11-03T16:36:15Z","id":"gid://shopify/Product/8737843216608","title":"Test Product","totalInventory":50}`,
		string(content),
	)

	content, err = os.ReadFile(path + "test/2024/11/6d/8737843216608/variants.json")
	assert.NoError(t, err)
	assert.Equal(
		t,
		`{"id":"gid://shopify/Product/8737843216608","variants":{"edges":[{"node":{"availableForSale":true,"createdAt":"2024-04-12T18:27:08Z","displayName":"Test Product"}}]}}`,
		string(content))

	content, err = os.ReadFile(path + "test/2024/11/6d/8737843216608/media.json")
	assert.NoError(t, err)
	assert.Equal(
		t,
		`{"id":"gid://shopify/Product/8737843216608","media":{"edges":[{"node":{"id":"gid://shopify/MediaImage/33201292214520","mediaContentType":"IMAGE","mediaErrors":[],"mediaWarnings":[],"preview":{"image":{"altText":"Test Media","height":1600,"metafield":null,"metafields":{"edges":null,"nodes":null,"pageInfo":{"hasNextPage":false,"hasPreviousPage":false}},"url":"https://cdn.shopify.com/s/files/1/0695/7373/8744/files/Main_b13ad453-477c-4ed1-9b43-81f3345adfd6.jpg?v=1712946428","width":1600},"status":"READY"},"status":"READY"}}]}}`,
		string(content),
	)

	content, err = os.ReadFile(path + "test/2024/12/ae/8773308023008/media.json")
	assert.NoError(t, err)
	assert.Equal(t, "{}", string(content))

	// Clean up.
	assert.NoError(t, os.RemoveAll(path))
}
