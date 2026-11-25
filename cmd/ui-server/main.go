package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	versionpkg "github.com/samzong/modelfs/internal/version"
	"github.com/samzong/modelfs/pkg/ui/kube"
	"github.com/samzong/modelfs/pkg/ui/provider"
	"github.com/samzong/modelfs/pkg/ui/server"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

type serverConfig struct {
	addr        string
	namespace   string
	readHeader  time.Duration
	useMock     bool
	logLevel    string
	showVersion bool
}

func parseServerConfig(fs *flag.FlagSet, args []string) (serverConfig, error) {
	var cfg serverConfig
	fs.StringVar(&cfg.addr, "addr", envOr("UI_SERVER_ADDR", ":8080"), "address to bind (e.g., :8080)")
	fs.StringVar(&cfg.namespace, "namespace", envOr("UI_SERVER_NAMESPACE", "model-system"), "default namespace to scope queries")
	fs.DurationVar(&cfg.readHeader, "read-header-timeout", 15*time.Second, "maximum time to read request headers")
	fs.BoolVar(&cfg.useMock, "mock", envOrBool("UI_SERVER_USE_MOCK", false), "force use mock store instead of Kubernetes")
	fs.StringVar(&cfg.logLevel, "log-level", envOr("UI_SERVER_LOG_LEVEL", "INFO"), "log level: DEBUG, INFO, WARN, ERROR")
	fs.BoolVar(&cfg.showVersion, "version", false, "print version information and exit")
	if err := fs.Parse(args); err != nil {
		return serverConfig{}, err
	}
	return cfg, nil
}

func runServer(
	args []string,
	stdout io.Writer,
	fs *flag.FlagSet,
	readBuildInfo func() (*debug.BuildInfo, bool),
	start func(serverConfig) error,
) error {
	cfg, err := parseServerConfig(fs, args)
	if err != nil {
		return err
	}
	if cfg.showVersion {
		_, err := fmt.Fprintln(stdout, versionpkg.Info{Version: version, Commit: commit, Date: date}.Format(readBuildInfo))
		return err
	}
	return start(cfg)
}

func startServer(cfg serverConfig) error {
	logger := newLogger(cfg.logLevel)
	logger.Info("starting UI gateway", "addr", cfg.addr, "namespace", cfg.namespace, "mock", cfg.useMock, "log_level", cfg.logLevel)

	store := initProvider(logger, cfg.useMock)

	srv := server.New(
		server.Config{
			DefaultNamespace:  cfg.namespace,
			ReadHeaderTimeout: cfg.readHeader,
		},
		server.WithProvider(store),
	)

	httpServer := &http.Server{
		Addr:              cfg.addr,
		Handler:           srv.Routes(),
		ReadHeaderTimeout: cfg.readHeader,
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
		return err
	}
	logger.Info("UI gateway stopped cleanly")
	return nil
}

func main() {
	if err := runServer(os.Args[1:], os.Stdout, flag.CommandLine, nil, startServer); err != nil {
		os.Exit(1)
	}
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
