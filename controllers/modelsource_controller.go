package controllers

import (
	"context"
	"fmt"

	modelv1 "github.com/samzong/modelfs/api/v1"
)

// ModelSourceReconciler reconciles ModelSource objects.
type ModelSourceReconciler struct{}

// Reconcile prints a log entry for reconciliation attempts.
func (r *ModelSourceReconciler) Reconcile(ctx context.Context, source modelv1.ModelSource) error {
	_ = ctx
	fmt.Printf("Reconciling model source %s/%s\n", source.Metadata.Namespace, source.Metadata.Name)
	return nil
}

// Setup registers the reconciler with the manager stub.
func (r *ModelSourceReconciler) Setup(_ any) error {
	return nil
}
