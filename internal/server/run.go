package server

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bShaak/habitui/internal/api"
)

// Run parses serve flags, opens the database and blocks serving HTTP until
// SIGINT/SIGTERM.
func Run(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var (
		addr   string
		dbPath string
	)
	fs.StringVar(&addr, "addr", "127.0.0.1:8080", "listen address")
	fs.StringVar(&dbPath, "db", "", "SQLite database path (default: ~/.habitui/habit.db or HABITUI_DB)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	svc, err := api.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer svc.Close()

	srv := &http.Server{
		Addr:              addr,
		Handler:           New(svc),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ln) }()

	fmt.Fprintf(os.Stderr, "habitui API listening on http://%s\n", ln.Addr())
	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case sig := <-stop:
		fmt.Fprintf(os.Stderr, "received %v, shutting down\n", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
	}
	return nil
}
