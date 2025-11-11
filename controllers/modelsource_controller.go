package controllers

import (
	"context"

	modelv1 "github.com/samzong/modelfs/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ModelSourceReconciler reconciles a ModelSource object
type ModelSourceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=model.samzong.dev,resources=modelsources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=model.samzong.dev,resources=modelsources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=model.samzong.dev,resources=modelsources/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop.
func (r *ModelSourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	source := &modelv1.ModelSource{}
	if err := r.Get(ctx, req.NamespacedName, source); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Validate source configuration
	// TODO: Add proper validation logic

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ModelSourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&modelv1.ModelSource{}).
		Complete(r)
}
