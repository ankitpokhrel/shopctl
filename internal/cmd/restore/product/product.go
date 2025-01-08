package product

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl/engine"
	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/pkg/file"
	"github.com/ankitpokhrel/shopctl/pkg/tlog"
	"github.com/ankitpokhrel/shopctl/schema"
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
	foundFiles, err := file.FindFilesInDir(path, fmt.Sprintf("%s.json", engine.Product))
	if err != nil {
		return err
	}

	for f := range foundFiles {
		if f.Err != nil {
			lgr.Warn("Skipping file due to read err", "file", f.Path, "error", f.Err)
			continue
		}

		productFn := func() (any, error) {
			return handleProductRestore(client, f)
		}

		eng.Add(engine.Product, engine.ResourceCollection{
			engine.NewResource(engine.Product, path, productFn),
		})
	}

	return nil
}

func handleProductRestore(client *api.GQLClient, f file.File) (*api.ProductCreateResponse, error) {
	product, err := file.ReadFileContents(f.Path)
	if err != nil {
		lgr.Error("Unable to read contents", "file", f.Path, "error", err)
		return nil, err
	}

	var prod schema.Product
	if err = json.Unmarshal(product, &prod); err != nil {
		lgr.Error("Unable to marshal contents", "file", f.Path, "error", err)
		return nil, err
	}

	// TODO: Handle/log error.
	res, err := createOrUpdateProduct(&prod, client)
	if err != nil {
		return nil, err
	}
	if len(res.Errors) > 0 {
		return res, fmt.Errorf("errors occurred while restoring product: %+v", res.Errors)
	}
	return res, nil
}

func createOrUpdateProduct(product *schema.Product, client *api.GQLClient) (*api.ProductCreateResponse, error) {
	res, err := client.CheckProductByID(product.ID)
	if err != nil {
		return nil, err
	}

	// TODO: Compare and extract fields that are actually updated.

	var category *string
	if product.Category != nil {
		category = &product.Category.Name
	}

	if res.Data.Product.ID != "" {
		// Product exists, execute update mutation.
		input := schema.ProductUpdateInput{
			ID:                     &product.ID,
			DescriptionHtml:        &product.DescriptionHtml,
			Handle:                 &product.Handle,
			Seo:                    &schema.SEOInput{Title: product.Seo.Title, Description: product.Seo.Description},
			ProductType:            &product.ProductType,
			Category:               category,
			Tags:                   product.Tags,
			TemplateSuffix:         product.TemplateSuffix,
			GiftCardTemplateSuffix: product.GiftCardTemplateSuffix,
			Title:                  &product.Title,
			Vendor:                 &product.Vendor,
			CollectionsToJoin:      nil,
			Status:                 &product.Status,
			RequiresSellingPlan:    &product.RequiresSellingPlan,
		}
		lgr.Warn("Product already exists, updating", "id", product.ID)
		return client.UpdateProduct(input)
	}

	// Product does not exist, execute create mutation.
	input := schema.ProductCreateInput{
		DescriptionHtml:        &product.DescriptionHtml,
		Handle:                 &product.Handle,
		Seo:                    &schema.SEOInput{Title: product.Seo.Title, Description: product.Seo.Description},
		ProductType:            &product.ProductType,
		Category:               category,
		Tags:                   product.Tags,
		TemplateSuffix:         product.TemplateSuffix,
		GiftCardTemplateSuffix: product.GiftCardTemplateSuffix,
		Title:                  &product.Title,
		Vendor:                 &product.Vendor,
		GiftCard:               &product.IsGiftCard,
		CollectionsToJoin:      nil,
		CombinedListingRole:    product.CombinedListingRole,
		Status:                 &product.Status,
		RequiresSellingPlan:    &product.RequiresSellingPlan,
	}
	lgr.Info("Creating product", "id", product.ID)
	return client.CreateProduct(input)
}
