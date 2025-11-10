package controllers

import (
	"context"
	"fmt"

	modelv1 "github.com/samzong/modelfs/api/v1"
	"github.com/samzong/modelfs/pkg/dataset"
)

// ModelSourceReconciler wires ModelSource definitions into BaizeAI/dataset.
type ModelSourceReconciler struct {
	Dataset  dataset.Client
	Registry *StaticRegistry
}

// Reconcile registers or updates the backing dataset source.
func (r *ModelSourceReconciler) Reconcile(ctx context.Context, source modelv1.ModelSource) error {
	if r.Dataset == nil {
		return fmt.Errorf("dataset client is not configured")
	}
	if err := r.Dataset.EnsureSource(ctx, source); err != nil {
		return fmt.Errorf("ensure dataset source: %w", err)
	}
	if r.Registry != nil {
		r.Registry.SetModelSource(source)
	}
	return nil
}

// Setup currently validates dependencies in lieu of a real manager.
func (r *ModelSourceReconciler) Setup(_ any) error {
	if r.Dataset == nil {
		return fmt.Errorf("dataset client is required")
	}
	if r.Registry == nil {
		r.Registry = NewStaticRegistry()
	}
	return nil
}
