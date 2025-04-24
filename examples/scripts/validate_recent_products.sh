#!/usr/bin/env bash

###############################################################################
# Script: validate_recent_products.sh
#
# Scenario:
# This script validates recently updated Shopify products to ensure variant data
# quality. It performs the following checks:
#
# 1. Scans for up to 100 products updated in the last 24 hours.
# 2. For each product, iterates through its variants and flags:
#    - Empty SKUs
#    - Prices outside the accepted range ($25–$100)
# 3. Outputs a CSV report of invalid variants with reasons for failure.
# 4. If any invalid variants are found, prints the report and exits with error.
#
# Intended for CI/CD pipelines or scheduled validation jobs to catch common
# catalog issues before publishing or syncing data.
###############################################################################

set -euo pipefail
export TZ="Europe/Berlin"

DAY_START=$(date -d '1 day ago' +%Y-%m-%d)

# Helper that appends a validation message to the global `reason` variable.
add_reason() {
  if [[ -z "$reason" ]]; then
    reason="$1"
  else
    reason+="; $1"
  fi
}

###############################################################################
# 1. Get 100 products that changed in the last 24h.
###############################################################################
echo "🔍  Scanning products updated since $DAY_START ..."
products=$(shopctl product list --columns id,status --updated ">=$DAY_START" --csv --no-headers --limit 100)

if [[ -z "$products" ]]; then
  echo "🟢  No products updated since $DAY_START — nothing to validate"
  exit 0
fi

echo "product_id,variant_id,variant_title,sku,price,status,reason" > invalid_products.csv
invalid=0

###############################################################################
# 2. For each product, check variants for empty SKU OR price out of range.
###############################################################################
while IFS=',' read -r id status; do
  pid="gid://shopify/Product/${id}"

  variants=$(shopctl product variant list "$pid" --columns id,title,sku,price --csv --no-headers)
  [[ -z "$variants" ]] && continue

  while IFS=',' read -r vid_raw title sku price; do
    vid="gid://shopify/ProductVariant/${vid_raw}"
    reason=""

    [[ -z "$sku" ]] && add_reason "Empty SKU"

    out_of_range=$(echo "$price < 25 || $price > 100" | bc || true)
    [[ $out_of_range -eq 1 ]] && add_reason "Price $price out of \$25–\$100"

    if [[ -n "$reason" ]]; then
      {
          printf '%s,%s,"%s",%s,%s,%s,%s\n' \
              "$pid" "$vid" "$title" "$sku" "$price" "$status" "$reason"
      } >> invalid_products.csv
      invalid=$((invalid + 1))
    fi
  done <<< "$variants"
done <<< "$products"

###############################################################################
# 3. Print the result.
###############################################################################
if (( invalid )); then
  echo "::error::${invalid} invalid variant(s) found"
  cat invalid_products.csv
  exit 1
else
  echo "✅  All checked variants passed"
fi

