## ShopCTL

<div align="center">
    <h1 align="center">ShopCTL</h1>
</div>

<div>
    <p align="center">
        <i>[WiP] Fuss-free command line utility for Shopify</i>
    </p>
    <img align="center" alt="ShopCTL Demo" src=".github/assets/demo.gif" /><br/><br/>
</div>

ShopCTL is a slightly opinionated, in-progress command-line utility for managing your Shopify data. It comes with a handful of easy-to-compose commands,
giving you a quick way to interact with your store's data straight from the terminal.

## Installation

Create a dummy app from the [Shopify partners dashboard](https://partners.shopify.com/) and get the client ID and secret. Make sure you've [required scopes](https://github.com/ankitpokhrel/shopctl/blob/main/internal/oauth/oauth.go#L35-L47).

Install the runnable binary to your `$GOPATH/bin`.

```go
SHOPCTL_CLIENT_ID=<id> SHOPCTL_CLIENT_SECRET=<secret> go install github.com/ankitpokhrel/shopctl/cmd/shopctl@main
```

`shopctl` will be available as a packaged downloadable binary for different platforms later.

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

Logging in to a store creates a context using store id as an alias.
