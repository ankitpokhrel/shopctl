package api

const (
	fieldsProduct = `id
 title
 handle
 description
 descriptionHtml
 productType
 isGiftCard
 status
 category {
  	id
    name
    fullName
    isArchived
    isLeaf
    isRoot
    parentId
 }
 tags
 totalInventory
 tracksInventory
 createdAt
 updatedAt
 publishedAt
 combinedListingRole
 defaultCursor
 giftCardTemplateSuffix
 hasOnlyDefaultVariant
 hasOutOfStockVariants
 hasVariantsThatRequiresComponents
 legacyResourceId
 onlineStorePreviewUrl
 onlineStoreUrl
 requiresSellingPlan
 templateSuffix
 vendor
 options {
   name
   values
   position
   optionValues {
     id
     name
     hasVariants
   }
 }`

	fieldsVariant = `id
 title
 displayName
 price
 sku
 position
 availableForSale
 barcode
 compareAtPrice
 inventoryQuantity
 sellableOnlineQuantity
 requiresComponents
 taxable
 taxCode
 createdAt
 updatedAt`

	fieldsMedia = `id
alt
status
preview {
  image {
    altText
    url
    height
    width
  }
  status
}
mediaContentType
mediaErrors { details }
mediaWarnings { message }`
)
