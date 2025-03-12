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

	"github.com/mholt/archives"

	"github.com/ankitpokhrel/shopctl/internal/api"
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

	// Skip if we don't find variants file.
	variantsRaw, _ := ReadFileContents(filepath.Join(loc, "variants.json"))
	if len(variantsRaw) > 0 {
		var variants api.ProductVariantData
		if err := json.Unmarshal(variantsRaw, &variants); err != nil {
			return nil, fmt.Errorf("error unmarshalling product variants: %w", err)
		}
		product.Variants.Nodes = variants.Variants.Nodes
	}

	// Skip if we don't find media file.
	mediasRaw, _ := ReadFileContents(filepath.Join(loc, "media.json"))
	if len(mediasRaw) > 0 {
		var medias api.ProductMediaData
		if err := json.Unmarshal(mediasRaw, &medias); err != nil {
			return nil, fmt.Errorf("error unmarshalling product medias: %w", err)
		}
		var nodes []any
		for _, n := range medias.Media.Nodes {
			nodes = append(nodes, n)
		}
		product.Media.Nodes = nodes
	}

	return &product, nil
}

func (r *Registry) getProductByIDFromZip(id string) (*schema.Product, error) {
	var (
		product        *schema.Product
		productVariant []schema.ProductVariant
		productMedia   []any
		format         archives.Tar

		pattern = fmt.Sprintf(`products/.*/%s/.*\.json$`, regexp.QuoteMeta(id))
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

		switch f.Name() {
		case "product.json":
			productRaw, err := io.ReadAll(file)
			if err != nil {
				return err
			}
			if err := json.Unmarshal(productRaw, &product); err != nil {
				return fmt.Errorf("error unmarshalling product: %w", err)
			}
		case "variants.json":
			variantsRaw, _ := io.ReadAll(file)
			if len(variantsRaw) > 0 {
				var variants api.ProductVariantData
				if err := json.Unmarshal(variantsRaw, &variants); err != nil {
					return fmt.Errorf("error unmarshalling product variants: %w", err)
				}
				productVariant = variants.Variants.Nodes
			}
		case "media.json":
			mediasRaw, _ := io.ReadAll(file)
			if len(mediasRaw) > 0 {
				var medias api.ProductMediaData
				if err := json.Unmarshal(mediasRaw, &medias); err != nil {
					return fmt.Errorf("error unmarshalling product medias: %w", err)
				}
				var nodes []any
				for _, n := range medias.Media.Nodes {
					nodes = append(nodes, n)
				}
				productMedia = nodes
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if product == nil {
		return nil, fmt.Errorf("product not found")
	}
	if productVariant != nil {
		product.Variants.Nodes = productVariant
	}
	if productMedia != nil {
		product.Media.Nodes = productMedia
	}
	return product, nil
}
