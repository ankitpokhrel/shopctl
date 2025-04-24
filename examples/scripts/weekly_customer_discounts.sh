#!/usr/bin/env bash

###############################################################################
# Script: weekly_customer_discounts.sh
#
# Scenario:
# This script identifies high-value customers from the past week and calculates
# personalized discount offers based on their total spend. The goal is to reward
# engaged customers and encourage repeat purchases.
#
# 1. Retrieves up to 20 customers who:
#    - Spent more than $100 in total
#    - Accepted marketing communications
#    - Were updated within the last 24 hours
# 2. For each customer:
#    - Proposes a 20% discount (or 30% if total spent ≥ $200)
#    - Calculates a rounded discount amount based on the spend
#    - Outputs the data to a CSV file including name, email, spend, and discount
# 3. If any eligible customers are found, prints the CSV summary to stdout.
#
# Ideal for weekly CRM workflows or automated marketing campaigns.
###############################################################################

set -euo pipefail
export TZ="Europe/Berlin"

DAY_START=$(date -d '1 days ago' +%Y-%m-%d)

###############################################################################
# 1. Get top 20 customers that spent more than $100 in the last 24 hours and
#    has accepted marketing emails.
###############################################################################
echo "🔍  Scanning customers updated since $DAY_START ..."
customers=$(shopctl customer list --total-spent ">=100" --updated=">=$DAY_START" \
  --accepts-marketing --columns id,first_name,last_name,email,amount_spent \
  --csv --no-headers --with-sensitive-data --limit 20)

if [[ -z "$customers" ]]; then
  echo "🟢  No customers spent more than \$100 since $DAY_START — nothing to do"
  exit 0
fi

echo "id,name,email,spent,proposed_discount_amount" > weekly_customer_discounts.csv

###############################################################################
# 2. Propose a 30% discount amount if the spent >= 200 else 20%.
###############################################################################
while IFS=$',' read -r id fn ln email spent; do
  rate=0.20; (( $(echo "$spent >= 200" | bc) )) && rate=0.30
  coupon=$(awk -v s="$spent" -v r="$rate" 'BEGIN{print int((s*r)+0.999)}')
  echo "\"$fn $ln\",$email,$spent,$coupon" >> weekly_customer_discounts.csv
done <<< "$customers"

###############################################################################
# 3. Print the result.
###############################################################################
if [[ $(wc -l < weekly_customer_discounts.csv) -le 1 ]]; then
    echo "✅  No discounts proposed for any customers"
else
    cat weekly_customer_discounts.csv
fi
