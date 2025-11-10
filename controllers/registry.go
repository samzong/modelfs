package controllers

import (
        "context"
        "fmt"

        modelv1 "github.com/samzong/modelfs/api/v1"
)

// ModelRegistry resolves Model and ModelSource references for reconciliation.
type ModelRegistry interface {
        GetModel(ctx context.Context, namespace, name string) (modelv1.Model, error)
        GetModelSource(ctx context.Context, namespace, name string) (modelv1.ModelSource, error)
}

// StaticRegistry stores resources in-memory for controller coordination without requiring a Kubernetes API server.
type StaticRegistry struct {
        Models  map[string]modelv1.Model
        Sources map[string]modelv1.ModelSource
}

// NewStaticRegistry initialises a StaticRegistry with optional seed data.
func NewStaticRegistry() *StaticRegistry {
        return &StaticRegistry{
                Models:  map[string]modelv1.Model{},
                Sources: map[string]modelv1.ModelSource{},
        }
}

// GetModel retrieves a model by namespace/name.
func (r *StaticRegistry) GetModel(_ context.Context, namespace, name string) (modelv1.Model, error) {
        key := fmt.Sprintf("%s/%s", namespace, name)
        model, ok := r.Models[key]
        if !ok {
                return modelv1.Model{}, fmt.Errorf("model %s not found", key)
        }
        return model, nil
}

// GetModelSource retrieves a model source by namespace/name.
func (r *StaticRegistry) GetModelSource(_ context.Context, namespace, name string) (modelv1.ModelSource, error) {
        key := fmt.Sprintf("%s/%s", namespace, name)
        source, ok := r.Sources[key]
        if !ok {
                return modelv1.ModelSource{}, fmt.Errorf("model source %s not found", key)
        }
        return source, nil
}

// SetModel stores or updates a model entry.
func (r *StaticRegistry) SetModel(model modelv1.Model) {
        if r.Models == nil {
                r.Models = map[string]modelv1.Model{}
        }
        key := fmt.Sprintf("%s/%s", model.Metadata.Namespace, model.Metadata.Name)
        r.Models[key] = model
}

// SetModelSource stores or updates a model source entry.
func (r *StaticRegistry) SetModelSource(source modelv1.ModelSource) {
        if r.Sources == nil {
                r.Sources = map[string]modelv1.ModelSource{}
        }
        key := fmt.Sprintf("%s/%s", source.Metadata.Namespace, source.Metadata.Name)
        r.Sources[key] = source
}
