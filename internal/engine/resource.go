package engine

const (
	Product           ResourceType = "product"
	ProductOption     ResourceType = "product_option"
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
	case ProductOption:
		return "product"
	case ProductVariant:
		return "product_variants"
	case ProductMedia:
		return "product_media"
	case ProductMetaField:
		return "product_metafields"
	case Customer:
		return "customer"
	case CustomerMetaField:
		return "customer_metafields"
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

// IsPrimary checks if the resource type is primary.
func (r ResourceType) IsPrimary() bool {
	return r == Product || r == Customer
}

// ResourceHandler is a handler for a resource.
type ResourceHandler interface {
	Handle(data any) (any, error)
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
type ResourceCollection struct {
	Parent   *Resource
	Children []Resource
}

// GetPrimaryResourceTypes returns primary resource types.
func GetPrimaryResourceTypes() []ResourceType {
	return []ResourceType{
		Product,
		Customer,
	}
}

// GetProductResourceTypes returns product resource types in order.
func GetProductResourceTypes() []ResourceType {
	return []ResourceType{
		Product,
		ProductOption,
		ProductVariant,
		ProductMetaField,
		ProductMedia,
	}
}

// GetCustomerResourceTypes returns all resource types in order.
func GetCustomerResourceTypes() []ResourceType {
	return []ResourceType{
		Customer,
		CustomerMetaField,
	}
}

// GetAllResourceTypes returns all resource types in order.
func GetAllResourceTypes() []ResourceType {
	return []ResourceType{
		Product,
		ProductOption,
		ProductVariant,
		ProductMetaField,
		ProductMedia,
		Customer,
		CustomerMetaField,
	}
}
