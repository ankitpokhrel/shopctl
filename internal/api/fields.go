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
 seo {
    title
    description
}
 options {
   id
   name
   position
   linkedMetafield {
     key
     namespace
   }
   optionValues {
     id
     name
     hasVariants
     linkedMetafieldValue
  }
}
compareAtPriceRange{
  maxVariantCompareAtPrice{
    amount
    currencyCode
  }
  minVariantCompareAtPrice{
     amount
     currencyCode
  }
}
priceRangeV2{
  maxVariantPrice{
    amount
    currencyCode
  }
  minVariantPrice{
    amount
    currencyCode
  }
}`

	fieldsVariant = `id
legacyResourceId
title
displayName
sku
position
price
unitPriceMeasurement {
  measuredType
  quantityUnit
  quantityValue
  referenceUnit
  referenceValue
}
image {
  id
  altText
  url
  height
  width
}
availableForSale
barcode
compareAtPrice
sellableOnlineQuantity
requiresComponents
taxable
taxCode
selectedOptions {
  name
  optionValue {
    id
    linkedMetafieldValue
  }
  value
}
sellingPlanGroupsCount {
  count
  precision
}
inventoryPolicy
inventoryQuantity
inventoryItem {
  id
  legacyResourceId
  sku
  countryCodeOfOrigin
  provinceCodeOfOrigin
  harmonizedSystemCode
  duplicateSkuCount
  locationsCount{
    count
    precision
  }
  inventoryHistoryUrl
  measurement {
    id
    weight {
      unit
      value
    }
  }
  requiresShipping
  tracked
  trackedEditable {
    locked
    reason
  }
  unitCost {
    amount
    currencyCode
  }
  createdAt
  updatedAt
}
createdAt
updatedAt`

	fieldsMedia = `id
alt
status
preview {
  image {
    id
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

	fieldsMetafields = `id
namespace
key
jsonValue
value
type
description
ownerType
createdAt
updatedAt`

	fieldsCustomer = `id
legacyResourceId
firstName
lastName
displayName
email
verifiedEmail
validEmailAddress
phone
tags
taxExempt
taxExemptions
state
note
numberOfOrders
image {
  id
  altText
  url
  height
  width
}
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
defaultAddress {
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
amountSpent {
  amount
  currencyCode
}
statistics {
    __typename
  predictedSpendTier
}
canDelete
dataSaleOptOut
productSubscriberStatus
unsubscribeUrl
createdAt
updatedAt`

	fieldsWebhook = `id
topic
format
apiVersion { handle }
endpoint {
  __typename
    ... on WebhookHttpEndpoint { callbackUrl }
    ... on WebhookEventBridgeEndpoint { arn }
    ... on WebhookPubSubEndpoint { pubSubProject pubSubTopic }
}
createdAt
updatedAt`
)
