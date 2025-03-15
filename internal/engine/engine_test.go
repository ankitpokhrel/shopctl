package engine

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockDoer is a mock implementation of the Doer interface.
type MockDoer struct {
	doFunc func(Resource, any) (any, error)
}

// Do mocks the Do method.
func (m *MockDoer) Do(r Resource, d any) (any, error) {
	return m.doFunc(r, d)
}

func TestEngine_Run(t *testing.T) {
	doer := &MockDoer{
		doFunc: func(r Resource, d any) (any, error) {
			if r.Type == "fail" {
				return nil, errors.New("mock error")
			}
			return nil, nil
		},
	}

	engine := New(doer)
	engine.Register(Product)

	done := make(chan struct{})
	go func() {
		defer close(done)
		engine.Add(Product, ResourceCollection{
			Parent: func() *Resource {
				r := Resource{Type: Product}
				return &r
			}(),
			Children: []Resource{
				{Type: ProductOption},
				{Type: ProductVariant},
				{Type: "fail"},
			},
		})
		engine.Done(Product)
	}()

	// Run the engine.
	results := engine.Run(Product)

	// Collect results.
	collected := make([]Result, 0)
	for res := range results {
		collected = append(collected, res)
	}

	assert.Len(t, collected, 4)
	assert.Equal(t, "product", string(collected[0].ResourceType))
	assert.Nil(t, collected[0].Err)
	assert.Equal(t, "product_option", string(collected[1].ResourceType))
	assert.Nil(t, collected[1].Err)
	assert.Equal(t, "product_variant", string(collected[2].ResourceType))
	assert.Nil(t, collected[2].Err)
	assert.Equal(t, "fail", string(collected[3].ResourceType))
	assert.NotNil(t, collected[3].Err)

	<-done
}
