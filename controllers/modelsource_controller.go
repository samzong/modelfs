package controllers

import (
	"context"
	"fmt"

	modelv1 "github.com/samzong/modelfs/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ModelSourceFinalizer = "modelfs.samzong.dev/modelsource-finalizer"
)

// ModelSourceReconciler reconciles a ModelSource object
type ModelSourceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=model.samzong.dev,resources=modelsources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=model.samzong.dev,resources=modelsources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=model.samzong.dev,resources=modelsources/finalizers,verbs=update
//+kubebuilder:rbac:groups=model.samzong.dev,resources=models,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop.
func (r *ModelSourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	source := &modelv1.ModelSource{}
	if err := r.Get(ctx, req.NamespacedName, source); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle deletion
	if !source.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, source)
	}

	// Ensure finalizer
	if err := r.ensureFinalizer(ctx, source); err != nil {
		return ctrl.Result{}, err
	}

	// Validate Secret (optional for some source types)
	if source.Spec.SecretRef != "" {
		if err := r.validateSecret(ctx, source); err != nil {
			if updateErr := r.updateCredentialsCondition(ctx, source, false, "SecretInvalid", err.Error()); updateErr != nil {
				return ctrl.Result{}, updateErr
			}
			return ctrl.Result{}, err
		}
		// Update CredentialsReady condition
		if err := r.updateCredentialsCondition(ctx, source, true, "SecretValid", "Secret is valid and accessible"); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		// No secret required (e.g., public HuggingFace models)
		if err := r.updateCredentialsCondition(ctx, source, true, "NoSecretRequired", "No secret required for this source type"); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Find and update referenced Models
	if err := r.updateReferencedBy(ctx, source); err != nil {
		return ctrl.Result{}, fmt.Errorf("update referencedBy: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *ModelSourceReconciler) handleDeletion(ctx context.Context, source *modelv1.ModelSource) (ctrl.Result, error) {
	// Check if any Model references this ModelSource
	models, err := r.findReferencingModels(ctx, source)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("find referencing models: %w", err)
	}

	if len(models) > 0 {
		// Block deletion: update status with referencedBy
		referencedBy := make([]string, 0, len(models))
		for _, m := range models {
			referencedBy = append(referencedBy, formatNamespacedName(m.Namespace, m.Name))
		}
		status := source.Status.DeepCopy()
		status.ReferencedBy = referencedBy
		source.Status = *status
		if err := r.Status().Update(ctx, source); err != nil {
			return ctrl.Result{}, err
		}
		// Requeue to retry deletion later
		return ctrl.Result{Requeue: true}, nil
	}

	// No references, remove finalizer
	if containsString(source.Finalizers, ModelSourceFinalizer) {
		source.Finalizers = removeString(source.Finalizers, ModelSourceFinalizer)
		if err := r.Update(ctx, source); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ModelSourceReconciler) ensureFinalizer(ctx context.Context, source *modelv1.ModelSource) error {
	if !containsString(source.Finalizers, ModelSourceFinalizer) {
		source.Finalizers = append(source.Finalizers, ModelSourceFinalizer)
		return r.Update(ctx, source)
	}
	return nil
}

func (r *ModelSourceReconciler) validateSecret(ctx context.Context, source *modelv1.ModelSource) error {
	secret := &corev1.Secret{}
	key := types.NamespacedName{Namespace: source.Namespace, Name: source.Spec.SecretRef}
	if err := r.Get(ctx, key, secret); err != nil {
		return fmt.Errorf("get secret %s: %w", source.Spec.SecretRef, err)
	}
	// Basic validation: secret exists and is not empty
	if len(secret.Data) == 0 && len(secret.StringData) == 0 {
		return fmt.Errorf("secret %s is empty", source.Spec.SecretRef)
	}
	return nil
}

func (r *ModelSourceReconciler) updateCredentialsCondition(ctx context.Context, source *modelv1.ModelSource, ready bool, reason, message string) error {
	status := source.Status.DeepCopy()
	now := metav1.Now()
	condition := metav1.Condition{
		Type:               "CredentialsReady",
		Status:             map[bool]metav1.ConditionStatus{true: metav1.ConditionTrue, false: metav1.ConditionFalse}[ready],
		Reason:             reason,
		Message:            message,
		LastTransitionTime: now,
		ObservedGeneration: source.Generation,
	}

	// Update or add condition
	found := false
	for i, c := range status.Conditions {
		if c.Type == "CredentialsReady" {
			if c.Status == condition.Status && c.Reason == condition.Reason && c.Message == condition.Message {
				return nil // No change
			}
			if c.Status != condition.Status {
				condition.LastTransitionTime = now
			} else {
				condition.LastTransitionTime = c.LastTransitionTime
			}
			status.Conditions[i] = condition
			found = true
			break
		}
	}
	if !found {
		status.Conditions = append(status.Conditions, condition)
	}

	source.Status = *status
	return r.Status().Update(ctx, source)
}

func (r *ModelSourceReconciler) findReferencingModels(ctx context.Context, source *modelv1.ModelSource) ([]modelv1.Model, error) {
	modelList := &modelv1.ModelList{}
	if err := r.List(ctx, modelList, client.InNamespace(source.Namespace)); err != nil {
		return nil, err
	}

	var referencing []modelv1.Model
	for _, model := range modelList.Items {
		if model.Spec.SourceRef == source.Name {
			referencing = append(referencing, model)
		}
	}
	return referencing, nil
}

func (r *ModelSourceReconciler) updateReferencedBy(ctx context.Context, source *modelv1.ModelSource) error {
	models, err := r.findReferencingModels(ctx, source)
	if err != nil {
		return err
	}

	referencedBy := make([]string, 0, len(models))
	for _, m := range models {
		referencedBy = append(referencedBy, fmt.Sprintf("%s/%s", m.Namespace, m.Name))
	}

	status := source.Status.DeepCopy()
	status.ReferencedBy = referencedBy
	source.Status = *status
	return r.Status().Update(ctx, source)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ModelSourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Index Models by sourceRef for efficient lookup
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &modelv1.Model{}, "spec.sourceRef", func(obj client.Object) []string {
		model := obj.(*modelv1.Model)
		return []string{model.Spec.SourceRef}
	}); err != nil {
		return fmt.Errorf("index field spec.sourceRef: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&modelv1.ModelSource{}).
		Complete(r)
}
