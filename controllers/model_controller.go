package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	datasetv1alpha1 "github.com/BaizeAI/dataset/api/dataset/v1alpha1"
	modelv1 "github.com/samzong/modelfs/api/v1"
	"github.com/samzong/modelfs/pkg/dataset"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	ModelFinalizer = "modelfs.samzong.dev/model-finalizer"
)

// ModelReconciler reconciles a Model object
type ModelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=model.samzong.dev,resources=models,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=model.samzong.dev,resources=models/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=model.samzong.dev,resources=models/finalizers,verbs=update
//+kubebuilder:rbac:groups=model.samzong.dev,resources=modelsources,verbs=get;list;watch
//+kubebuilder:rbac:groups=dataset.baizeai.io,resources=datasets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop.
func (r *ModelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	model := &modelv1.Model{}
	if err := r.Get(ctx, req.NamespacedName, model); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle deletion
	if !model.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, model)
	}

	// Ensure finalizer
	if err := r.ensureFinalizer(ctx, model); err != nil {
		return ctrl.Result{}, err
	}

	// Get ModelSource
	source := &modelv1.ModelSource{}
	sourceKey := types.NamespacedName{Namespace: model.Namespace, Name: model.Spec.SourceRef}
	if err := r.Get(ctx, sourceKey, source); err != nil {
		return r.updateStatusWithError(ctx, model, "ModelSourceNotFound", fmt.Errorf("get modelsource: %w", err))
	}

	// Check if ModelSource credentials are ready (if secret is required)
	if source.Spec.SecretRef != "" && !r.isCredentialsReady(source) {
		return r.updateStatusWithError(ctx, model, "ModelSourceNotReady", fmt.Errorf("modelsource %s credentials not ready", source.Name))
	}

	// Reconcile versions
	if err := r.reconcileVersions(ctx, model, source); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconcile versions: %w", err)
	}

	// Sync status from Datasets
	if err := r.syncStatus(ctx, model); err != nil {
		return ctrl.Result{}, fmt.Errorf("sync status: %w", err)
	}

	// Handle sharing
	if err := r.reconcileSharing(ctx, model); err != nil {
		return ctrl.Result{}, fmt.Errorf("reconcile sharing: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *ModelReconciler) handleDeletion(ctx context.Context, model *modelv1.Model) (ctrl.Result, error) {
	if !containsString(model.Finalizers, ModelFinalizer) {
		return ctrl.Result{}, nil
	}

	// Delete all reference Datasets (shared Datasets in other namespaces)
	if err := r.deleteReferenceDatasets(ctx, model); err != nil {
		return ctrl.Result{}, fmt.Errorf("delete reference datasets: %w", err)
	}

	// Delete all main Datasets
	if err := r.deleteMainDatasets(ctx, model); err != nil {
		return ctrl.Result{}, fmt.Errorf("delete main datasets: %w", err)
	}

	// Remove finalizer
	model.Finalizers = removeString(model.Finalizers, ModelFinalizer)
	if err := r.Update(ctx, model); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ModelReconciler) ensureFinalizer(ctx context.Context, model *modelv1.Model) error {
	if !containsString(model.Finalizers, ModelFinalizer) {
		model.Finalizers = append(model.Finalizers, ModelFinalizer)
		return r.Update(ctx, model)
	}
	return nil
}

func (r *ModelReconciler) reconcileVersions(ctx context.Context, model *modelv1.Model, source *modelv1.ModelSource) error {
	for _, version := range model.Spec.Versions {
		state := version.State
		if state == "" {
			state = modelv1.ModelVersionStatePresent
		}

		if state == modelv1.ModelVersionStatePresent {
			if err := r.ensureVersionDataset(ctx, model, source, version); err != nil {
				return fmt.Errorf("ensure version %s dataset: %w", version.Name, err)
			}
		} else {
			if err := r.deleteVersionDataset(ctx, model, version.Name); err != nil {
				return fmt.Errorf("delete version %s dataset: %w", version.Name, err)
			}
		}
	}
	return nil
}

func (r *ModelReconciler) ensureVersionDataset(ctx context.Context, model *modelv1.Model, source *modelv1.ModelSource, version modelv1.ModelVersion) error {
	// Build Dataset spec
	spec, err := dataset.BuildDatasetSpec(ctx, r.Client, version, *source, model.Namespace)
	if err != nil {
		return fmt.Errorf("build dataset spec: %w", err)
	}

	// Get dataset name
	datasetName := dataset.GetDatasetName(model.Name, version.Name)

	// Create owner reference
	gvk := modelv1.GroupVersion.WithKind("Model")
	ownerRef := metav1.NewControllerRef(model, gvk)

	// Ensure Dataset
	if err := dataset.EnsureDataset(ctx, r.Client, datasetName, model.Namespace, spec, ownerRef); err != nil {
		return fmt.Errorf("ensure dataset: %w", err)
	}

	return nil
}

func (r *ModelReconciler) deleteVersionDataset(ctx context.Context, model *modelv1.Model, versionName string) error {
	datasetName := dataset.GetDatasetName(model.Name, versionName)
	ds := &datasetv1alpha1.Dataset{}
	key := types.NamespacedName{Name: datasetName, Namespace: model.Namespace}
	if err := r.Get(ctx, key, ds); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	// Mark observedState=ABSENT in status first
	if err := r.markVersionAbsent(ctx, model, versionName); err != nil {
		return err
	}

	// Delete Dataset
	return r.Delete(ctx, ds)
}

func (r *ModelReconciler) markVersionAbsent(ctx context.Context, model *modelv1.Model, versionName string) error {
	status := model.Status.DeepCopy()
	for i, sv := range status.SyncedVersions {
		if sv.Name == versionName {
			status.SyncedVersions[i].ObservedState = modelv1.ModelVersionStateAbsent
			model.Status = *status
			return r.Status().Update(ctx, model)
		}
	}
	return nil
}

func (r *ModelReconciler) syncStatus(ctx context.Context, model *modelv1.Model) error {
	status := model.Status.DeepCopy()
	status.ObservedGeneration = model.Generation

	// Build map of version names from spec
	specVersions := make(map[string]bool)
	for _, v := range model.Spec.Versions {
		specVersions[v.Name] = true
	}

	// Sync each version
	syncedVersions := make([]modelv1.SyncedVersion, 0)
	for _, version := range model.Spec.Versions {
		sv, err := r.syncVersionStatus(ctx, model, version.Name)
		if err != nil {
			return fmt.Errorf("sync version %s status: %w", version.Name, err)
		}
		syncedVersions = append(syncedVersions, *sv)
	}

	// Keep versions that were removed from spec but still exist
	for _, sv := range status.SyncedVersions {
		if !specVersions[sv.Name] && sv.ObservedState != modelv1.ModelVersionStateAbsent {
			// Check if Dataset still exists
			datasetName := dataset.GetDatasetName(model.Name, sv.Name)
			ds := &datasetv1alpha1.Dataset{}
			key := types.NamespacedName{Name: datasetName, Namespace: model.Namespace}
			if err := r.Get(ctx, key, ds); err == nil {
				// Dataset still exists, keep status entry
				sv.ObservedState = modelv1.ModelVersionStateAbsent
				syncedVersions = append(syncedVersions, sv)
			}
		}
	}

	status.SyncedVersions = syncedVersions
	model.Status = *status
	return r.Status().Update(ctx, model)
}

func (r *ModelReconciler) syncVersionStatus(ctx context.Context, model *modelv1.Model, versionName string) (*modelv1.SyncedVersion, error) {
	datasetName := dataset.GetDatasetName(model.Name, versionName)
	ds := &datasetv1alpha1.Dataset{}
	key := types.NamespacedName{Name: datasetName, Namespace: model.Namespace}
	if err := r.Get(ctx, key, ds); err != nil {
		if errors.IsNotFound(err) {
			return &modelv1.SyncedVersion{
				Name:          versionName,
				ObservedState: modelv1.ModelVersionStateAbsent,
			}, nil
		}
		return nil, err
	}

	sv := &modelv1.SyncedVersion{
		Name:          versionName,
		Phase:         string(ds.Status.Phase),
		ActiveDataset: datasetName,
		Conditions:    ds.Status.Conditions,
		ObservedState: modelv1.ModelVersionStatePresent,
	}

	if !ds.Status.LastSyncTime.IsZero() {
		sv.LastSyncTime = &ds.Status.LastSyncTime
	}

	// Get PVC name from Dataset status or spec
	if ds.Status.PVCName != "" {
		sv.PVCName = ds.Status.PVCName
	} else if ds.Spec.VolumeClaimTemplate.Name != "" {
		sv.PVCName = ds.Spec.VolumeClaimTemplate.Name
	}

	// Get observed storage from PVC
	if sv.PVCName != "" {
		pvc := &corev1.PersistentVolumeClaim{}
		pvcKey := types.NamespacedName{Name: sv.PVCName, Namespace: model.Namespace}
		if err := r.Get(ctx, pvcKey, pvc); err == nil {
			if storage, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
				sv.ObservedStorage = &storage
			}
		}
	}

	// Calculate version hash
	versionHash := r.calculateVersionHash(model, versionName)
	sv.ObservedVersionHash = versionHash

	return sv, nil
}

func (r *ModelReconciler) calculateVersionHash(model *modelv1.Model, versionName string) string {
	for _, v := range model.Spec.Versions {
		if v.Name == versionName {
			// Hash: repo + revision + storage spec
			data := fmt.Sprintf("%s:%s:%v", v.Repo, v.Revision, v.Storage)
			hash := sha256.Sum256([]byte(data))
			return hex.EncodeToString(hash[:])[:8]
		}
	}
	return ""
}

func (r *ModelReconciler) reconcileSharing(ctx context.Context, model *modelv1.Model) error {
	// Find all versions with sharing enabled
	for _, version := range model.Spec.Versions {
		if version.Share != nil && version.Share.Enabled {
			if err := r.reconcileVersionSharing(ctx, model, version); err != nil {
				return fmt.Errorf("reconcile version %s sharing: %w", version.Name, err)
			}
		} else {
			// Clean up sharing if disabled
			if err := r.cleanupVersionSharing(ctx, model, version.Name); err != nil {
				return fmt.Errorf("cleanup version %s sharing: %w", version.Name, err)
			}
		}
	}
	return nil
}

func (r *ModelReconciler) reconcileVersionSharing(ctx context.Context, model *modelv1.Model, version modelv1.ModelVersion) error {
	// Get source dataset name
	sourceDatasetName := dataset.GetDatasetName(model.Name, version.Name)

	// Find matching namespaces
	namespaces, err := r.findMatchingNamespaces(ctx, version.Share)
	if err != nil {
		return fmt.Errorf("find matching namespaces: %w", err)
	}

	// Create REFERENCE Datasets
	labels := map[string]string{
		"modelfs.samzong.dev/model":   fmt.Sprintf("%s/%s", model.Namespace, model.Name),
		"modelfs.samzong.dev/version": version.Name,
	}

	for _, ns := range namespaces {
		if ns == model.Namespace {
			continue // Skip source namespace
		}

		targetName := dataset.GetReferenceDatasetName(model.Namespace, model.Name, version.Name)
		if err := dataset.EnsureReferenceDataset(ctx, r.Client, model.Namespace, sourceDatasetName, ns, targetName, labels); err != nil {
			return fmt.Errorf("ensure reference dataset in %s: %w", ns, err)
		}
	}

	return nil
}

func (r *ModelReconciler) findMatchingNamespaces(ctx context.Context, share *modelv1.ShareSpec) ([]string, error) {
	nsList := &corev1.NamespaceList{}
	if err := r.List(ctx, nsList); err != nil {
		return nil, err
	}

	var matching []string
	for _, ns := range nsList.Items {
		// Check namespace selector
		if share.NamespaceSelector != nil {
			selector, err := metav1.LabelSelectorAsSelector(share.NamespaceSelector)
			if err != nil {
				return nil, err
			}
			if !selector.Matches(labels.Set(ns.Labels)) {
				continue
			}
		}

		// Check opt-in label
		if share.RequireOptInLabel != "" {
			key := share.RequireOptInLabel
			value := "true"
			if idx := len(key) - 1; idx >= 0 && key[idx] == '=' {
				// Format: "key=value"
				parts := splitOptInLabel(share.RequireOptInLabel)
				if len(parts) == 2 {
					key, value = parts[0], parts[1]
				}
			}
			if ns.Labels[key] != value {
				continue
			}
		}

		matching = append(matching, ns.Name)
	}

	return matching, nil
}

func splitOptInLabel(s string) []string {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '=' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

func (r *ModelReconciler) cleanupVersionSharing(ctx context.Context, model *modelv1.Model, versionName string) error {
	// Find all REFERENCE Datasets with matching labels
	datasetList := &datasetv1alpha1.DatasetList{}
	labelSelector := client.MatchingLabels{
		"modelfs.samzong.dev/model":   fmt.Sprintf("%s/%s", model.Namespace, model.Name),
		"modelfs.samzong.dev/version": versionName,
	}
	if err := r.List(ctx, datasetList, labelSelector); err != nil {
		return err
	}

	for _, ds := range datasetList.Items {
		if ds.Spec.Source.Type == datasetv1alpha1.DatasetTypeReference {
			if err := r.Delete(ctx, &ds); err != nil && !errors.IsNotFound(err) {
				return err
			}
		}
	}

	return nil
}

func (r *ModelReconciler) deleteReferenceDatasets(ctx context.Context, model *modelv1.Model) error {
	// Find all REFERENCE Datasets with matching labels
	datasetList := &datasetv1alpha1.DatasetList{}
	labelSelector := client.MatchingLabels{
		"modelfs.samzong.dev/model": fmt.Sprintf("%s/%s", model.Namespace, model.Name),
	}
	if err := r.List(ctx, datasetList, labelSelector); err != nil {
		return err
	}

	for _, ds := range datasetList.Items {
		if ds.Spec.Source.Type == datasetv1alpha1.DatasetTypeReference {
			if err := r.Delete(ctx, &ds); err != nil && !errors.IsNotFound(err) {
				return err
			}
		}
	}

	return nil
}

func (r *ModelReconciler) deleteMainDatasets(ctx context.Context, model *modelv1.Model) error {
	// Delete all Datasets owned by this Model
	datasetList := &datasetv1alpha1.DatasetList{}
	if err := r.List(ctx, datasetList, client.InNamespace(model.Namespace)); err != nil {
		return err
	}

	for _, ds := range datasetList.Items {
		for _, ownerRef := range ds.OwnerReferences {
			if ownerRef.Kind == "Model" && ownerRef.Name == model.Name {
				if err := r.Delete(ctx, &ds); err != nil && !errors.IsNotFound(err) {
					return err
				}
			}
		}
	}

	return nil
}

func (r *ModelReconciler) isCredentialsReady(source *modelv1.ModelSource) bool {
	for _, cond := range source.Status.Conditions {
		if cond.Type == "CredentialsReady" && cond.Status == metav1.ConditionTrue {
			return true
		}
	}
	return false
}

func (r *ModelReconciler) updateStatusWithError(ctx context.Context, model *modelv1.Model, reason string, err error) (ctrl.Result, error) {
	status := model.Status.DeepCopy()
	now := metav1.Now()
	condition := metav1.Condition{
		Type:               "ReconcileError",
		Status:             metav1.ConditionFalse,
		Reason:             reason,
		Message:            err.Error(),
		LastTransitionTime: now,
		ObservedGeneration: model.Generation,
	}

	// Update condition
	found := false
	for i, c := range status.Conditions {
		if c.Type == "ReconcileError" {
			status.Conditions[i] = condition
			found = true
			break
		}
	}
	if !found {
		status.Conditions = append(status.Conditions, condition)
	}

	model.Status = *status
	if updateErr := r.Status().Update(ctx, model); updateErr != nil {
		return ctrl.Result{}, fmt.Errorf("update status: %w (original error: %v)", updateErr, err)
	}

	return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *ModelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&modelv1.Model{}).
		Watches(
			&modelv1.ModelSource{},
			handler.EnqueueRequestsFromMapFunc(r.mapModelSourceToModel),
		).
		Watches(
			&datasetv1alpha1.Dataset{},
			handler.EnqueueRequestsFromMapFunc(r.mapDatasetToModel),
		).
		Watches(
			&corev1.Namespace{},
			handler.EnqueueRequestsFromMapFunc(r.mapNamespaceToModel),
		).
		Complete(r)
}

func (r *ModelReconciler) mapModelSourceToModel(ctx context.Context, obj client.Object) []reconcile.Request {
	source, ok := obj.(*modelv1.ModelSource)
	if !ok {
		return []reconcile.Request{}
	}

	// Find all Models that reference this ModelSource
	modelList := &modelv1.ModelList{}
	if err := r.List(ctx, modelList, client.InNamespace(source.Namespace)); err != nil {
		return []reconcile.Request{}
	}

	var requests []reconcile.Request
	for _, model := range modelList.Items {
		if model.Spec.SourceRef == source.Name {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      model.Name,
					Namespace: model.Namespace,
				},
			})
		}
	}

	return requests
}

func (r *ModelReconciler) mapDatasetToModel(ctx context.Context, obj client.Object) []reconcile.Request {
	ds, ok := obj.(*datasetv1alpha1.Dataset)
	if !ok {
		return []reconcile.Request{}
	}

	// Find Model that owns this Dataset
	for _, ownerRef := range ds.OwnerReferences {
		if ownerRef.Kind == "Model" {
			return []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Name:      ownerRef.Name,
						Namespace: ds.Namespace,
					},
				},
			}
		}
	}

	// Also check labels for REFERENCE Datasets
	if modelLabel, ok := ds.Labels["modelfs.samzong.dev/model"]; ok {
		// Format: namespace/name
		parts := splitModelLabel(modelLabel)
		if len(parts) == 2 {
			return []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Name:      parts[1],
						Namespace: parts[0],
					},
				},
			}
		}
	}

	return []reconcile.Request{}
}

func (r *ModelReconciler) mapNamespaceToModel(ctx context.Context, obj client.Object) []reconcile.Request {
	ns, ok := obj.(*corev1.Namespace)
	if !ok {
		return []reconcile.Request{}
	}

	// Find all Models with sharing enabled
	modelList := &modelv1.ModelList{}
	if err := r.List(ctx, modelList); err != nil {
		return []reconcile.Request{}
	}

	var requests []reconcile.Request
	for _, model := range modelList.Items {
		for _, version := range model.Spec.Versions {
			if version.Share != nil && version.Share.Enabled {
				// Check if namespace matches selector
				if version.Share.NamespaceSelector != nil {
					selector, err := metav1.LabelSelectorAsSelector(version.Share.NamespaceSelector)
					if err == nil && selector.Matches(labels.Set(ns.Labels)) {
						requests = append(requests, reconcile.Request{
							NamespacedName: types.NamespacedName{
								Name:      model.Name,
								Namespace: model.Namespace,
							},
						})
						break
					}
				}
			}
		}
	}

	return requests
}

func splitModelLabel(s string) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}
