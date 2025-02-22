package registry

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ankitpokhrel/shopctl/schema"
	"github.com/mholt/archives"
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
	if !strings.HasSuffix(info.Name(), ".tar.gz") && !info.IsDir() {
		return nil, fmt.Errorf("provided path is not a directory or .tar.gz file")
	}
	return &Registry{dir: path}, nil
}

// GetProductByID fetches a product by ID.
func (r *Registry) GetProductByID(id string) (*schema.Product, error) {
	if strings.HasSuffix(r.dir, ".tar.gz") {
		return r.getProductByIDFromZip(id)
	}

	loc, err := LookForDir(id, r.dir)
	if err != nil {
		if errors.Is(err, ErrNoTargetFound) {
			return nil, fmt.Errorf("product not found")
		}
		return nil, err
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

func (r *Registry) getProductByIDFromZip(id string) (*schema.Product, error) {
	var (
		product *schema.Product
		format  archives.Tar

		pattern = fmt.Sprintf(`products/.*/%s/product\.json`, regexp.QuoteMeta(id))
	)

	matcher, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	zipFile, err := os.Open(r.dir)
	if err != nil {
		return nil, err
	}
	defer func() { _ = zipFile.Close() }()

	gz, err := gzip.NewReader(zipFile)
	if err != nil {
		return nil, err
	}
	defer func() { _ = gz.Close() }()

	err = format.Extract(context.Background(), gz, func(ctx context.Context, f archives.FileInfo) error {
		if !matcher.MatchString(f.NameInArchive) {
			return nil
		}

		file, err := f.Open()
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()

		productRaw, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(productRaw, &product); err != nil {
			return fmt.Errorf("error unmarshalling product: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, fmt.Errorf("product not found")
	}
	return product, nil
}
