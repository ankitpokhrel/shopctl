import sys
import os
import json
import tarfile
import tempfile
from openai import OpenAI

client = OpenAI()

# --- Extract lightweight product info for review from the exported .tar.gz file ---
def extract_reviewable_products(tar_path, max_items=30):
    temp_dir = tempfile.mkdtemp()
    with tarfile.open(tar_path, "r:gz") as tar:
        def safe_extract_filter(tarinfo, path):
            tarinfo.name = os.path.normpath(tarinfo.name).lstrip(os.sep)
            return tarinfo
        tar.extractall(path=temp_dir, filter=safe_extract_filter)

    product_dir = os.path.join(temp_dir, "products")
    products = []

    for folder in os.listdir(product_dir):
        product_path = os.path.join(product_dir, folder, "product.json")
        variant_path = os.path.join(product_dir, folder, "product_variants.json")

        if os.path.isfile(product_path):
            with open(product_path, "r", encoding="utf-8") as f:
                try:
                    product = json.load(f)
                except json.JSONDecodeError:
                    continue

                product_entry = {
                    "id": str(product.get("id")),
                    "productType": product.get("productType", ""),
                    "tags": product.get("tags", []),
                    "status": product.get("status", ""),
                    "publishedScope": product.get("publishedScope", ""),
                    "vendor": product.get("vendor", ""),
                    "mediaCount": product.get("mediaCount", 0),
                    "variantsCount": product.get("variantsCount", 0),
                    "variants": []
                }

                if os.path.isfile(variant_path):
                    with open(variant_path, "r", encoding="utf-8") as vf:
                        variant_data = json.load(vf)
                        variants = variant_data.get("variants", {}).get("nodes", [])

                        for v in variants:
                            print(v)
                            variant = {
                                "title": v.get("title", ""),
                                "sku": v.get("sku", ""),
                                "price": v.get("price", ""),
                                "inventoryQuantity": v.get("inventoryQuantity", 0),
                                "inventoryPolicy": v.get("inventoryPolicy", ""),
                                "requiresShipping": v.get("requiresShipping", True),
                                "taxable": v.get("taxable", True),
                                "optionValues": v.get("optionValues", []),
                                "availableForSale": v.get("availableForSale", False),
                                "sellableOnlineQuantity": v.get("sellableOnlineQuantity", 0)
                            }
                            product_entry["variants"].append(variant)

                products.append(product_entry)

        if len(products) >= max_items:
            break

    return products

# --- Send data to GPT for catalog insights ---
def generate_catalog_review(products):
    prompt = f"""
You are an expert eCommerce catalog auditor. The product titles and descriptions have already been optimized, so exclude those.

Please review the following product catalog sample and identify:
1. Issues or inconsistencies in tags, product types, or variants
2. Missing or inconsistent inventory information
3. Gaps in product configuration or variant structure
4. Duplicate or overly similar products
5. General recommendations to improve catalog quality and completeness

Respond in clear, concise Markdown.

Sample products:
{json.dumps(products, indent=2)}
"""

    response = client.chat.completions.create(
        model="gpt-4",
        messages=[{"role": "user", "content": prompt}],
        temperature=0.7
    )

    return response.choices[0].message.content.strip()

# --- Main entry ---
def main(tar_path, output_md):
    print("📦 Extracting reviewable product data...")
    products = extract_reviewable_products(tar_path)

    print(f"🧠 Generating catalog review for {len(products)} products...")
    summary = generate_catalog_review(products)

    with open(output_md, "w", encoding="utf-8") as f:
        f.write(summary)

    print("✅ Catalog review saved to:", output_md)

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: review_products.py <input_tar.gz> <output_md>")
        sys.exit(1)
    main(sys.argv[1], sys.argv[2])
