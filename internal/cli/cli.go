package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/alekpopovic/emulith/internal/config"
	"github.com/alekpopovic/emulith/internal/manifest"
	"github.com/alekpopovic/emulith/internal/server"
	"github.com/alekpopovic/emulith/internal/state"
	awsprovider "github.com/alekpopovic/emulith/providers/aws"
	"github.com/alekpopovic/emulith/providers/aws/s3"
	"github.com/alekpopovic/emulith/providers/aws/sqs"
	"github.com/alekpopovic/emulith/providers/aws/sts"
	"github.com/spf13/cobra"
)

func NewCommand(out, errOut io.Writer, version, commit, built string) *cobra.Command {
	return NewCommandWithClient(out, errOut, version, commit, built, &http.Client{Timeout: 15 * time.Second})
}

func NewCommandWithClient(out, errOut io.Writer, version, commit, built string, client *http.Client) *cobra.Command {
	root := &cobra.Command{Use: "emulith", SilenceUsage: true, SilenceErrors: true}
	root.SetOut(out)
	root.SetErr(errOut)
	root.AddCommand(newVersionCommand(out, version, commit, built), newServeCommand(errOut, version), newResetCommand(out, client), newApplyCommand(out, client))
	return root
}

func newApplyCommand(out io.Writer, client *http.Client) *cobra.Command {
	endpoint := os.Getenv("EMULITH_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:4566"
	}
	var file string
	cmd := &cobra.Command{Use: "apply", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		if file == "" {
			return errors.New("--file is required")
		}
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read manifest: %w", err)
		}
		m, err := manifest.Parse(data)
		if err != nil {
			return err
		}
		applier, err := manifest.NewApplier(endpoint, client)
		if err != nil {
			return err
		}
		return applier.Apply(cmd.Context(), m, out)
	}}
	cmd.Flags().StringVarP(&file, "file", "f", "", "manifest file")
	cmd.Flags().StringVar(&endpoint, "endpoint", endpoint, "Emulith base endpoint")
	return cmd
}

func newResetCommand(out io.Writer, client *http.Client) *cobra.Command {
	endpoint := os.Getenv("EMULITH_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:4566"
	}
	cmd := &cobra.Command{Use: "reset", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		base, err := url.Parse(endpoint)
		if err != nil || base.Scheme == "" || base.Host == "" {
			return fmt.Errorf("invalid endpoint %q", endpoint)
		}
		base.Path = strings.TrimRight(base.Path, "/") + "/_emulith/reset"
		request, err := http.NewRequestWithContext(cmd.Context(), http.MethodPost, base.String(), nil)
		if err != nil {
			return err
		}
		response, err := client.Do(request)
		if err != nil {
			return fmt.Errorf("reset request failed: %w", err)
		}
		defer response.Body.Close()
		var result struct {
			Status string `json:"status"`
			Reset  bool   `json:"reset"`
		}
		if err := json.NewDecoder(io.LimitReader(response.Body, 64<<10)).Decode(&result); err != nil {
			return fmt.Errorf("invalid reset response: %w", err)
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 || result.Status != "ok" || !result.Reset {
			return fmt.Errorf("reset failed with HTTP %d", response.StatusCode)
		}
		_, err = fmt.Fprintln(out, "Emulith state reset successfully.")
		return err
	}}
	cmd.Flags().StringVar(&endpoint, "endpoint", endpoint, "Emulith base endpoint")
	return cmd
}

func Execute(out, errOut io.Writer, version, commit, built string) error {
	return NewCommand(out, errOut, version, commit, built).Execute()
}

func newVersionCommand(out io.Writer, version, commit, built string) *cobra.Command {
	return &cobra.Command{Use: "version", Args: cobra.NoArgs, RunE: func(*cobra.Command, []string) error {
		_, err := fmt.Fprintf(out, "emulith %s\ncommit: %s\nbuilt: %s\n", version, commit, built)
		return err
	}}
}

func newServeCommand(errOut io.Writer, version string) *cobra.Command {
	cfg := config.FromEnvironment()
	cmd := &cobra.Command{Use: "serve", Args: cobra.NoArgs, RunE: func(*cobra.Command, []string) error {
		logger := slog.New(slog.NewJSONHandler(errOut, nil))
		store, err := state.Open(context.Background(), cfg.DataDir)
		if err != nil {
			return fmt.Errorf("open state: %w", err)
		}
		defer store.Close()
		gateway := awsprovider.NewGateway(store, logger)
		gateway.SetSTS(sts.New())
		gateway.SetS3(s3.New(store))
		gateway.SetSQS(sqs.New(store))
		srv := server.NewWithState(cfg.Addr, version, store, logger, gateway)
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()
		go func() {
			<-ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := srv.HTTPServer().Shutdown(shutdownCtx); err != nil {
				logger.Error("shutdown failed", "error", err)
			}
		}()
		logger.Info("server starting", "addr", cfg.Addr, "data_dir", cfg.DataDir)
		err = srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}}
	cmd.Flags().StringVar(&cfg.Addr, "addr", cfg.Addr, "listen address")
	cmd.Flags().StringVar(&cfg.DataDir, "data-dir", cfg.DataDir, "state data directory")
	return cmd
}
