package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/samzong/modelfs/pkg/ui/kube"
	"github.com/samzong/modelfs/pkg/ui/provider"
	"github.com/samzong/modelfs/pkg/ui/server"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func main() {
	var (
		addr       string
		namespace  string
		readHeader time.Duration
		useMock    bool
		logLevel   string
	)

	flag.StringVar(&addr, "addr", envOr("UI_SERVER_ADDR", ":8080"), "address to bind (e.g., :8080)")
	flag.StringVar(&namespace, "namespace", envOr("UI_SERVER_NAMESPACE", "model-system"), "default namespace to scope queries")
	flag.DurationVar(&readHeader, "read-header-timeout", 15*time.Second, "maximum time to read request headers")
	flag.BoolVar(&useMock, "mock", envOrBool("UI_SERVER_USE_MOCK", false), "force use mock store instead of Kubernetes")
	flag.StringVar(&logLevel, "log-level", envOr("UI_SERVER_LOG_LEVEL", "INFO"), "log level: DEBUG, INFO, WARN, ERROR")
	flag.Parse()

	logger := newLogger(logLevel)
	logger.Info("starting UI gateway", "addr", addr, "namespace", namespace, "mock", useMock, "log_level", logLevel)

	store := initProvider(logger, useMock)

	srv := server.New(
		server.Config{
			DefaultNamespace:  namespace,
			ReadHeaderTimeout: readHeader,
		},
		server.WithProvider(store),
	)

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           srv.Routes(),
		ReadHeaderTimeout: readHeader,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down UI gateway")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	logger.Info("UI gateway stopped cleanly")
}

func envOr(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func envOrBool(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v == "true" || v == "1" || v == "yes"
	}
	return fallback
}

func newLogger(levelStr string) *slog.Logger {
	level := parseLogLevel(levelStr)
	opts := &slog.HandlerOptions{Level: level}
	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}

func parseLogLevel(levelStr string) slog.Level {
	switch strings.ToUpper(strings.TrimSpace(levelStr)) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func initProvider(logger *slog.Logger, forceMock bool) provider.Store {
	if forceMock {
		logger.Info("using mock provider (forced via --mock flag or UI_SERVER_USE_MOCK)")
		return provider.NewMockStore()
	}
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Warn("falling back to mock provider; kube config not available", "error", err)
		return provider.NewMockStore()
	}
	store, err := kube.NewStore(cfg)
	if err != nil {
		logger.Warn("falling back to mock provider; kube store init failed", "error", err)
		return provider.NewMockStore()
	}
	logger.Info("initialized Kubernetes-backed provider")
	return store
}
