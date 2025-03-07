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
		"definition",
	}
	return sanitizeAndMarshal(patch(m), rmAlways, nil)
}

func (m MetafieldDefinition) MarshalJSON() ([]byte, error) {
	type patch MetafieldDefinition

	rmAlways := []string{
		"access",
		"capabilities",
		"metafields",
		"validations",
		"validationStatus",
		"standardTemplate",
	}
	return sanitizeAndMarshal(patch(m), rmAlways, nil)
}

func (c Customer) MarshalJSON() ([]byte, error) {
	type patch Customer

	rmAlways := []string{
		"addresses",
		"companyContactProfiles",
		"events",
		"orders",
		"mergeable",
		"metafields",
		"paymentMethods",
		"storeCreditAccounts",
		"subscriptionContracts",
	}
	rmIfNil := []string{
		"lastOrder",
		"metafield",
	}
	return sanitizeAndMarshal(patch(c), rmAlways, rmIfNil)
}

func (m MailingAddressConnection) MarshalJSON() ([]byte, error) {
	type patch MailingAddressConnection

	rmAlways := []string{
		"edges",
		"pageInfo",
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

func (i ProductVariantsBulkInput) MarshalJSON() ([]byte, error) {
	type patch ProductVariantsBulkInput

	rmAlways := []string{
		"mediaSrc",
	}
	rmIfNil := []string{
		"metafields",
		"inventoryQuantities",
	}
	return sanitizeAndMarshal(patch(i), rmAlways, rmIfNil)
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
