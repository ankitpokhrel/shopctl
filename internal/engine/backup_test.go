package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

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
	path := "./testdata/.tmp"
	now := time.Now().Format("2006_01_02_15_04_05")

	bkpEng := NewBackup("teststore.example.com", WithBackupRoot(path+"/test"), WithBackupPrefix("bkp"))
	root := fmt.Sprintf("%s/test/bkp_%s_%s", path, now, bkpEng.id)

	eng := New(bkpEng)
	eng.Register(Product)

	jobs := []ResourceCollection{
		{
			Parent: func() *Resource {
				r := NewResource(
					Product,
					"2024/11/6d/8737843216608",
					&mockHandler{dataFile: "./testdata/product.json"},
				)
				return &r
			}(),
			Children: []Resource{
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
		},
		{
			Parent: func() *Resource {
				r := NewResource(
					Product,
					"2024/11/6d/8737843347680",
					&mockHandler{dataFile: "./testdata/empty.json"},
				)
				return &r
			}(),
			Children: []Resource{
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
		},
		{
			Parent: func() *Resource {
				r := NewResource(
					ProductMedia,
					"2024/12/ae/8773308023008",
					&mockHandler{dataFile: "./testdata/empty.json"},
				)
				return &r
			}(),
		},
	}

	go func() {
		defer eng.Done(Product)

		for _, j := range jobs {
			eng.Add(Product, j)
		}
	}()

	for res := range eng.Run(Product) {
		assert.NoError(t, res.Err)
	}

	// Assert that folder and files were created.
	assert.DirExists(t, path+"/test")
	assert.DirExists(t, root+"/2024/11")
	assert.DirExists(t, root+"/2024/11/6d")
	assert.DirExists(t, root+"/2024/11/6d/8737843216608")
	assert.DirExists(t, root+"/2024/11/6d/8737843347680")
	assert.DirExists(t, root+"/2024/12")
	assert.DirExists(t, root+"/2024/12/ae")
	assert.DirExists(t, root+"/2024/12/ae/8773308023008")

	assert.FileExists(t, root+"/2024/11/6d/8737843216608/product.json")
	assert.FileExists(t, root+"/2024/11/6d/8737843216608/variants.json")
	assert.FileExists(t, root+"/2024/11/6d/8737843216608/media.json")

	assert.FileExists(t, root+"/2024/11/6d/8737843347680/product.json")
	assert.FileExists(t, root+"/2024/11/6d/8737843347680/variants.json")
	assert.FileExists(t, root+"/2024/11/6d/8737843347680/media.json")

	assert.FileExists(t, root+"/2024/12/ae/8773308023008/media.json")

	// Assert file contents.
	content, err := os.ReadFile(root + "/2024/11/6d/8737843216608/product.json")
	assert.NoError(t, err)
	assert.Equal(
		t,
		`{"createdAt":"2024-11-03T16:36:15Z","id":"gid://shopify/Product/8737843216608","title":"Test Product","totalInventory":50}`,
		string(content),
	)

	content, err = os.ReadFile(root + "/2024/11/6d/8737843216608/variants.json")
	assert.NoError(t, err)
	assert.Equal(
		t,
		`{"id":"gid://shopify/Product/8737843216608","variants":{"edges":[{"node":{"availableForSale":true,"createdAt":"2024-04-12T18:27:08Z","displayName":"Test Product"}}]}}`,
		string(content))

	content, err = os.ReadFile(root + "/2024/11/6d/8737843216608/media.json")
	assert.NoError(t, err)
	assert.Equal(
		t,
		`{"id":"gid://shopify/Product/8737843216608","media":{"edges":[{"node":{"id":"gid://shopify/MediaImage/33201292214520","mediaContentType":"IMAGE","mediaErrors":[],"mediaWarnings":[],"preview":{"image":{"altText":"Test Media","height":1600,"metafield":null,"metafields":{"edges":null,"nodes":null,"pageInfo":{"hasNextPage":false,"hasPreviousPage":false}},"url":"https://cdn.shopify.com/s/files/1/0695/7373/8744/files/Main_b13ad453-477c-4ed1-9b43-81f3345adfd6.jpg?v=1712946428","width":1600},"status":"READY"},"status":"READY"}}]}}`,
		string(content),
	)

	content, err = os.ReadFile(root + "/2024/12/ae/8773308023008/media.json")
	assert.NoError(t, err)
	assert.Equal(t, "{}", string(content))

	// Clean up.
	assert.NoError(t, os.RemoveAll(path))
}
