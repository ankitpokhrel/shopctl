package schema

import "encoding/json"

func (p Product) MarshalJSON() ([]byte, error) {
	type patch Product

	rmAlways := []string{
		"media",
		"variants",
		"bundleComponents",
		"collections",
		"events",
		"inCollection",
		"defaultCursor",
		"metafield",
		"metafields",
		"publishedInContext",
		"publishedOnPublication",
		"publishedOnCurrentPublication",
		"resourcePublicationOnCurrentPublication",
		"resourcePublications",
		"resourcePublicationsV2",
		"sellingPlanGroups",
		"unpublishedPublications",
	}
	rmIfNil := []string{
		"combinedListing",
		"contextualPricing",
	}
	return sanitizeAndMarshal(patch(p), rmAlways, rmIfNil)
}

func (p ProductVariant) MarshalJSON() ([]byte, error) {
	type patch ProductVariant

	rmAlways := []string{
		"defaultCursor",
		"deliveryProfile",
		"events",
		"media",
		"metafield",
		"metafields",
		"productVariantComponents",
		"sellingPlanGroups",
		"translations",
	}
	rmIfNil := []string{
		"contextualPricing",
	}
	return sanitizeAndMarshal(patch(p), rmAlways, rmIfNil)
}

func (i InventoryItem) MarshalJSON() ([]byte, error) {
	type patch InventoryItem

	rmAlways := []string{
		"countryHarmonizedSystemCodes",
		"inventoryLevel",
		"inventoryLevels",
	}
	rmIfNil := []string{
		"variant",
	}
	return sanitizeAndMarshal(patch(i), rmAlways, rmIfNil)
}

func (i Image) MarshalJSON() ([]byte, error) {
	type patch Image

	rmAlways := []string{
		"metafields",
	}
	rmIfNil := []string{
		"metafield",
	}
	return sanitizeAndMarshal(patch(i), rmAlways, rmIfNil)
}

func (m Metafield) MarshalJSON() ([]byte, error) {
	type patch Metafield

	rmAlways := []string{
		"compareDigest",
		"owner",
		"reference",
		"references",
		"value",
	}
	rmIfNil := []string{
		"definition",
	}
	return sanitizeAndMarshal(patch(m), rmAlways, rmIfNil)
}

func (m MetafieldDefinition) MarshalJSON() ([]byte, error) {
	type patch MetafieldDefinition

	rmAlways := []string{
		"access",
		"capabilities",
		"metafields",
	}
	return sanitizeAndMarshal(patch(m), rmAlways, nil)
}

func (i ProductInput) MarshalJSON() ([]byte, error) {
	type patch ProductInput

	rmIfNil := []string{
		"metafields",
		"productOptions",
		"collectionsToJoin",
		"collectionsToLeave",
	}
	return sanitizeAndMarshal(patch(i), nil, rmIfNil)
}

func (i CustomerInput) MarshalJSON() ([]byte, error) {
	type patch CustomerInput

	rmIfNil := []string{
		"addresses",
		"metafields",
		"taxExemptions",
	}
	return sanitizeAndMarshal(patch(i), nil, rmIfNil)
}

func sanitizeAndMarshal(input any, rmAlways, rmIfNil []string) ([]byte, error) {
	b, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	for _, f := range rmAlways {
		delete(m, f)
	}
	for _, f := range rmIfNil {
		if val, ok := m[f]; ok && val == nil {
			delete(m, f)
		}
	}

	return json.Marshal(m)
}
