package controllers

import (
	"context"
	"fmt"

	modelv1 "github.com/samzong/modelfs/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ModelRegistry resolves Model and ModelSource references for reconciliation.
type ModelRegistry interface {
	GetModel(ctx context.Context, namespace, name string) (*modelv1.Model, error)
	GetModelSource(ctx context.Context, namespace, name string) (*modelv1.ModelSource, error)
}

// KubernetesRegistry uses Kubernetes API to resolve Model and ModelSource references.
type KubernetesRegistry struct {
	client client.Client
}

// NewKubernetesRegistry creates a new Kubernetes-backed registry.
func NewKubernetesRegistry(c client.Client) *KubernetesRegistry {
	return &KubernetesRegistry{client: c}
}

// GetModel retrieves a model by namespace/name.
func (r *KubernetesRegistry) GetModel(ctx context.Context, namespace, name string) (*modelv1.Model, error) {
	model := &modelv1.Model{}
	key := client.ObjectKey{Namespace: namespace, Name: name}
	if err := r.client.Get(ctx, key, model); err != nil {
		return nil, fmt.Errorf("get model %s/%s: %w", namespace, name, err)
	}
	return model, nil
}

// GetModelSource retrieves a model source by namespace/name.
func (r *KubernetesRegistry) GetModelSource(ctx context.Context, namespace, name string) (*modelv1.ModelSource, error) {
	source := &modelv1.ModelSource{}
	key := client.ObjectKey{Namespace: namespace, Name: name}
	if err := r.client.Get(ctx, key, source); err != nil {
		return nil, fmt.Errorf("get model source %s/%s: %w", namespace, name, err)
	}
	return source, nil
}
