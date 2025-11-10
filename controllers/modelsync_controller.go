package controllers

import (
	"context"
	"fmt"

	modelv1 "github.com/samzong/modelfs/api/v1"
	"github.com/samzong/modelfs/pkg/dataset"
)

// ModelSyncReconciler bridges ModelSync resources with the dataset sync pipeline.
type ModelSyncReconciler struct {
	Dataset  dataset.Client
	Registry ModelRegistry
}

// Reconcile resolves dependencies and triggers a dataset synchronization.
func (r *ModelSyncReconciler) Reconcile(ctx context.Context, sync modelv1.ModelSync) error {
	if r.Dataset == nil {
		return fmt.Errorf("dataset client is not configured")
	}
	if r.Registry == nil {
		return fmt.Errorf("model registry is not configured")
	}
	model, err := r.Registry.GetModel(ctx, sync.Metadata.Namespace, sync.Spec.ModelRef)
	if err != nil {
		return fmt.Errorf("resolve model: %w", err)
	}
	source, err := r.Registry.GetModelSource(ctx, sync.Metadata.Namespace, sync.Spec.SourceRef)
	if err != nil {
		return fmt.Errorf("resolve model source: %w", err)
	}
	if err := r.Dataset.EnsureSource(ctx, source); err != nil {
		return fmt.Errorf("ensure dataset source: %w", err)
	}
	input := dataset.SyncInput{
		Namespace:      sync.Metadata.Namespace,
		ModelName:      model.Metadata.Name,
		ModelVersion:   model.Spec.Version,
		SourceName:     source.Metadata.Name,
		SourceType:     source.Spec.Type,
		SourceConfig:   source.Spec.Config,
		Schedule:       sync.Spec.Schedule,
		RetentionCount: sync.Spec.RetentionCount,
	}
	if err := r.Dataset.TriggerSync(ctx, input); err != nil {
		return fmt.Errorf("trigger dataset sync: %w", err)
	}
	return nil
}

// Setup currently validates dependencies in lieu of a controller-runtime manager.
func (r *ModelSyncReconciler) Setup(_ any) error {
	if r.Dataset == nil {
		return fmt.Errorf("dataset client is required")
	}
	if r.Registry == nil {
		r.Registry = NewStaticRegistry()
	}
	return nil
}
