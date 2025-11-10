package controllers

import (
	"context"
	"fmt"

	modelv1 "github.com/samzong/modelfs/api/v1"
)

// ModelReconciler performs reconciliation for Model resources.
type ModelReconciler struct{}

// Reconcile runs the reconciliation logic for a Model resource.
func (r *ModelReconciler) Reconcile(ctx context.Context, model modelv1.Model) error {
	_ = ctx
	fmt.Printf("Reconciling model %s/%s\n", model.Metadata.Namespace, model.Metadata.Name)
	return nil
}

// Setup registers the reconciler with the supplied manager interface.
func (r *ModelReconciler) Setup(_ any) error {
	return nil
}
