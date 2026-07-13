package cli

import (
	"fmt"
	"github.com/alekpopovic/emulith/providers/azure"
	"github.com/spf13/cobra"
	"io"
	"net"
	"strings"
)

func newAzureCommand(out io.Writer) *cobra.Command {
	root := &cobra.Command{Use: "azure"}
	var host, account, key string
	var bp, qp, tp int
	cmd := &cobra.Command{Use: "connection-string", Args: cobra.NoArgs, RunE: func(*cobra.Command, []string) error {
		if host == "" {
			host = "127.0.0.1"
		}
		if account == "" {
			account = azure.DefaultAccountName
		}
		if key == "" {
			key = azure.DefaultAccountKey
		}
		if net.ParseIP(host) == nil && !strings.HasSuffix(host, "localhost") {
			return fmt.Errorf("host must be local")
		}
		h := host
		if strings.Contains(h, ":") {
			h = "[" + h + "]"
		}
		_, e := fmt.Fprintf(out, "DefaultEndpointsProtocol=http;AccountName=%s;AccountKey=%s;BlobEndpoint=http://%s:%d/%s;QueueEndpoint=http://%s:%d/%s;TableEndpoint=http://%s:%d/%s;\n", account, key, h, bp, account, h, qp, account, h, tp, account)
		return e
	}}
	cmd.Flags().StringVar(&host, "host", "127.0.0.1", "local host")
	cmd.Flags().IntVar(&bp, "blob-port", 10000, "blob port")
	cmd.Flags().IntVar(&qp, "queue-port", 10001, "queue port")
	cmd.Flags().IntVar(&tp, "table-port", 10002, "table port")
	cmd.Flags().StringVar(&account, "account-name", "", "account name")
	cmd.Flags().StringVar(&key, "account-key", "", "account key")
	root.AddCommand(cmd)
	return root
}
