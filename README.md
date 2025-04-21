<div align="center">
    <h1 align="center">ShopCTL</h1>
</div>

<div align="center">
    <p>
        <a href="https://github.com/ankitpokhrel/shopctl/actions?query=workflow%3Abuild+branch%3Amain">
            <img alt="Build" src="https://img.shields.io/github/actions/workflow/status/ankitpokhrel/shopctl/ci.yml?branch=main&style=flat-square" />
        </a>
        <a href="https://goreportcard.com/report/github.com/ankitpokhrel/shopctl">
            <img alt="GO Report-card" src="https://goreportcard.com/badge/github.com/ankitpokhrel/shopctl?style=flat-square" />
        </a><br/>
        <a href="#"><img alt="Linux" src="https://img.shields.io/badge/Linux-%E2%9C%93-dark--green?logo=linux&logoColor=white&style=flat-square" /></a>
        <a href="#"><img alt="macOS" src="https://img.shields.io/badge/macOS-%E2%9C%93-dark--green?logo=apple&style=flat-square" /></a>
        <a href="#"><img alt="Windows" src="https://img.shields.io/badge/Windows-%E2%9C%93WSL-dark--green?logo=windows&style=flat-square" /></a>
    </p>
    <p>
        <i>[WiP] Command line Utility for Shopify Data Management</i>
    </p>
    <img align="center" alt="ShopCTL Demo" src=".github/assets/demo.gif" /><br/><br/>
</div>

ShopCTL is a slightly opinionated, in-progress command-line utility for managing your Shopify store data. It comes with a handful of easy-to-compose commands,
giving you a quick way to interact with your store's data straight from the terminal.

## Installation
1. Create a dummy app from the [Shopify Partners Dashboard](https://partners.shopify.com/) and get the client ID and secret.
   - Add `http://127.0.0.1/shopctl/auth/callback` to the list of Allowed redirection URL(s)
   - Make sure to request for [required scopes](https://github.com/ankitpokhrel/shopctl/blob/main/internal/oauth/oauth.go#L35-L47)

   See https://github.com/ankitpokhrel/shopctl/discussions/3 for detailed instructions on how to setup an app for client ID and secret.

2. Export client ID and secret from the first step to your shell.
   ```sh
   export SHOPCTL_CLIENT_ID=<client-id>
   export SHOPCTL_CLIENT_SECRET=<client-secret>
   ```

3. Install the runnable binary to your `$GOPATH/bin`.

   ```sh
   go install github.com/ankitpokhrel/shopctl/cmd/shopctl@main
   ```

<br/>

The `shopctl` will be available as a packaged downloadable binary for different platforms later.

## Resources
- Check [this post](https://medium.com/@ankitpokhrel/shopctl-a-developer-first-toolkit-for-shopify-automation-d4800f1662bf) for some real life automation and scripting use-cases.
- The [scripts](./scripts/) directory contains examples of how to automate tasks using ShopCTL.
- Learn how to use [ShopCTL in the GitHub CI](https://github.com/ankitpokhrel/shopctl/discussions/6).

## Getting started

### Authentication
Before you begin using ShopCTL, you need to configure your store(s). ShopCTL supports managing multiple Shopify stores by using contexts — each representing a store.

You can either directly use the access token or login to your store using the oAuth flow.

#### Direct Access Token
If you already have an access token you can set it directly to your shell session. Note that your token needs to have [access to resources](https://github.com/ankitpokhrel/shopctl/blob/main/internal/oauth/oauth.go#L35-L47) you're trying to backup/restore.

```
SHOPIFY_ACCESS_TOKEN=<token>
```

#### OAuth
You can log in to your Shopify store using the store's MyShopify URL and would need to assign a unique alias (context name) for that store.
This will generate the config and create a new context in your configuration file.

```sh
# Login to first store
$ shopctl auth login --store mystore1.myshopify.com

# Login to another store
$ shopctl auth login -s mystore2.myshopify.com
```

Logging in to a store creates a context using store id as an alias. The token is saved in a `keychain` if its available, else it will be
stored as a plain text in the config file. The support for `.netrc` might be added later as requested.

### Config Management
ShopCTL lets you manage multiple stores.

The config manager is inspired by the [k8s/kubectl](https://github.com/kubernetes/kubectl) CLI and make use of `context` which represents
a store in our case. Contexts are auto-created when you login to the store with `shopctl auth login` command mentioned above.

We can set default context using the `use-context` command.

```sh
# We already logged in to mystore1 in authentication step
# We're now going to use 'mystore1' as our current store
$ shopctl config use-context store1
```

See `shopctl config -h` for more details on the config command.

### Shell completion
Check `shopctl completion --help` for details on setting up a bash/zsh shell completion.

## Usage
The tool currently comes with product and customer related commands. The flags are [POSIX-compliant](https://www.gnu.org/software/libc/manual/html_node/Argument-Syntax.html).
You can combine available flags in any order to create a unique query. For example, the command below will give you all gift cards on status DRAFT that were created after
2025 and has tags on-sale and premium.

```sh
shopctl product list --gift-card -sDRAFT --tags on-sale,premium --created ">=2025-01-01"
```

## Commands

### Export
The `export` command can be used to extract and save data from your store into external files. You can export multiple resources as products & customers in a single command.
The command supports complex filtering, allowing you to narrow down the exported data using queries.

```sh
$ shopctl export --resource product --output-dir /path/to/dir --name product_export

# Export products and customers from another store
$ shopctl export -c store2 -r product -r customer -o /path/to/dir

# Export premium on-sale products and customers created starting 2025
$ shopctl export -r product="tag:on-sale AND tag:premium" -r customer=created_at:>=2025-01-01 -o /path/to/dir

# You can use 'list' command to prepare filters
$ shopctl export -c mycontext -r product="$(shopctl product list --tags on-sale --type Bags --print-query)" -o /path/to/dir

# Dry run executes the export without creating final files. This will still create files in temporary location.
# Use this option if you want to verify your export without the risk of saving data to the unintented location.
$ shopctl export run -r product="tag:on-sale" --dry-run
```

### Import
The `import` command is designed to restore data into your store from exported files. It allows you to selectively import resources and apply filters to control which items are restored.

```sh
# Import products from the given path
$ shopctl import --resource product --from /path/to/import/dir

# Restore some products on status DRAFT for the context from the latest backup
$ shopctl import -r product="id:id1,id2,id3 AND status:DRAFT" --from /path/to/import/dir

# Restore specific products and verified customers from the latest backup
$ shopctl import -r product="tags:premium,on-sale" -r customer="verifiedemail:true" --from /path/to/import/dir

# Dry run executes the restoration process and print logs without making an actual API call
$ shopctl import -r product --from /path/to/import/dir --dry-run -vvv
```

### Product

#### List
You can search and navigate your products using the `list` command. The command accepts a [Shopify Search query](https://shopify.dev/docs/api/usage/search-syntax) syntax as the first argument.

```sh
# List recent products
$ shopctl product list

# Search for products with specific text anywhere in the product
$ shopctl product list "text in title or description" --limit 20

# List products using combination of raw query and available flags
$ shopctl product list "(title:Caramel Apple) OR (inventory_total:>500 inventory_total:<=1000)" --tags premium

# List products in status DRAFT and ARCHIVED created in 2025
$ shopctl product list -sDRAFT,ARCHIVED --created ">=2025-01-01"

# List products with tag 'on-sale' but without tag 'summer'
$ shopctl product list --tags on-sale,-summer

 # Get products with empty sku and non-empty product type
$ shopctl product list --sku "" --type -

# List products in a plain view without headers
$ shopctl product list --plain --no-headers
```

#### Create
The `create` command lets you create a product. The tool comes with easy-to-use subcommands that you can use to add options and variants to a product.

Check out some examples below.

```sh
# Creating a product could be as simple as this
$ shopctl product create --title "Product title"

# Create active product in the current context
$ shopctl product create -tTitle -d"Product description" --status active

# Create product with tags in the current context
$ shopctl product create -tTitle -d"Product description" --tags tag1,tag2

# Create product in another store
$ shopctl product create -c store2 -tTitle -d"Product description" --type Bags
```

Use the `option add` command to attach options to an existing product.

```sh
$ shopctl product option add <product_id> --name Title --value "Special product"

# Option with multiple values
$ shopctl product option add 8856145494 -nSize -lxs -lsm -lxl

# Set variant strategy to CREATE; default is LEAVE_AS_IS
# With '--create' flag, existing variants are updated with the first option value
# See https://shopify.dev/docs/api/admin-graphql/latest/enums/ProductOptionCreateVariantStrategy
$ shopctl product option add 8856145494 -nStyle -lCasual -lInformal --create
```

Variants provide the ability to offer multiple options (such as different sizes or colors) for a product. You can use the `variant add` command to define these variants.

```sh
# Add a variant 'xs' for color 'Blue'
# In the example below, option 'Color' and 'Size' must exist
$ shopctl product variant add <product_id> -o"Color:Blue" -o"Size:xs"

# Add a variant of price 20, unit cost 10 and compare at price of 30
$ shopctl product variant add 8856145494 -o"Color:Black" -o"Size:s" --price 20 --unit-cost 10 --regular-price 30

# Add a variant with SKU and barcode that requires shipping with inventory tracked
$ shopctl product variant add 8856145494 -o"Color:Red" -o"Size:xl" -p20 --weight "GRAMS:100" --sku 123 --barcode 456 --tracked --requires-shipping
```

See `shopctl product -h` for details on all available commands.

#### Update
The `update` command follows same structure as the create command. Check `shopctl product update -h` for more details.

#### Peek

The `peek` command gives you a quick glance into Shopify product right from your terminal. The source could either be upstream or local imports.

```sh
# Peek by id
$ shopctl peek product <product_id>

# Peek a product from the import folder
# Context and strategy is skipped for direct path
$ shopctl peek product <product_id> --from </path/to/backup>

# Render json output
$ shopctl peek product <product_id> --json
```

### Customer

#### Create
The `create` command lets you create a customer. You can add multiple addresses when creating a customer. The `address` and `meta` flag accepts tagged fields
as shown in the example below.

```sh
# Quickly create a customer
$ shopctl customer create --first-name Jon --last-name Doe

# Create customer with tags and a note
$ shopctl customer create --email janedoe@example.com --note "Example user" --tags example,dummy

# Create customer with multiple addresses (accepts tagged fields)
# See https://shopify.dev/docs/api/admin-graphql/latest/input-objects/MailingAddressInput for the list of accepted tags
$ shopctl customer create -lHolmes --address "221B Baker Street country:GB city:London zip:NW1" --address "country:NP firstname:Jon lastname:Doe"

# Create customer with metafields (accepts tagged fields)
# See https://shopify.dev/docs/apps/build/custom-data/metafields/list-of-data-types#supported-types for valid metafield types
$ shopctl customer create -fJane -lDoe --meta "custom.preferred_color:#95BF47 type:color"
```

#### Update
The `update` command follows same structure as the create command. However, you can only update default address when updating. This will be addressed in the future.
Check `shopctl customer update -h` for more details.

#### Delete
You can delete a customer by its ID, email or phone using the `delete` command.

```sh
# Delete customer by its ID
$ shopctl customer delete 8370159190
$ shopctl customer delete gid://shopify/Customer/8370159190

# Delete customer by its email
$ shopctl customer delete --email example@domain.com

# Delete customer by its phone number
$ shopctl customer delete --phone +1234567890
```

## Development
1. Clone the repo.
   ```sh
   git clone git@github.com:ankitpokhrel/shopctl.git
   ```

2. Setup a dummy Shopify App since we need app tokens for development.
   ```sh
   export SHOPCTL_CLIENT_ID=<client-id>
   export SHOPCTL_CLIENT_SECRET=<client-secret>
   ```

4. Make changes, build the binary, and test your changes.
   ```sh
   make deps install
   ```

5. Run CI steps locally before submitting a PR.
   ```sh
   make ci
   ```
