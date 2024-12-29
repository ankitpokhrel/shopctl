package backup

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

// BackupFunc is defines how a resource is prepared.
type BackupFunc func() (any, error)

// Resource represents a backup resource.
type Resource struct {
	Type     ResourceType
	BackupFn BackupFunc
}

// NewResource constructs a new backup resource.
func NewResource(rt ResourceType, bkpFn BackupFunc) Resource {
	return Resource{
		Type:     rt,
		BackupFn: bkpFn,
	}
}

// ResourceCollection represents a collection of backup resources.
type ResourceCollection struct {
	RootType  ResourceType
	RootID    string
	Path      string
	Resources []Resource
}

// NewResourceCollection constructs a new collection of backup resources.
func NewResourceCollection(id, path string, rsc ...Resource) *ResourceCollection {
	rc := ResourceCollection{
		RootID: id,
		Path:   path,
	}
	rc.Resources = append(rc.Resources, rsc...)
	return &rc
}
