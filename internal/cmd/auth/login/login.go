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

	authFlow := oauth.NewFlow(store)
	if err := authFlow.Initiate(); err != nil {
		fmt.Printf("\n! Failed to authenticate with Shopify: %s\n", err)
		return err
	}

	service := fmt.Sprintf("shopctl:%s", cmdutil.GetStoreSlug(store))
	storeCfg := config.NewStoreConfig(store)

	if err := keyring.Set(service, store, authFlow.Token.AccessToken); err != nil {
		fmt.Printf("\n! Failed to save token to a secure storage")
		fmt.Printf("\n! Using insecure plain text storage\n")

		storeCfg.SetToken(authFlow.Token.AccessToken)
	}

	if err := storeCfg.Save(); err != nil {
		return err
	}

	fmt.Printf("\n! Successfully authenticated with Shopify\n")
	return nil
}
