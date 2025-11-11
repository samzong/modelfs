package controllers

import (
	"context"
	"fmt"

	modelv1 "github.com/samzong/modelfs/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ModelReferenceReconciler reconciles a ModelReference object
type ModelReferenceReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Registry ModelRegistry
}

//+kubebuilder:rbac:groups=model.samzong.dev,resources=modelreferences,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=model.samzong.dev,resources=modelreferences/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=model.samzong.dev,resources=modelreferences/finalizers,verbs=update
//+kubebuilder:rbac:groups=model.samzong.dev,resources=models,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop.
func (r *ModelReferenceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ref := &modelv1.ModelReference{}
	if err := r.Get(ctx, req.NamespacedName, ref); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Validate that the referenced model exists
	_, err := r.Registry.GetModel(ctx, ref.Namespace, ref.Spec.ModelName)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("resolve model for reference: %w", err)
	}

	// Update resolved version
	ref.Status.ResolvedVersion = ref.Spec.Version

	if err := r.Status().Update(ctx, ref); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ModelReferenceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&modelv1.ModelReference{}).
		Complete(r)
}
