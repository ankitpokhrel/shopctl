package engine

const (
	Product           ResourceType = "product"
	ProductVariant    ResourceType = "product_variant"
	ProductMedia      ResourceType = "product_media"
	ProductMetaField  ResourceType = "product_metafield"
	Customer          ResourceType = "customer"
	CustomerMetaField ResourceType = "customer_metafield"
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
	case ProductMetaField:
		return "metafields"
	case Customer:
		return "customer"
	case CustomerMetaField:
		return "metafields"
	}
	panic("unknown resource type")
}

// RootDir returns a root level dir for the resource type.
func (r ResourceType) RootDir() string {
	switch r {
	case Product:
		return "products"
	case Customer:
		return "customers"
	}
	panic("unknown root resource type")
}

// ResourceHandler is a handler for a resource.
type ResourceHandler interface {
	Handle() (any, error)
}

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
