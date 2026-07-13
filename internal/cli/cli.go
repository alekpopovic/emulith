package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/emulith/emulith/internal/config"
	"github.com/emulith/emulith/internal/server"
	"github.com/emulith/emulith/internal/state"
	awsprovider "github.com/emulith/emulith/providers/aws"
	"github.com/emulith/emulith/providers/aws/s3"
	"github.com/emulith/emulith/providers/aws/sqs"
	"github.com/emulith/emulith/providers/aws/sts"
	"github.com/spf13/cobra"
)

func NewCommand(out, errOut io.Writer, version, commit, built string) *cobra.Command {
	root := &cobra.Command{Use: "emulith", SilenceUsage: true, SilenceErrors: true}
	root.SetOut(out)
	root.SetErr(errOut)
	root.AddCommand(newVersionCommand(out, version, commit, built), newServeCommand(errOut, version))
	return root
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
		srv := server.New(cfg.Addr, version, gateway)
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
