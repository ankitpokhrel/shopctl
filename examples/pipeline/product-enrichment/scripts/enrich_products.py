import sys
import os
import json
import csv
import tarfile
import tempfile

from openai import OpenAI
from time import sleep

client = OpenAI()
BATCH_SIZE = 15

# --- Function Schema for GPT ---
function_def = {
    "name": "enrich_products",
    "description": "Generate enriched product fields for a batch of Shopify products.",
    "parameters": {
        "type": "object",
        "properties": {
            "products": {
                "type": "array",
                "items": {
                    "type": "object",
                    "properties": {
                        "product_id": {"type": "string"},
                        "new_title": {"type": "string"},
                        "seo_title": {"type": "string"},
                        "seo_description": {"type": "string"},
                    },
                    "required": ["product_id", "new_title", "seo_title", "seo_description"]
                }
            }
        },
        "required": ["products"]
    }
}

# --- Extract files from the exported .tar.gz file ---
def extract_products_from_tar(tar_path):
    temp_dir = tempfile.mkdtemp()
    with tarfile.open(tar_path, "r:gz") as tar:
        def safe_extract_filter(tarinfo, path):
            tarinfo.name = os.path.normpath(tarinfo.name).lstrip(os.sep)
            return tarinfo
        tar.extractall(path=temp_dir, filter=safe_extract_filter)

    product_dir = os.path.join(temp_dir, "products")
    if not os.path.isdir(product_dir):
        raise FileNotFoundError("products/ folder not found in archive")

    product_data = []
    for folder in os.listdir(product_dir):
        path = os.path.join(product_dir, folder, "product.json")
        if os.path.isfile(path):
            with open(path, "r", encoding="utf-8") as f:
                try:
                    product = json.load(f)
                    product_data.append(product)
                except json.JSONDecodeError:
                    print(f"⚠️  Skipping {folder}: invalid JSON")
    return product_data

# --- Prompt builder (user content) ---
def build_batch_prompt(batch):
    return f"""
Enrich each product below by generating:
- new_title: improved title (if needed)
- seo_title: keyword-rich, short title for SEO
- seo_description: 1–2 sentence summary for search engines

Use the 'enrich_products' function to return a JSON array of enriched results.

Input:
{json.dumps([
    {
        "product_id": str(p["id"]),
        "title": p.get("title", ""),
        "description": p.get("body_html", ""),
        "tags": p.get("tags", []),
    } for p in batch
], indent=2)}
"""

# --- GPT-4 Function Call API ---
def enrich_batch(batch):
    prompt = build_batch_prompt(batch)

    try:
        response = client.chat.completions.create(
            model="gpt-4-0613",
            messages=[
                {"role": "user", "content": prompt}
            ],
            tools=[{
                "type": "function",
                "function": function_def
            }],
            tool_choice={
                "type": "function",
                "function": {"name": "enrich_products"}
            },
        )

        args = json.loads(response.choices[0].message.tool_calls[0].function.arguments)
        return args["products"]

    except Exception as e:
        print(f"[ERROR] Failed to enrich batch: {e}")
        sleep(2)
        return []

# --- Main ---
def main(tar_path, output_csv):
    print("📦 Reading archive:", tar_path)
    products = extract_products_from_tar(tar_path)

    enriched_rows = []
    for i in range(0, len(products), BATCH_SIZE):
        batch = products[i:i + BATCH_SIZE]
        print(f"✨ Enriching products {i+1} to {i+len(batch)}...")
        enriched_rows.extend(enrich_batch(batch))

    print(f"💾 Writing enriched data to: {output_csv}")
    with open(output_csv, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=[
            "product_id", "new_title", "seo_title", "seo_description"
        ])
        writer.writeheader()
        writer.writerows(enriched_rows)

    print("✅ Done. Enriched", len(enriched_rows), "products.")

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: enrich_products.py <input_tar.gz> <output_csv>")
        sys.exit(1)
    main(sys.argv[1], sys.argv[2])
