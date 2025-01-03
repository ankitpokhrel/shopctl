package engine

const (
	Product        ResourceType = "product"
	ProductVariant ResourceType = "product_variant"
	ProductMedia   ResourceType = "product_media"
)

// ResourceType represents a type of a resource to backup.
type ResourceType string

// File returns the file name for the backup based on resource type.
func (r ResourceType) File() string {
	switch r {
	case Product:
		return "product"
	case ProductVariant:
		return "variants"
	case ProductMedia:
		return "media"
	}
	panic("unknown backup data type")
}

// ResourceHandler defines how a resource is prepared.
type ResourceHandler func() (any, error)

// Resource represents a backup resource.
type Resource struct {
	Type    ResourceType
	Path    string
	Handler ResourceHandler
}

// NewResource constructs a new backup resource.
func NewResource(rt ResourceType, path string, rh ResourceHandler) Resource {
	return Resource{
		Type:    rt,
		Path:    path,
		Handler: rh,
	}
}

// ResourceCollection is a collection of resources.
type ResourceCollection []Resource
