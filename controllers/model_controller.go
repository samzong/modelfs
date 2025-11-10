package controllers

import (
	"context"
	"fmt"

	modelv1 "github.com/samzong/modelfs/api/v1"
)

// ModelReconciler tracks model resources and keeps the shared registry in sync.
type ModelReconciler struct {
	Registry *StaticRegistry
}

// Reconcile records model metadata for downstream sync controllers.
func (r *ModelReconciler) Reconcile(ctx context.Context, model modelv1.Model) error {
	if r.Registry == nil {
		return fmt.Errorf("model registry is not configured")
	}
	_ = ctx
	r.Registry.SetModel(model)
	return nil
}

// Setup currently initialises the registry for offline testing.
func (r *ModelReconciler) Setup(_ any) error {
	if r.Registry == nil {
		r.Registry = NewStaticRegistry()
	}
	return nil
}
