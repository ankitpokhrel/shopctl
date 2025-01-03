package engine

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBackup_Do(t *testing.T) {
	path := "./testdata/.tmp/"

	bkpEngine := New(NewBackup(WithBackupDir(path + "test")))
	bkpEngine.Register(Product)

	jobs := []ResourceCollection{
		{
			NewResource(
				Product,
				"2024/11/6d/8737843216608",
				func() (any, error) {
					content, err := os.ReadFile("./testdata/product.json")
					assert.NoError(t, err)

					var jsonContent map[string]any
					err = json.Unmarshal(content, &jsonContent)
					assert.NoError(t, err)

					return jsonContent, nil
				},
			),
			NewResource(
				ProductVariant,
				"2024/11/6d/8737843216608",
				func() (any, error) {
					return "variants", nil
				},
			),
			NewResource(
				ProductMedia,
				"2024/11/6d/8737843216608",
				func() (any, error) {
					return "media", nil
				},
			),
		},
		{
			NewResource(
				Product,
				"2024/11/6d/8737843347680",
				func() (any, error) {
					return "product", nil
				},
			),
			NewResource(
				ProductVariant,
				"2024/11/6d/8737843347680",
				func() (any, error) {
					return "variants", nil
				},
			),
			NewResource(
				ProductMedia,
				"2024/11/6d/8737843347680",
				func() (any, error) {
					return "media", nil
				},
			),
		},
		{
			NewResource(
				ProductMedia,
				"2024/12/ae/8773308023008",
				func() (any, error) {
					return "media", nil
				},
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
	assert.Equal(t, "\"variants\"", string(content))

	content, err = os.ReadFile(path + "test/2024/11/6d/8737843216608/media.json")
	assert.NoError(t, err)
	assert.Equal(t, "\"media\"", string(content))

	content, err = os.ReadFile(path + "test/2024/12/ae/8773308023008/media.json")
	assert.NoError(t, err)
	assert.Equal(t, "\"media\"", string(content))

	// Clean up.
	assert.NoError(t, os.RemoveAll(path))
}
