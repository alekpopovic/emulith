package cli

import (
	"fmt"
	"github.com/alekpopovic/emulith/internal/config"
	"github.com/spf13/cobra"
	"io"
	"net"
	"strings"
)

func newGCPCommand(out io.Writer) *cobra.Command {
	c := config.FromEnvironment()
	var host, project string
	var pp, fp, sp int
	root := &cobra.Command{Use: "gcp"}
	e := &cobra.Command{Use: "env", Args: cobra.NoArgs, RunE: func(*cobra.Command, []string) error {
		h := host
		if h == "" {
			h = "127.0.0.1"
		}
		if net.ParseIP(h) == nil && !strings.EqualFold(h, "localhost") {
			return fmt.Errorf("host must be loopback")
		}
		if strings.Contains(h, ":") {
			h = "[" + h + "]"
		}
		fmt.Fprintf(out, "PUBSUB_EMULATOR_HOST=%s:%d\nPUBSUB_PROJECT_ID=%s\nFIRESTORE_EMULATOR_HOST=%s:%d\nFIRESTORE_PROJECT_ID=%s\nSTORAGE_EMULATOR_HOST=http://%s:%d\nGOOGLE_CLOUD_PROJECT=%s\n", h, pp, project, h, fp, project, h, sp, project)
		return nil
	}}
	e.Flags().StringVar(&host, "host", "127.0.0.1", "local host")
	e.Flags().IntVar(&pp, "pubsub-port", 8085, "port")
	e.Flags().IntVar(&fp, "firestore-port", 8080, "port")
	e.Flags().IntVar(&sp, "storage-port", 9023, "port")
	e.Flags().StringVar(&project, "project", c.GCPProject, "project ID")
	root.AddCommand(e)
	return root
}
