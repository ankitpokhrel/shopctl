// Code generated by introspect; DO NOT EDIT.

package schema

type ProductInput struct {
	DescriptionHtml        *string                     `json:"descriptionHtml,omitempty"`
	Handle                 *string                     `json:"handle,omitempty"`
	Seo                    *SEOInput                   `json:"seo,omitempty"`
	ProductType            *string                     `json:"productType,omitempty"`
	Category               *string                     `json:"category,omitempty"`
	Tags                   []any                       `json:"tags"`
	TemplateSuffix         *string                     `json:"templateSuffix,omitempty"`
	GiftCardTemplateSuffix *string                     `json:"giftCardTemplateSuffix,omitempty"`
	Title                  *string                     `json:"title,omitempty"`
	Vendor                 *string                     `json:"vendor,omitempty"`
	GiftCard               *bool                       `json:"giftCard,omitempty"`
	RedirectNewHandle      *bool                       `json:"redirectNewHandle,omitempty"`
	CollectionsToJoin      []any                       `json:"collectionsToJoin"`
	CollectionsToLeave     []any                       `json:"collectionsToLeave"`
	CombinedListingRole    *CombinedListingsRole       `json:"combinedListingRole,omitempty"`
	ID                     *string                     `json:"id,omitempty"`
	Metafields             []any                       `json:"metafields"`
	ProductOptions         []any                       `json:"productOptions"`
	Status                 *ProductStatus              `json:"status,omitempty"`
	RequiresSellingPlan    *bool                       `json:"requiresSellingPlan,omitempty"`
	ClaimOwnership         *ProductClaimOwnershipInput `json:"claimOwnership,omitempty"`
}

type ProductClaimOwnershipInput struct {
	Bundles *bool `json:"bundles,omitempty"`
}

type SEOInput struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
}

type CustomerInput struct {
	Addresses             []any                               `json:"addresses"`
	Email                 *string                             `json:"email,omitempty"`
	FirstName             *string                             `json:"firstName,omitempty"`
	ID                    *string                             `json:"id,omitempty"`
	LastName              *string                             `json:"lastName,omitempty"`
	Locale                *string                             `json:"locale,omitempty"`
	Metafields            []any                               `json:"metafields"`
	Note                  *string                             `json:"note,omitempty"`
	Phone                 *string                             `json:"phone,omitempty"`
	Tags                  []any                               `json:"tags"`
	EmailMarketingConsent *CustomerEmailMarketingConsentInput `json:"emailMarketingConsent,omitempty"`
	SmsMarketingConsent   *CustomerSmsMarketingConsentInput   `json:"smsMarketingConsent,omitempty"`
	TaxExempt             *bool                               `json:"taxExempt,omitempty"`
	TaxExemptions         []any                               `json:"taxExemptions"`
}

type CustomerEmailMarketingConsentInput struct {
	MarketingOptInLevel *CustomerMarketingOptInLevel `json:"marketingOptInLevel,omitempty"`
	MarketingState      CustomerEmailMarketingState  `json:"marketingState"`
	ConsentUpdatedAt    *string                      `json:"consentUpdatedAt,omitempty"`
}

type CustomerSmsMarketingConsentInput struct {
	MarketingOptInLevel *CustomerMarketingOptInLevel `json:"marketingOptInLevel,omitempty"`
	MarketingState      CustomerSmsMarketingState    `json:"marketingState"`
	ConsentUpdatedAt    *string                      `json:"consentUpdatedAt,omitempty"`
}

type ProductOptionCreateVariantStrategy string

const (
	ProductOptionCreateVariantStrategyLeaveAsIs ProductOptionCreateVariantStrategy = "LEAVE_AS_IS"
	ProductOptionCreateVariantStrategyCreate    ProductOptionCreateVariantStrategy = "CREATE"
)

type ProductOptionUpdateVariantStrategy string

const (
	ProductOptionUpdateVariantStrategyLeaveAsIs ProductOptionUpdateVariantStrategy = "LEAVE_AS_IS"
	ProductOptionUpdateVariantStrategyManage    ProductOptionUpdateVariantStrategy = "MANAGE"
)

type OptionCreateInput struct {
	Name            *string                     `json:"name,omitempty"`
	Position        *int                        `json:"position,omitempty"`
	Values          []any                       `json:"values"`
	LinkedMetafield *LinkedMetafieldCreateInput `json:"linkedMetafield,omitempty"`
}

type OptionValueCreateInput struct {
	Name                 *string `json:"name,omitempty"`
	LinkedMetafieldValue *string `json:"linkedMetafieldValue,omitempty"`
}

type OptionValueUpdateInput struct {
	ID                   string  `json:"id"`
	Name                 *string `json:"name,omitempty"`
	LinkedMetafieldValue *string `json:"linkedMetafieldValue,omitempty"`
}

type LinkedMetafieldCreateInput struct {
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
	Values    []any  `json:"values"`
}

type OptionUpdateInput struct {
	ID              string                      `json:"id"`
	Name            *string                     `json:"name,omitempty"`
	Position        *int                        `json:"position,omitempty"`
	LinkedMetafield *LinkedMetafieldUpdateInput `json:"linkedMetafield,omitempty"`
}

type LinkedMetafieldUpdateInput struct {
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
}

type ProductVariantsBulkCreateStrategy string

const (
	ProductVariantsBulkCreateStrategyDefault                 ProductVariantsBulkCreateStrategy = "DEFAULT"
	ProductVariantsBulkCreateStrategyRemoveStandaloneVariant ProductVariantsBulkCreateStrategy = "REMOVE_STANDALONE_VARIANT"
)

type ProductVariantsBulkInput struct {
	Barcode             *string                        `json:"barcode,omitempty"`
	CompareAtPrice      *string                        `json:"compareAtPrice,omitempty"`
	ID                  *string                        `json:"id,omitempty"`
	MediaSrc            []any                          `json:"mediaSrc"`
	InventoryPolicy     *ProductVariantInventoryPolicy `json:"inventoryPolicy,omitempty"`
	InventoryQuantities []any                          `json:"inventoryQuantities"`
	InventoryItem       *InventoryItemInput            `json:"inventoryItem,omitempty"`
	MediaID             *string                        `json:"mediaId,omitempty"`
	Metafields          []any                          `json:"metafields"`
	OptionValues        []any                          `json:"optionValues"`
	Price               *string                        `json:"price,omitempty"`
	Taxable             *bool                          `json:"taxable,omitempty"`
	TaxCode             *string                        `json:"taxCode,omitempty"`
	RequiresComponents  *bool                          `json:"requiresComponents,omitempty"`
}

type InventoryItemInput struct {
	Sku                          *string                        `json:"sku,omitempty"`
	Cost                         *float64                       `json:"cost,omitempty"`
	Tracked                      *bool                          `json:"tracked,omitempty"`
	CountryCodeOfOrigin          *CountryCode                   `json:"countryCodeOfOrigin,omitempty"`
	HarmonizedSystemCode         *string                        `json:"harmonizedSystemCode,omitempty"`
	CountryHarmonizedSystemCodes []any                          `json:"countryHarmonizedSystemCodes"`
	ProvinceCodeOfOrigin         *string                        `json:"provinceCodeOfOrigin,omitempty"`
	Measurement                  *InventoryItemMeasurementInput `json:"measurement,omitempty"`
	RequiresShipping             *bool                          `json:"requiresShipping,omitempty"`
}

type InventoryItemMeasurementInput struct {
	Weight *WeightInput `json:"weight,omitempty"`
}

type VariantOptionValueInput struct {
	ID                   *string `json:"id,omitempty"`
	Name                 *string `json:"name,omitempty"`
	LinkedMetafieldValue *string `json:"linkedMetafieldValue,omitempty"`
	OptionID             *string `json:"optionId,omitempty"`
	OptionName           *string `json:"optionName,omitempty"`
}

type WeightInput struct {
	Value float64    `json:"value"`
	Unit  WeightUnit `json:"unit"`
}

type MetafieldInput struct {
	ID        *string `json:"id,omitempty"`
	Namespace *string `json:"namespace,omitempty"`
	Key       *string `json:"key,omitempty"`
	Value     *string `json:"value,omitempty"`
	Type      *string `json:"type,omitempty"`
}

type MetafieldsSetInput struct {
	OwnerID       string  `json:"ownerId"`
	Namespace     *string `json:"namespace,omitempty"`
	Key           string  `json:"key"`
	Value         string  `json:"value"`
	CompareDigest *string `json:"compareDigest,omitempty"`
	Type          *string `json:"type,omitempty"`
}

type MetafieldIdentifierInput struct {
	OwnerID   string `json:"ownerId"`
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
}

type CreateMediaInput struct {
	OriginalSource   string           `json:"originalSource"`
	Alt              *string          `json:"alt,omitempty"`
	MediaContentType MediaContentType `json:"mediaContentType"`
}

type MailingAddressInput struct {
	Address1     *string      `json:"address1,omitempty"`
	Address2     *string      `json:"address2,omitempty"`
	City         *string      `json:"city,omitempty"`
	Company      *string      `json:"company,omitempty"`
	CountryCode  *CountryCode `json:"countryCode,omitempty"`
	FirstName    *string      `json:"firstName,omitempty"`
	LastName     *string      `json:"lastName,omitempty"`
	Phone        *string      `json:"phone,omitempty"`
	ProvinceCode *string      `json:"provinceCode,omitempty"`
	Zip          *string      `json:"zip,omitempty"`
}

type FileUpdateInput struct {
	ID                 string  `json:"id"`
	Alt                *string `json:"alt,omitempty"`
	OriginalSource     *string `json:"originalSource,omitempty"`
	PreviewImageSource *string `json:"previewImageSource,omitempty"`
	Filename           *string `json:"filename,omitempty"`
	ReferencesToAdd    []any   `json:"referencesToAdd"`
	ReferencesToRemove []any   `json:"referencesToRemove"`
}
