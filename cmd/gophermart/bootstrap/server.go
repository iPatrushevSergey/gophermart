package bootstrap

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gophermart/internal/gophermart/application/port"
)

const serverShutdownTimeout = 5 * time.Second

// StartServer starts the HTTP server in a goroutine.
func StartServer(server *http.Server, log port.Logger) {
	go func() {
		log.Info("gophermart listening", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()
}

// WaitForShutdown waits for SIGINT/SIGTERM and performs graceful shutdown.
func WaitForShutdown(server *http.Server, log port.Logger) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutdown signal received, stopping server...")
	ctx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("server shutdown failed", "error", err)
		return err
	}
	log.Info("server stopped gracefully")
	return nil
}
