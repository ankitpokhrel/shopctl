package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ankitpokhrel/shopctl/schema"
)

var (
	// ErrTargetFound is returned when target is found.
	ErrTargetFound = fmt.Errorf("target found")
	// ErrNoTargetFound is returned if a target is not found.
	ErrNoTargetFound = fmt.Errorf("no target found")
)

// Registry is a backup registry.
type Registry struct {
	dir string
}

// NewRegistry constructs a new registry.
func NewRegistry(path string) (*Registry, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("error accessing path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("provided path is not a directory")
	}
	return &Registry{dir: path}, nil
}

// GetProductByID fetches a product by ID.
func (r *Registry) GetProductByID(id string) (*schema.Product, error) {
	loc, err := LookForDir(id, r.dir)
	if err != nil {
		return nil, err
	}

	if loc == "" {
		return nil, fmt.Errorf("product not found")
	}

	productRaw, err := ReadFileContents(filepath.Join(loc, "product.json"))
	if err != nil {
		return nil, err
	}

	var product schema.Product

	if err := json.Unmarshal(productRaw, &product); err != nil {
		return nil, fmt.Errorf("error unmarshalling product: %w", err)
	}
	return &product, nil
}
