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

	fieldsMetaFields = `id
namespace
key
jsonValue
description
type
ownerType
createdAt
updatedAt`

	fieldsCustomer = `id
firstName
lastName
displayName
email
verifiedEmail
validEmailAddress
phone
tags
taxExempt
state
note
numberOfOrders
emailMarketingConsent {
  consentUpdatedAt
  marketingOptInLevel
  marketingState
}
smsMarketingConsent {
  consentUpdatedAt
  marketingOptInLevel
  marketingState
  consentCollectedFrom
}
addressesV2(first: 100) {
  nodes {
    id
    address1
    address2
    zip
    city
    country
    countryCodeV2
    firstName
    lastName
    company
    province
    provinceCode
  }
}
amountSpent {
  amount
  currencyCode
}
createdAt
updatedAt`
)
