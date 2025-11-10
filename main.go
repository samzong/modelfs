package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/samzong/modelfs/controllers"
	"github.com/samzong/modelfs/pkg/dataset"
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
	endpoint := os.Getenv("DATASET_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:8080"
	}

	datasetClient, err := dataset.NewHTTPClient(endpoint)
	if err != nil {
		return fmt.Errorf("initialise dataset client: %w", err)
	}

	registry := controllers.NewStaticRegistry()

	modelReconciler := &controllers.ModelReconciler{Registry: registry}
	if err := modelReconciler.Setup(nil); err != nil {
		return fmt.Errorf("setup model reconciler: %w", err)
	}

	modelSourceReconciler := &controllers.ModelSourceReconciler{Dataset: datasetClient, Registry: registry}
	if err := modelSourceReconciler.Setup(nil); err != nil {
		return fmt.Errorf("setup model source reconciler: %w", err)
	}

	modelSyncReconciler := &controllers.ModelSyncReconciler{Dataset: datasetClient, Registry: registry}
	if err := modelSyncReconciler.Setup(nil); err != nil {
		return fmt.Errorf("setup model sync reconciler: %w", err)
	}

	modelReferenceReconciler := &controllers.ModelReferenceReconciler{Registry: registry}
	if err := modelReferenceReconciler.Setup(nil); err != nil {
		return fmt.Errorf("setup model reference reconciler: %w", err)
	}

	mgr := newManager()
	fmt.Printf("modelFS manager booted at %s (dataset endpoint %s)\n", mgr.startedAt.Format(time.RFC3339), endpoint)
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
