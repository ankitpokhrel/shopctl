package schema

import "encoding/json"

// MarshalJSON patches and then marshals ProductInput struct.
func (c ProductInput) MarshalJSON() ([]byte, error) {
	type patch ProductInput

	b, err := json.Marshal(patch(c))
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	rmIfNil := []string{"metafields", "productOptions", "collectionsToJoin", "collectionsToLeave"}

	for _, f := range rmIfNil {
		if val, ok := m[f]; ok && val == nil {
			delete(m, f)
		}
	}
	return json.Marshal(m)
}

// MarshalJSON patches and then marshals CustomerInput struct.
func (c CustomerInput) MarshalJSON() ([]byte, error) {
	type patch CustomerInput

	b, err := json.Marshal(patch(c))
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	rmIfNil := []string{"addresses", "metafields", "taxExemptions"}

	for _, f := range rmIfNil {
		if val, ok := m[f]; ok && val == nil {
			delete(m, f)
		}
	}
	return json.Marshal(m)
}
