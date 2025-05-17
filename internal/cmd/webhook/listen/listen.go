package listen

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/shopctl"
	"github.com/ankitpokhrel/shopctl/internal/api"
	"github.com/ankitpokhrel/shopctl/internal/cmdutil"
	"github.com/ankitpokhrel/shopctl/schema"
)

const (
	helpText = `Listen to the webhook events as per the user config.`
	examples = `# Run a js script to enrich products on creation
$ shopctl webhook listen --topic PRODUCTS_CREATE --exec "node enrich.js" --url https://example.com/products/create

# Run a python script to sync changes to marketplaces on product update
$ shopctl webhook listen --topic PRODUCTS_UPDATE --exec "python sync.py" --url https://example.com/products/update --port 8080

# Execute a curl directly
$ shopctl webhook listen --topic PRODUCTS_CREATE --exec "curl -X POST http://httpbin.org/post -H 'Content-Type:application/json' -d @-" --url https://example.com/products/create

# Listen to webhook by its ID
$ shopctl webhook listen --id 1434973307104 --exec "./process.sh"
`
)

type flag struct {
	id    string
	topic string
	exec  string
	url   string
	port  uint
}

func (f *flag) parse(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		f.id = shopctl.ShopifyWebhookSubscriptionID(args[0])
	}

	topic, err := cmd.Flags().GetString("topic")
	cmdutil.ExitOnErr(err)

	handler, err := cmd.Flags().GetString("exec")
	cmdutil.ExitOnErr(err)

	url, err := cmd.Flags().GetString("url")
	cmdutil.ExitOnErr(err)

	port, err := cmd.Flags().GetUint("port")
	cmdutil.ExitOnErr(err)

	if f.id == "" && (topic == "" || url == "") {
		cmdutil.ExitOnErr(
			cmdutil.HelpErrorf("Either webhook subscription id or topic and url is required", examples),
		)
	}
	if handler == "" {
		cmdutil.ExitOnErr(
			cmdutil.HelpErrorf("Flag '--exec' is required", examples),
		)
	}

	f.topic = topic
	f.exec = handler
	f.url = url
	f.port = port
}

// NewCmdListen configures an event listener.
func NewCmdListen() *cobra.Command {
	cmd := cobra.Command{
		Use:     "listen",
		Short:   "Listen to the configured webhook events",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"lsn"},
		RunE: func(cmd *cobra.Command, args []string) error {
			client := cmd.Context().Value(cmdutil.KeyGQLClient).(*api.GQLClient)

			cmdutil.ExitOnErr(run(cmd, args, client))
			return nil
		},
	}

	cmd.Flags().String("id", "", "Registered webhook id")
	cmd.Flags().StringP("topic", "t", "", "Webhook topic to listen to")
	cmd.Flags().StringP("exec", "e", "", "Handler to execute")
	cmd.Flags().String("url", "", "Endpoint for the webhook registration")
	cmd.Flags().Uint("port", 4726, "Port to use for local webhook server") //nolint:mnd

	cmd.Flags().SortFlags = false

	return &cmd
}

func run(cmd *cobra.Command, args []string, client *api.GQLClient) error {
	flag := &flag{}
	flag.parse(cmd, args)

	var (
		sub *schema.WebhookSubscription
		err error
	)

	if flag.id != "" {
		sub, err = client.GetWebhookByID(flag.id)
		if err != nil {
			return err
		}
		endpoint := sub.Endpoint.(map[string]any)
		cbURL := endpoint["callbackUrl"].(string)
		fmt.Printf("Webhook ID %q is registerd for topic %q with endpoint %q\n", sub.ID, sub.Topic, cbURL)
	} else {
		// Register webhook to Shopify.
		sub, err = client.SubscribeWebhook(flag.topic, flag.url)
		if err != nil && !errors.Is(err, api.ErrAddrTaken) {
			return err
		}
		if err != nil && errors.Is(err, api.ErrAddrTaken) {
			fmt.Printf("Webhook for topic %q exists with endpoint %q\n", flag.topic, flag.url)
		} else {
			fmt.Printf("Webhook registered for topic %q with endpoint %q on api version %q\n", sub.Topic, flag.url, sub.ApiVersion.Handle)
		}
	}

	// Listen to the event.
	listen(sub.Topic, flag.port, func(payload map[string]any) error {
		data, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		return handle(flag.exec, data)
	})

	q := make(chan os.Signal, 1)
	signal.Notify(q, syscall.SIGINT, syscall.SIGTERM)
	<-q

	return nil
}

func listen(topic schema.WebhookSubscriptionTopic, port uint, handler func(map[string]any) error) {
	whTopic := strings.ToLower(strings.ReplaceAll(string(topic), "_", "/"))
	http.HandleFunc("/"+whTopic, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		receivedTopic := r.Header.Get("X-Shopify-Topic")
		if receivedTopic != whTopic {
			http.Error(w, "Topic does not match", http.StatusForbidden)
			return
		}

		go func() {
			if err := handler(payload); err != nil {
				fmt.Printf("Handler error: %v\n", err)
			}
		}()

		w.WriteHeader(http.StatusOK)
	})

	go func() {
		addr := fmt.Sprintf(":%d", port)
		fmt.Printf("Listening for events on %s\n", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			fmt.Printf("Error starting webhook listener: %v\n", err)
		}
	}()
}

func handle(h string, payload []byte) error {
	var cmd *exec.Cmd

	parts := strings.Fields(h)
	if len(parts) == 0 {
		return fmt.Errorf("invalid handler: %s", h)
	}
	script := parts[0]
	args := parts[1:]

	cmd = exec.Command(script, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Pass payload through stdin.
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin: %w", err)
	}

	//nolint:errcheck
	go func() {
		defer stdin.Close()
		stdin.Write(payload)
	}()

	return cmd.Run()
}
