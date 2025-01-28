package product

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/engine"
	"github.com/ankitpokhrel/shopctl/internal/registry"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
)

const (
	helpText = `Product initiates a product restoration process.

Use this command to restore the entire product catalog or a subset of data.`

	examples = `$ shopctl restore product --from </path/to/bkp>`
)

var lgr *tlog.Logger

// NewCmdProduct creates a new product restore command.
func NewCmdProduct(eng *engine.Engine) *cobra.Command {
	return &cobra.Command{
		Use:     "product BACKUP_PATH",
		Short:   "Product initiates a product restoration process",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"products"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Type assert engine.doer with engine.Restore type.
			_ = eng.Doer().(*engine.Restore)

			client := cmd.Context().Value("gqlClient").(*api.GQLClient)
			bkpPath := args[0]

			v, _ := cmd.Flags().GetCount("verbose")
			lgr = tlog.New(tlog.VerboseLevel(v))

			return restore(eng, bkpPath, client)
		},
	}
}

func restore(eng *engine.Engine, path string, client *api.GQLClient) error {
	eng.Register(engine.Product)
	restoreStart := time.Now()

	go func() {
		defer eng.Done(engine.Product)

		// TODO: Handle/log error.
		_ = restoreProduct(eng, client, path)
	}()

	for res := range eng.Run(engine.Product) {
		if res.Err != nil {
			lgr.Errorf("Failed to restore resource %s: %v\n", res.ResourceType, res.Err)
		}
	}

	lgr.V(tlog.VL3).Infof(
		"Product restoration complete in %v",
		time.Since(restoreStart),
	)
	return nil
}

func restoreProduct(eng *engine.Engine, client *api.GQLClient, path string) error {
	foundFiles, err := registry.FindFilesInDir(path, fmt.Sprintf("%s.json", engine.Product))
	if err != nil {
		return err
	}

	for f := range foundFiles {
		if f.Err != nil {
			lgr.Warn("Skipping file due to read err", "file", f.Path, "error", f.Err)
			continue
		}

		productFn := &productHandler{client: client, file: f}

		eng.Add(engine.Product, engine.ResourceCollection{
			engine.NewResource(engine.Product, path, productFn),
		})
	}

	return nil
}
