package auth

import (
	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/cmd/auth/login"
)

// NewCmdAuth is an auth command.
func NewCmdAuth() *cobra.Command {
	cmd := cobra.Command{
		Use:         "auth",
		Short:       "Initiate authentication request",
		Long:        "Initiate authentication request to the Shopify host.",
		Annotations: map[string]string{"cmd:main": "true"},
		RunE:        auth,
	}

	cmd.PersistentFlags().StringP("store", "s", "", "Shopify store to login to")

	cmd.AddCommand(
		login.NewCmdLogin(),
	)

	cmd.SetHelpFunc(func(c *cobra.Command, s []string) {
		_ = c.Flags().MarkHidden("context")
		c.Root().HelpFunc()(c, s)
	})

	return &cmd
}

func auth(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}
