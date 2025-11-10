package controllers

import (
	"context"
	"fmt"

	modelv1 "github.com/samzong/modelfs/api/v1"
)

// ModelReferenceReconciler reconciles ModelReference resources.
type ModelReferenceReconciler struct{}

// Reconcile prints a stub message.
func (r *ModelReferenceReconciler) Reconcile(ctx context.Context, ref modelv1.ModelReference) error {
	_ = ctx
	fmt.Printf("Reconciling model reference %s/%s\n", ref.Metadata.Namespace, ref.Metadata.Name)
	return nil
}

// Setup registers the reconciler with the manager stub.
func (r *ModelReferenceReconciler) Setup(_ any) error {
	return nil
}
