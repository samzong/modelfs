package controllers

import (
	"context"
	"fmt"

	modelv1 "github.com/samzong/modelfs/api/v1"
)

// ModelReferenceReconciler validates that references resolve to known models.
type ModelReferenceReconciler struct {
	Registry ModelRegistry
}

// Reconcile ensures the referenced model exists in the shared registry.
func (r *ModelReferenceReconciler) Reconcile(ctx context.Context, ref modelv1.ModelReference) error {
	if r.Registry == nil {
		return fmt.Errorf("model registry is not configured")
	}
	_, err := r.Registry.GetModel(ctx, ref.Metadata.Namespace, ref.Spec.ModelName)
	if err != nil {
		return fmt.Errorf("resolve model for reference: %w", err)
	}
	return nil
}

// Setup ensures the reconciler has the required dependencies.
func (r *ModelReferenceReconciler) Setup(_ any) error {
	if r.Registry == nil {
		r.Registry = NewStaticRegistry()
	}
	return nil
}
