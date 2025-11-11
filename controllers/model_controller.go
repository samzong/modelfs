package controllers

import (
	"context"
	"fmt"

	modelv1 "github.com/samzong/modelfs/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

// ModelReconciler reconciles a Model object
type ModelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=model.samzong.dev,resources=models,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=model.samzong.dev,resources=models/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=model.samzong.dev,resources=models/finalizers,verbs=update
//+kubebuilder:rbac:groups=model.samzong.dev,resources=modelsyncs,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ModelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	model := &modelv1.Model{}
	if err := r.Get(ctx, req.NamespacedName, model); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Find all ModelSync instances that reference this Model
	syncedVersions, err := r.findSyncedVersions(ctx, model)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("find synced versions: %w", err)
	}

	// Update Model status with synced versions
	status := model.Status.DeepCopy()
	status.SyncedVersions = syncedVersions

	// Update conditions based on synced versions
	now := metav1.Now()
	if len(syncedVersions) == 0 {
		// No versions being synced
		r.updateCondition(status, metav1.Condition{
			Type:               "VersionsAvailable",
			Status:             metav1.ConditionFalse,
			Reason:             "NoSyncTasks",
			Message:            "No ModelSync instances found for this model",
			LastTransitionTime: now,
			ObservedGeneration: model.Generation,
		})
	} else {
		// Has versions being synced
		r.updateCondition(status, metav1.Condition{
			Type:               "VersionsAvailable",
			Status:             metav1.ConditionTrue,
			Reason:             "SyncTasksFound",
			Message:            fmt.Sprintf("Found %d version(s) being synced", len(syncedVersions)),
			LastTransitionTime: now,
			ObservedGeneration: model.Generation,
		})
	}

	model.Status = *status
	if err := r.Status().Update(ctx, model); err != nil {
		return ctrl.Result{}, fmt.Errorf("update status: %w", err)
	}

	return ctrl.Result{}, nil
}

// findSyncedVersions finds all ModelSync instances that reference this Model and collects version information.
func (r *ModelReconciler) findSyncedVersions(ctx context.Context, model *modelv1.Model) ([]modelv1.SyncedVersion, error) {
	// List all ModelSync instances in the same namespace
	syncList := &modelv1.ModelSyncList{}
	if err := r.List(ctx, syncList, client.InNamespace(model.Namespace)); err != nil {
		return nil, fmt.Errorf("list modelsyncs: %w", err)
	}

	var syncedVersions []modelv1.SyncedVersion
	for _, sync := range syncList.Items {
		// Check if this ModelSync references our Model
		if sync.Spec.ModelRef != model.Name {
			continue
		}

		// Determine the version being synced (from ModelSync)
		version := sync.Spec.Version

		// Check if this version is ready
		ready := false
		for _, cond := range sync.Status.Conditions {
			if cond.Type == "SyncReady" && cond.Status == metav1.ConditionTrue {
				ready = true
				break
			}
		}

		syncedVersions = append(syncedVersions, modelv1.SyncedVersion{
			Version:       version,
			ModelSyncName: sync.Name,
			LastSyncedAt:  sync.Status.LastSyncedAt,
			Ready:         ready,
		})
	}

	return syncedVersions, nil
}

// updateCondition updates or adds a condition to the status.
func (r *ModelReconciler) updateCondition(status *modelv1.ModelStatus, condition metav1.Condition) {
	if status.Conditions == nil {
		status.Conditions = []metav1.Condition{}
	}

	now := metav1.Now()
	condition.LastTransitionTime = now

	for i, existing := range status.Conditions {
		if existing.Type == condition.Type {
			// Check if condition actually changed
			if existing.Status == condition.Status &&
				existing.Reason == condition.Reason &&
				existing.Message == condition.Message {
				return // No change
			}
			// Preserve LastTransitionTime if status didn't change
			if existing.Status == condition.Status {
				condition.LastTransitionTime = existing.LastTransitionTime
			}
			condition.ObservedGeneration = existing.ObservedGeneration
			status.Conditions[i] = condition
			return
		}
	}

	// Condition doesn't exist, add it
	status.Conditions = append(status.Conditions, condition)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ModelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&modelv1.Model{}).
		Watches(
			&modelv1.ModelSync{},
			handler.EnqueueRequestsFromMapFunc(r.mapModelSyncToModel),
		).
		Complete(r)
}

// mapModelSyncToModel maps a ModelSync to Model reconciliation requests.
func (r *ModelReconciler) mapModelSyncToModel(ctx context.Context, obj client.Object) []ctrl.Request {
	sync, ok := obj.(*modelv1.ModelSync)
	if !ok {
		return []ctrl.Request{}
	}

	// Enqueue the Model that this ModelSync references
	return []ctrl.Request{
		{
			NamespacedName: client.ObjectKey{
				Name:      sync.Spec.ModelRef,
				Namespace: sync.Namespace,
			},
		},
	}
}
