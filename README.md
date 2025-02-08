## ShopCTL

ShopCTL is a work-in-progress command line utility for Shopify data management. It enables you to backup and restore Shopify resources locally,
eliminating the need to route your data through third-party plugins. It comes with a few handy commands to quickly peek and inspect your data,
and gives you a lightweight, no-nonsense way to manage your store's data straight from the terminal.

> This tool is currently a work in progress and not production ready.

<a href="#"><img alt="Linux" src="https://img.shields.io/badge/Linux-%E2%9C%93-dark--green?logo=linux&logoColor=white&style=flat-square" /></a>
<a href="#"><img alt="macOS" src="https://img.shields.io/badge/macOS-%E2%9C%93-dark--green?logo=apple&style=flat-square" /></a>
<a href="#"><img alt="Windows" src="https://img.shields.io/badge/Windows-%E2%9C%93-dark--green?logo=windows&style=flat-square" /></a>

## Installation

Install the runnable binary to your `$GOPATH/bin`.

```go
go install github.com/ankitpokhrel/shopctl/cmd/shopctl@main
```

`shopctl` will be available as a packaged downloadable binary for different platforms later.

## Getting started

Before you begin using ShopCTL, you need to configure your store(s). ShopCTL supports managing multiple Shopify stores by using contexts
— each representing a store — and by allowing you to define backup strategies for each store.

#### 1. Login to your store

Log in to your Shopify store using its MyShopify URL and assign a unique alias (context name) for that store.
This will create a new context in your configuration file.

```sh
# Login to first store
$ shopctl auth login --store mystore1.myshopify.com --alias store1

# Login to anothe store
$ shopctl auth login --store mystore2.myshopify.com --alias store2
```

#### 2. Set current active context

Switch to the context corresponding to the store you want to work with. For example, to set `store1` as the current active context:

```sh
shopctl config use-context store1
```

#### 3. Define a backup strategy

A backup strategy defines what resources to back up, where to store the backups, and additional parameters such as backup type or prefix.
With your desired context active (in this case, `store1`), add one or more backup strategies as shown below.

```sh
# Daily product backup
$ shopctl config set-strategy daily --dir "/path/to/backups/daily" --resources product

# Weekly full backup for product and customer
$ shopctl config set-strategy weekly -d "/path/to/backups/weekly" -r product,customer --type full --prefix wk_
```

#### 4. Set current strategy

Let's set current active strategy to `daily`
```sh
shopctl config use-strategy daily
```

#### 5. Run the backup

With the active context and strategy configured, you can now run your backup. The command below will execute the backup for `store1`
using the `daily` strategy we defined above:

```sh
shopctl backup run
```
