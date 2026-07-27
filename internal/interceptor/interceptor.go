package interceptor

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

// Blocks the work flow, use with "go"
func SetInterceptorOn(f func()) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	sig := <-sigCh
	slog.Info("Signal received, shutting down", "signal", sig)
	f()
}
