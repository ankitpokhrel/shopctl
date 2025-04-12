package peek

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	html2md "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/charmbracelet/glamour"
	"github.com/fatih/color"

	"github.com/ankitpokhrel/shopctl"
	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/schema"
)

const (
	notAvailable = "n/a"

	spacing   = 2
	padXSmall = 5
	padSmall  = 7
	padMid    = 10

	defaultIDLen    = 12
	defaultTitleLen = 33
)

type fragment struct {
	Body  string
	Parse bool
}

func newBlankFragment(n int) fragment {
	var buf strings.Builder
	for range n {
		buf.WriteRune('\n')
	}
	return fragment{
		Body:  buf.String(),
		Parse: false,
	}
}

// Formatter converts struct to a markdown.
type Formatter struct {
	store   string
	product *schema.Product
}

// NewFormatter creates a new formatter.
func NewFormatter(store string, product *schema.Product) Formatter {
	return Formatter{store: store, product: product}
}

// Render renders the view.
func (f Formatter) Render() error {
	if cmdutil.IsDumbTerminal() || cmdutil.IsNotTTY() {
		return f.renderPlain(os.Stdout)
	}
	r, err := cmdutil.MDRenderer()
	if err != nil {
		return err
	}
	out, err := f.renderedOut(r)
	if err != nil {
		return err
	}
	return cmdutil.PagerOut(out)
}

// renderedOut translates raw data to the format we want to display in.
func (f Formatter) renderedOut(renderer *glamour.TermRenderer) (string, error) {
	var res strings.Builder

	for _, p := range f.fragments() {
		if p.Parse {
			out, err := renderer.Render(p.Body)
			if err != nil {
				return "", err
			}
			res.WriteString(out)
		} else {
			res.WriteString(p.Body)
		}
	}

	return res.String(), nil
}

func (f Formatter) String() string {
	var s strings.Builder

	s.WriteString(f.header())

	desc := f.description()
	if desc != "" {
		s.WriteString(fmt.Sprintf("\n\n%s\n\n%s", f.separator("Product Description"), desc))
	}
	s.WriteString(f.footer())

	return s.String()
}

func (f Formatter) fragments() []fragment {
	scraps := []fragment{
		{Body: f.header(), Parse: true},
	}

	desc := f.description()
	if desc != "" {
		scraps = append(
			scraps,
			newBlankFragment(1),
			fragment{Body: f.separator("Product Description")},
			newBlankFragment(2),
			fragment{Body: desc, Parse: true},
		)
	}

	if len(f.product.Options) > 0 {
		scraps = append(
			scraps,
			newBlankFragment(1),
			fragment{Body: f.separator("Product Options")},
			newBlankFragment(2),
			fragment{Body: f.options()},
			newBlankFragment(1),
		)
	}

	if len(f.product.Variants.Nodes) > 0 {
		scraps = append(
			scraps,
			newBlankFragment(1),
			fragment{Body: f.separator("Product Variants")},
			newBlankFragment(2),
			fragment{Body: f.variants()},
			newBlankFragment(1),
		)
	}

	if len(f.product.Media.Nodes) > 0 {
		scraps = append(
			scraps,
			newBlankFragment(1),
			fragment{Body: f.separator("Product Media")},
			newBlankFragment(2),
			fragment{Body: f.media()},
			newBlankFragment(1),
		)
	}

	return append(scraps, newBlankFragment(1), fragment{Body: f.footer()}, newBlankFragment(2))
}

func (f Formatter) header() string {
	var (
		iconID              = "🆔"
		iconStatus          string
		iconCreated         = "🕒"
		iconUpdated         = "🔄"
		iconPublished       = "📅"
		iconVendor          = "🏭"
		iconCategory        = "📂"
		iconTags            = "🏷️"
		iconProductType     = "📦"
		iconTracksInventory = "📊"

		id              string
		status          string
		created         string
		published       string
		updated         string
		vendor          string
		category        string
		tags            string
		productType     string
		tracksInventory string
	)

	id = fmt.Sprintf("%s %s", iconID, shopctl.ExtractNumericID(f.product.ID))

	switch f.product.Status {
	case schema.ProductStatusDraft:
		iconStatus = "📝"
	case schema.ProductStatusArchived:
		iconStatus = "🗑️"
	case schema.ProductStatusActive:
		iconStatus = "✅"
	}
	status = fmt.Sprintf("%s %s", iconStatus, f.product.Status)

	created = fmt.Sprintf("%s %s", iconCreated, cmdutil.FormatDateTimeHuman(f.product.CreatedAt, time.RFC3339))
	updated = fmt.Sprintf("%s %s", iconUpdated, cmdutil.FormatDateTimeHuman(f.product.UpdatedAt, time.RFC3339))
	if f.product.PublishedAt != nil {
		published = fmt.Sprintf("%s %s", iconPublished, cmdutil.FormatDateTimeHuman(*f.product.PublishedAt, time.RFC3339))
	} else {
		published = fmt.Sprintf("%s %s", iconPublished, notAvailable)
	}
	vendor = fmt.Sprintf("%s %s", iconVendor, f.product.Vendor)

	if f.product.TracksInventory {
		tracksInventory = fmt.Sprintf("%s %s", iconTracksInventory, "Yes")
	} else {
		tracksInventory = fmt.Sprintf("%s %s", iconTracksInventory, "No")
	}

	if f.product.ProductType != "" {
		productType = fmt.Sprintf("%s %s", iconProductType, f.product.ProductType)
	} else {
		productType = fmt.Sprintf("%s %s", iconProductType, notAvailable)
	}

	if f.product.Category != nil {
		category = fmt.Sprintf("%s %s", iconCategory, f.product.Category.Name)
	} else {
		category = fmt.Sprintf("%s %s", iconCategory, notAvailable)
	}

	productTags := make([]string, 0, len(f.product.Tags))
	for _, t := range f.product.Tags {
		productTags = append(productTags, fmt.Sprintf("%s", t))
	}
	if len(productTags) == 0 {
		productTags = append(productTags, notAvailable)
	}
	tags = fmt.Sprintf("%s  %s", iconTags, strings.Join(productTags, ", "))

	return fmt.Sprintf(
		"%s  %s  %s  %s  %s  %s\n# %s\n> %s\n\n%s  %s  %s  %s",
		status,
		id,
		created,
		updated,
		published,
		vendor,
		f.product.Title,
		f.product.Handle,
		tracksInventory,
		productType,
		category,
		tags,
	)
}

func (f Formatter) description() string {
	desc, err := html2md.ConvertString(f.product.DescriptionHtml)
	if err != nil {
		desc = f.product.Description
	}
	return desc
}

func (f Formatter) options() string {
	var s strings.Builder

	s.WriteString(
		fmt.Sprintf("\n %s\n\n", cmdutil.ColoredOut(fmt.Sprintf("Options (%d)", len(f.product.Options)), color.FgWhite, color.Bold)),
	)

	for _, o := range f.product.Options {
		s.WriteString(fmt.Sprintf("  %s: ", cmdutil.ColoredOut(o.Name, color.FgGreen, color.Bold)))

		var (
			options = make([]string, 0, len(o.OptionValues))
			marker  = cmdutil.ColoredOut("✔", color.FgGreen, color.Bold)
		)
		for _, v := range o.OptionValues {
			tmpl := v.Name
			if v.HasVariants {
				tmpl = fmt.Sprintf("%s%s", marker, v.Name)
			}
			options = append(options, tmpl)
		}
		s.WriteString(fmt.Sprintf("%s\n", strings.Join(options, ", ")))
	}

	return s.String()
}

func (f Formatter) variants() string {
	var (
		s strings.Builder

		maxTitleLen = defaultTitleLen
		maxSkuLen   = padXSmall
	)

	for _, v := range f.product.Variants.Nodes {
		if v.Sku != nil {
			maxSkuLen = max(maxSkuLen, len(*v.Sku))
		}
		maxTitleLen = max(maxTitleLen, len(v.Title))
	}

	variantsCount := len(f.product.Variants.Nodes)
	if f.product.VariantsCount != nil {
		variantsCount = f.product.VariantsCount.Count
	}

	s.WriteString(
		fmt.Sprintf("\n %s\n\n", cmdutil.ColoredOut(fmt.Sprintf("VARIANTS (%d)", variantsCount), color.FgWhite, color.Bold)),
	)
	s.WriteString(
		cmdutil.Gray(
			fmt.Sprintf(
				"  %s  %s %s %s %s %s %s\n",
				cmdutil.Pad("Variant ID", defaultIDLen+spacing),
				cmdutil.Pad("Title", maxTitleLen+spacing),
				cmdutil.Pad("SKU", maxSkuLen+spacing),
				cmdutil.Pad("Price", padMid),
				cmdutil.Pad("Sellable", padMid),
				cmdutil.Pad("Inventory", padMid),
				cmdutil.Pad("Updated At", padMid),
			),
		),
	)

	for _, v := range f.product.Variants.Nodes {
		sku := notAvailable
		if v.Sku != nil {
			sku = *v.Sku
		}
		s.WriteString(
			fmt.Sprintf(
				"  %s  %s %s %s %s %s %s\n",
				cmdutil.ColoredOut(
					cmdutil.Pad(shopctl.ExtractNumericID(v.ID), defaultIDLen), color.FgGreen, color.Bold,
				),
				cmdutil.ShortenAndPad(v.Title, maxTitleLen+spacing),
				cmdutil.ShortenAndPad(sku, maxSkuLen+spacing),
				cmdutil.Pad(v.Price, padMid),
				cmdutil.Pad(fmt.Sprintf("%d", v.SellableOnlineQuantity), padMid),
				cmdutil.Pad(fmt.Sprintf("%d", *v.InventoryQuantity), padMid),
				cmdutil.Pad(v.UpdatedAt, padMid),
			),
		)
	}
	return s.String()
}

func (f Formatter) media() string {
	type media struct {
		id         string
		mimeType   string
		previewURL string
		alt        string
		status     string
	}
	var (
		s strings.Builder

		maxFilenameLen int
		maxAltTextLen  int
	)

	s.WriteString(
		fmt.Sprintf("\n %s\n\n", cmdutil.ColoredOut(fmt.Sprintf("Media (%d)", len(f.product.Media.Nodes)), color.FgWhite, color.Bold)),
	)

	medias := make([]media, 0, len(f.product.Media.Nodes))
	for _, node := range f.product.Media.Nodes {
		var m api.ProductMediaNode
		if n, ok := node.(api.ProductMediaNode); ok {
			m = n
		} else {
			n := node.(map[string]any)
			jsonData, _ := json.Marshal(n)
			_ = json.Unmarshal(jsonData, &m)
		}

		media := media{
			id:         m.ID,
			mimeType:   string(m.MediaContentType),
			previewURL: m.Preview.Image.URL,
			alt:        *m.Preview.Image.AltText,
			status:     string(m.Status),
		}
		medias = append(medias, media)

		maxFilenameLen = max(maxFilenameLen, len(getFileName(m.Preview.Image.URL)))
		maxAltTextLen = max(maxAltTextLen, len(*m.Preview.Image.AltText))
	}
	maxFilenameLen = min(maxFilenameLen, 41) //nolint:mnd
	maxAltTextLen = min(maxAltTextLen, 61)   //nolint:mnd
	maxAltTextLen = max(maxAltTextLen, defaultTitleLen)

	s.WriteString(
		cmdutil.Gray(
			fmt.Sprintf(
				"  %s  %s %s %s %s\n",
				cmdutil.Pad("Media ID", defaultIDLen+spacing),
				cmdutil.Pad("Type", padSmall),
				cmdutil.Pad("Filename", maxFilenameLen+spacing),
				cmdutil.Pad("Alt text", maxAltTextLen+spacing),
				cmdutil.Pad("Status", padSmall),
			),
		),
	)

	for _, m := range medias {
		alt := m.alt
		if alt == "" {
			alt = notAvailable
		}

		s.WriteString(
			fmt.Sprintf(
				"  %s  %s %s %s %s\n",
				cmdutil.ColoredOut(
					cmdutil.Pad(shopctl.ExtractNumericID(m.id), defaultIDLen), color.FgGreen, color.Bold,
				),
				cmdutil.Pad(m.mimeType, padSmall),
				cmdutil.ShortenAndPad(getFileName(m.previewURL), maxFilenameLen+spacing),
				cmdutil.ShortenAndPad(alt, maxAltTextLen+spacing),
				cmdutil.Pad(m.status, padSmall),
			),
		)
	}

	return s.String()
}

func (f Formatter) footer() string {
	url := fmt.Sprintf("https://%s/products/%s", f.store, f.product.Handle)
	if f.product.Status == schema.ProductStatusDraft {
		url = *f.product.OnlineStorePreviewURL
	}
	return cmdutil.Gray(
		fmt.Sprintf("View this product in browser: %s", url),
	)
}

func (f Formatter) separator(msg string) string {
	pad := func(m string) string {
		if m != "" {
			return fmt.Sprintf(" %s ", m)
		}
		return m
	}

	sep := "————————————————————————"
	if msg == "" {
		return cmdutil.Gray(fmt.Sprintf("%s%s", sep, sep))
	}
	return cmdutil.Gray(fmt.Sprintf("%s%s%s", sep, pad(msg), sep))
}

// renderPlain renders in plain view for notty env.
func (f Formatter) renderPlain(wt io.Writer) error {
	r, err := cmdutil.NoTTYRenderer()
	if err != nil {
		return err
	}
	out, err := r.Render(f.String())
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(wt, out)
	return err
}

func getFileName(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return path.Base(parsedURL.Path)
}
