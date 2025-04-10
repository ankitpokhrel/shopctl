package login

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"

	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/internal/config"
	"github.com/ankitpokhrel/shopctl/internal/oauth"
)

const (
	helpText = "Initiate oAuth login to the Shopify account."
	examples = `$ shopctl auth login`
)

// NewCmdLogin is a login command.
func NewCmdLogin() *cobra.Command {
	return &cobra.Command{
		Use:         "login",
		Short:       "Login to a Shopify account",
		Long:        helpText,
		Example:     examples,
		Annotations: map[string]string{"cmd:main": "true"},
		RunE:        login,
	}
}

func login(cmd *cobra.Command, _ []string) error {
	store, err := cmd.Flags().GetString("store")
	if err != nil {
		return fmt.Errorf("please pass in the store you want to operate on")
	}

	host, alias, err := getHostAndAlias(store)
	if err != nil {
		return err
	}

	authFlow := oauth.NewFlow(host)
	if err := authFlow.Initiate(); err != nil {
		fmt.Printf("\n! Failed to authenticate with Shopify: %s\n", err)
		return err
	}

	shopCfg, err := config.NewShopConfig()
	if err != nil {
		return err
	}
	storeCtx := config.StoreContext{
		Alias: alias,
		Store: host,
	}
	service := fmt.Sprintf("shopctl:%s", cmdutil.GetStoreSlug(store))

	if err := keyring.Set(service, store, authFlow.Token.AccessToken); err != nil {
		fmt.Printf("\n! Failed to save token to a secure storage")
		fmt.Printf("\n! Using insecure plain text storage\n")

		storeCtx.Token = &authFlow.Token.AccessToken
	}

	shopCfg.SetStoreContext(&storeCtx)
	if err := shopCfg.Save(); err != nil {
		return err
	}

	fmt.Printf("\n! Successfully authenticated with Shopify\n")
	return nil
}

func getHostAndAlias(store string) (string, string, error) {
	if !strings.HasPrefix(store, "http://") && !strings.HasPrefix(store, "https://") {
		store = "https://" + store
	}

	myshopifyURL, err := url.Parse(store)
	if err != nil {
		return "", "", err
	}

	host := myshopifyURL.Hostname()
	suffix := ".myshopify.com"

	if !strings.HasSuffix(host, suffix) {
		return "", "", fmt.Errorf("URL is not a valid myshopify domain")
	}
	return host, strings.TrimSuffix(host, suffix), nil
}
