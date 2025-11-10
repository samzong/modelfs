package controllers

import (
	"context"
	"fmt"

	modelv1 "github.com/samzong/modelfs/api/v1"
)

// ModelSyncReconciler reconciles ModelSync objects.
type ModelSyncReconciler struct{}

// Reconcile prints a stub message.
func (r *ModelSyncReconciler) Reconcile(ctx context.Context, sync modelv1.ModelSync) error {
	_ = ctx
	fmt.Printf("Reconciling model sync %s/%s\n", sync.Metadata.Namespace, sync.Metadata.Name)
	return nil
}

// Setup registers the reconciler with the manager stub.
func (r *ModelSyncReconciler) Setup(_ any) error {
	return nil
}
