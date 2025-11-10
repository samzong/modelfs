package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// manager is a minimal stub that imitates a controller-runtime manager.
type manager struct {
	startedAt time.Time
}

func newManager() *manager {
	return &manager{startedAt: time.Now().UTC()}
}

func (m *manager) Start(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func run(ctx context.Context) error {
	mgr := newManager()
	fmt.Printf("modelFS manager booted at %s\n", mgr.startedAt.Format(time.RFC3339))
	select {
	case <-time.After(2 * time.Second):
		fmt.Println("manager completed dry-run interval")
		return nil
	case <-ctx.Done():
		if err := ctx.Err(); err != nil && err != context.Canceled {
			return fmt.Errorf("manager stopped: %w", err)
		}
		return nil
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
