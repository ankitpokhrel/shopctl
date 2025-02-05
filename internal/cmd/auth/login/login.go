package login

import (
	"fmt"

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
	cmd := cobra.Command{
		Use:         "login",
		Short:       "Login to a Shopify account",
		Long:        helpText,
		Example:     examples,
		Annotations: map[string]string{"cmd:main": "true"},
		RunE:        login,
	}

	cmd.Flags().StringP("alias", "a", "", "Store alias")

	return &cmd
}

func login(cmd *cobra.Command, _ []string) error {
	store, err := cmd.Flags().GetString("store")
	if err != nil {
		return fmt.Errorf("please pass in the store you want to operate on")
	}

	alias, err := cmd.Flags().GetString("alias")
	if err != nil || alias == "" {
		return fmt.Errorf("please add a unique alias for the store")
	}

	authFlow := oauth.NewFlow(store)
	if err := authFlow.Initiate(); err != nil {
		fmt.Printf("\n! Failed to authenticate with Shopify: %s\n", err)
		return err
	}

	shopCfg, err := config.NewShopConfig()
	if err != nil {
		return err
	}
	storeCfg := config.NewStoreConfig(store, alias)
	storeCtx := config.StoreContext{
		Alias: alias,
		Store: store,
	}
	service := fmt.Sprintf("shopctl:%s", cmdutil.GetStoreSlug(store))

	if err := keyring.Set(service, store, authFlow.Token.AccessToken); err != nil {
		fmt.Printf("\n! Failed to save token to a secure storage")
		fmt.Printf("\n! Using insecure plain text storage\n")

		storeCtx.Token = &authFlow.Token.AccessToken
	}

	if err := storeCfg.Save(); err != nil {
		return err
	}

	shopCfg.SetStoreContext(&storeCtx)
	if err := shopCfg.Save(); err != nil {
		return err
	}

	fmt.Printf("\n! Successfully authenticated with Shopify\n")
	return nil
}
