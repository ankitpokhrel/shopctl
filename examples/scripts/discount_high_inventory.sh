#!/usr/bin/env bash

###############################################################################
# Script: discount_high_inventory.sh
#
# Scenario:
# This script identifies high-inventory Shopify products recently updated, and
# applies a conditional 10% discount to their variants. The goal is to optimize
# pricing for overstocked items without compromising acceptable profit margins.
#
# 1. Scans for products updated in the last 24 hours with inventory_total ≥ 300.
# 2. For each variant, calculates a 10% discounted price and applies it only if:
#      new_price ≥ unit_cost × 1.15 (to preserve at least 15% margin)
# 3. Outputs a CSV summary of all discounted variants including old and new prices.
# 4. If no discounts were applied, reports that and exits cleanly.
#
# Intended for use in automated price adjustment workflows and clearance routines.
###############################################################################

set -euo pipefail
export TZ="Europe/Berlin"

DAY_START=$(date -d '1 day ago' +%Y-%m-%d)

###############################################################################
# 1. Find products updated in the last 24 hours and has large inventory.
###############################################################################
echo "🔍  Scanning products with high inventory since $DAY_START ..."
products=$(shopctl product list "inventory_total:>=300" --updated=">=$DAY_START" --columns id,title --csv --no-headers)

if [[ -z "$products" ]]; then
  echo "🟢  No high inventory updated products since $DAY_START — nothing to do"
  exit 0
fi

# Create CSV summary file
echo "product_id,product_title,variant_id,old_price,new_price" > inventory_discounts.csv

###############################################################################
# 2. Apply a 10% discount only if the resulting price is still ≥ cost × 1.15.
###############################################################################
 while IFS=',' read -r pid title; do
  variants=$(shopctl product variant list $pid --columns id,price,unit_cost --csv --no-headers)
  [[ -z "$variants" ]] && continue

  while IFS=',' read -r variant_id price unit_cost; do
    new_price=$(echo "scale=2; $price * 0.9" | bc) # 10% discount
    margin_ok=$(echo "$new_price >= ($unit_cost*1.15)" | bc)

    if [[ $margin_ok -eq 1 ]]; then
      shopctl product variant edit $pid --id $variant_id --price "$new_price"
      echo "$pid,\"$title\",$variant_id,$price,$new_price" >> inventory_discounts.csv
    fi
  done <<< "$variants"
done <<< "$products"

###############################################################################
# 3. Print the result.
###############################################################################
if [[ $(wc -l < inventory_discounts.csv) -le 1 ]]; then
    echo "✅  No discounts added to any of the products"
else
    cat inventory_discounts.csv
fi

