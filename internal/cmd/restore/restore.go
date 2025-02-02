package restore

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmd/restore/run"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
)

const helpText = `Restore initiates data restoration process.

You can either restore an entire store or a filtered subset, including products, customers and orders.`

// NewCmdRestore creates a new restore command.
func NewCmdRestore() *cobra.Command {
	var (
		store string
		err   error
	)

	cmd := cobra.Command{
		Use:         "restore",
		Short:       "Restore initiates a data restoration process",
		Long:        helpText,
		Annotations: map[string]string{"cmd:main": "true"},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Parent().Name() != "restore" {
				return nil
			}

			store, err = cmd.Flags().GetString("store")
			if err != nil {
				return err
			}

			v, _ := cmd.Flags().GetCount("verbose")
			lgr := tlog.New(tlog.VerboseLevel(v))

			gqlClient := api.NewGQLClient(store)
			cmd.SetContext(context.WithValue(cmd.Context(), "gqlClient", gqlClient))
			cmd.SetContext(context.WithValue(cmd.Context(), "logger", lgr))

			return nil
		},
		RunE: restore,
	}

	cmd.AddCommand(
		run.NewCmdRun(),
	)

	return &cmd
}

func restore(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}
