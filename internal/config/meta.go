package config

import (
	"os"
	"sync"

	"github.com/knadh/koanf/parsers/yaml"
)

const (
	KeyStatus    = "status"
	KeyResources = "resources"
	KeyTimeStart = "timeStart"
	KeyTimeEnd   = "timeEnd"

	keyID    = "id"
	keyStore = "store"

	metaConfigFile = "metadata"
)

// MetaItems defines item in a metadata file.
type RootMetaItems struct {
	ID        string   `koanf:"id"`
	Store     string   `koanf:"store"`
	TimeInit  int64    `koanf:"timeInitiated"`
	TimeStart int64    `koanf:"timeStart"`
	TimeEnd   int64    `koanf:"timeEnd"`
	Resources []string `koanf:"resources"`
	Kind      string   `koanf:"type"`
	Status    string   `koanf:"status"`
	User      string   `koanf:"user"`
}

// RootMeta is a root metadata for the initiated backup.
type RootMeta struct {
	*config
	mux  *sync.Mutex
	data RootMetaItems
}

// NewRootMeta builds a RootMeta object for the initiated backup.
func NewRootMeta(loc string, items RootMetaItems) (*RootMeta, error) {
	cfg, err := newConfig(loc, metaConfigFile, fileTypeYaml)
	if err != nil {
		return nil, err
	}

	return &RootMeta{
		config: cfg,
		mux:    &sync.Mutex{},
		data:   items,
	}, nil
}

// Set saves allowed key val to the file.
func (r *RootMeta) Set(keyAndVal map[string]any) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	for k, v := range keyAndVal {
		// ID and store could only be set once during meta creation.
		if k != keyID && k != keyStore {
			r.setData(k, v)
			if err := r.writer.Set(k, v); err != nil {
				return err
			}
		}
	}

	data, err := yaml.Parser().Marshal(r.writer.All())
	if err != nil {
		return err
	}
	return os.WriteFile(r.path, data, modeFile)
}

// Save writes metadata to the file.
func (r *RootMeta) Save() error {
	return writeConfig(r.path, r.data)
}

func (r *RootMeta) setData(key string, val any) {
	switch key {
	case KeyStatus:
		r.data.Status = val.(string)
	case KeyResources:
		r.data.Resources = val.([]string)
	case KeyTimeStart:
		r.data.TimeStart = val.(int64)
	case KeyTimeEnd:
		r.data.TimeEnd = val.(int64)
	}
}
