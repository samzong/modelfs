package controllers

import (
	"context"
	"fmt"
	"time"

	datasetv1alpha1 "github.com/BaizeAI/dataset/api/dataset/v1alpha1"
	modelv1 "github.com/samzong/modelfs/api/v1"
	"github.com/samzong/modelfs/pkg/dataset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ModelSyncReconciler reconciles a ModelSync object
type ModelSyncReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Dataset  dataset.Client
	Registry ModelRegistry
}

//+kubebuilder:rbac:groups=model.samzong.dev,resources=modelsyncs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=model.samzong.dev,resources=modelsyncs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=model.samzong.dev,resources=modelsyncs/finalizers,verbs=update
//+kubebuilder:rbac:groups=model.samzong.dev,resources=models,verbs=get;list;watch
//+kubebuilder:rbac:groups=model.samzong.dev,resources=modelsources,verbs=get;list;watch
//+kubebuilder:rbac:groups=dataset.baizeai.io,resources=datasets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop.
func (r *ModelSyncReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	sync := &modelv1.ModelSync{}
	if err := r.Get(ctx, req.NamespacedName, sync); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Resolve Model reference
	model, err := r.Registry.GetModel(ctx, sync.Namespace, sync.Spec.ModelRef)
	if err != nil {
		return r.updateStatusWithError(ctx, sync, "ModelNotFound", fmt.Errorf("resolve model: %w", err))
	}

	// Get ModelSource from Model
	source, err := r.Registry.GetModelSource(ctx, sync.Namespace, model.Spec.SourceRef)
	if err != nil {
		return r.updateStatusWithError(ctx, sync, "ModelSourceNotFound", fmt.Errorf("resolve model source: %w", err))
	}

	// Validate that the version specified in ModelSync exists in Model.versionConfigs
	if model.Spec.VersionConfigs == nil {
		return r.updateStatusWithError(ctx, sync, "InvalidModel", fmt.Errorf("model %s has no versionConfigs", model.Name))
	}
	if _, exists := model.Spec.VersionConfigs[sync.Spec.Version]; !exists {
		versions := make([]string, 0, len(model.Spec.VersionConfigs))
		for v := range model.Spec.VersionConfigs {
			versions = append(versions, v)
		}
		return r.updateStatusWithError(ctx, sync, "InvalidVersion", fmt.Errorf("version %s not found in Model %s. Available versions: %v", sync.Spec.Version, model.Name, versions))
	}

	// Create or update Dataset CR
	if err := r.Dataset.EnsureDataset(ctx, *model, *source, *sync); err != nil {
		return r.updateStatusWithError(ctx, sync, "DatasetCreationFailed", fmt.Errorf("ensure dataset: %w", err))
	}

	// Get the Dataset CR and sync its status
	datasetName := r.Dataset.GetDatasetName(*model, *sync)
	dataset := &datasetv1alpha1.Dataset{}
	if err := r.Get(ctx, types.NamespacedName{Name: datasetName, Namespace: sync.Namespace}, dataset); err != nil {
		if errors.IsNotFound(err) {
			// Dataset not found yet, wait for it to be created
			return r.updateStatusWithCondition(ctx, sync, "DatasetPending", metav1.ConditionFalse, "DatasetNotFound", "Waiting for Dataset to be created")
		}
		return r.updateStatusWithError(ctx, sync, "DatasetFetchFailed", fmt.Errorf("get dataset: %w", err))
	}

	// Sync Dataset status to ModelSync status
	return r.syncDatasetStatus(ctx, sync, dataset)
}

// syncDatasetStatus synchronizes the Dataset status to ModelSync status.
func (r *ModelSyncReconciler) syncDatasetStatus(ctx context.Context, sync *modelv1.ModelSync, dataset *datasetv1alpha1.Dataset) (ctrl.Result, error) {
	status := sync.Status.DeepCopy()
	needsUpdate := false

	// Update LastSyncedAt if Dataset has LastSyncTime
	if !dataset.Status.LastSyncTime.IsZero() {
		if status.LastSyncedAt.IsZero() || !status.LastSyncedAt.Time.Equal(dataset.Status.LastSyncTime.Time) {
			status.LastSyncedAt = dataset.Status.LastSyncTime
			needsUpdate = true
		}
	}

	// Sync Conditions based on Dataset Phase
	phaseCondition := r.getConditionForPhase(dataset.Status.Phase)
	if r.updateCondition(status, phaseCondition) {
		needsUpdate = true
	}

	// Add Dataset-specific conditions
	if len(dataset.Status.Conditions) > 0 {
		// Sync important conditions from Dataset
		for _, dsCond := range dataset.Status.Conditions {
			if dsCond.Type == "JobStatus" || dsCond.Type == "PVC" || dsCond.Type == "Config" {
				cond := metav1.Condition{
					Type:               fmt.Sprintf("Dataset%s", dsCond.Type),
					Status:             dsCond.Status,
					Reason:             dsCond.Reason,
					Message:            dsCond.Message,
					LastTransitionTime: dsCond.LastTransitionTime,
					ObservedGeneration: sync.Generation,
				}
				if r.updateCondition(status, cond) {
					needsUpdate = true
				}
			}
		}
	}

	if needsUpdate {
		sync.Status = *status
		if err := r.Status().Update(ctx, sync); err != nil {
			return ctrl.Result{}, fmt.Errorf("update status: %w", err)
		}
	}

	// Determine if we need to requeue based on Dataset phase
	switch dataset.Status.Phase {
	case datasetv1alpha1.DatasetStatusPhaseProcessing:
		// Requeue after 30 seconds to check progress
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	case datasetv1alpha1.DatasetStatusPhasePending:
		// Requeue after 10 seconds
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	case datasetv1alpha1.DatasetStatusPhaseFailed:
		// Failed, but still requeue after 5 minutes to retry
		return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
	case datasetv1alpha1.DatasetStatusPhaseReady:
		// Ready, no need to requeue unless there's a schedule
		if sync.Spec.Schedule != "" {
			// If there's a schedule, we'll rely on CronJob or external scheduler
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, nil
	default:
		return ctrl.Result{}, nil
	}
}

// getConditionForPhase returns a Condition based on Dataset phase.
func (r *ModelSyncReconciler) getConditionForPhase(phase datasetv1alpha1.DatasetStatusPhase) metav1.Condition {
	now := metav1.Now()
	switch phase {
	case datasetv1alpha1.DatasetStatusPhaseReady:
		return metav1.Condition{
			Type:               "SyncReady",
			Status:             metav1.ConditionTrue,
			Reason:             "DatasetReady",
			Message:            "Dataset synchronization completed successfully",
			LastTransitionTime: now,
		}
	case datasetv1alpha1.DatasetStatusPhaseProcessing:
		return metav1.Condition{
			Type:               "SyncReady",
			Status:             metav1.ConditionFalse,
			Reason:             "DatasetProcessing",
			Message:            "Dataset synchronization in progress",
			LastTransitionTime: now,
		}
	case datasetv1alpha1.DatasetStatusPhaseFailed:
		return metav1.Condition{
			Type:               "SyncReady",
			Status:             metav1.ConditionFalse,
			Reason:             "DatasetFailed",
			Message:            "Dataset synchronization failed",
			LastTransitionTime: now,
		}
	default: // PENDING
		return metav1.Condition{
			Type:               "SyncReady",
			Status:             metav1.ConditionFalse,
			Reason:             "DatasetPending",
			Message:            "Dataset synchronization pending",
			LastTransitionTime: now,
		}
	}
}

// updateCondition updates or adds a condition to the status.
func (r *ModelSyncReconciler) updateCondition(status *modelv1.ModelSyncStatus, condition metav1.Condition) bool {
	if status.Conditions == nil {
		status.Conditions = []metav1.Condition{}
	}

	for i, existing := range status.Conditions {
		if existing.Type == condition.Type {
			// Check if condition actually changed
			if existing.Status == condition.Status &&
				existing.Reason == condition.Reason &&
				existing.Message == condition.Message {
				return false // No change
			}
			status.Conditions[i] = condition
			return true
		}
	}

	// Condition doesn't exist, add it
	status.Conditions = append(status.Conditions, condition)
	return true
}

// updateStatusWithError updates the status with an error condition.
func (r *ModelSyncReconciler) updateStatusWithError(ctx context.Context, sync *modelv1.ModelSync, reason string, err error) (ctrl.Result, error) {
	status := sync.Status.DeepCopy()
	cond := metav1.Condition{
		Type:               "SyncReady",
		Status:             metav1.ConditionFalse,
		Reason:             reason,
		Message:            err.Error(),
		LastTransitionTime: metav1.Now(),
		ObservedGeneration: sync.Generation,
	}
	r.updateCondition(status, cond)
	sync.Status = *status
	if updateErr := r.Status().Update(ctx, sync); updateErr != nil {
		return ctrl.Result{}, fmt.Errorf("update status: %w (original error: %v)", updateErr, err)
	}
	// Requeue after 1 minute to retry
	return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
}

// updateStatusWithCondition updates the status with a specific condition.
func (r *ModelSyncReconciler) updateStatusWithCondition(ctx context.Context, sync *modelv1.ModelSync, reason string, status metav1.ConditionStatus, messageType, message string) (ctrl.Result, error) {
	statusObj := sync.Status.DeepCopy()
	cond := metav1.Condition{
		Type:               messageType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
		ObservedGeneration: sync.Generation,
	}
	r.updateCondition(statusObj, cond)
	sync.Status = *statusObj
	if err := r.Status().Update(ctx, sync); err != nil {
		return ctrl.Result{}, fmt.Errorf("update status: %w", err)
	}
	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ModelSyncReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&modelv1.ModelSync{}).
		Watches(
			&datasetv1alpha1.Dataset{},
			handler.EnqueueRequestsFromMapFunc(r.mapDatasetToModelSync),
		).
		Complete(r)
}

// mapDatasetToModelSync maps a Dataset to ModelSync reconciliation requests.
func (r *ModelSyncReconciler) mapDatasetToModelSync(ctx context.Context, obj client.Object) []reconcile.Request {
	dataset, ok := obj.(*datasetv1alpha1.Dataset)
	if !ok {
		return []reconcile.Request{}
	}

	// Find ModelSync that owns this Dataset
	syncList := &modelv1.ModelSyncList{}
	if err := r.List(ctx, syncList, client.InNamespace(dataset.Namespace)); err != nil {
		return []reconcile.Request{}
	}

	var requests []reconcile.Request
	for _, sync := range syncList.Items {
		// Check if this Dataset is owned by the ModelSync
		for _, ownerRef := range dataset.OwnerReferences {
			if ownerRef.Kind == "ModelSync" && ownerRef.Name == sync.Name {
				requests = append(requests, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      sync.Name,
						Namespace: sync.Namespace,
					},
				})
				break
			}
		}
	}

	return requests
}
