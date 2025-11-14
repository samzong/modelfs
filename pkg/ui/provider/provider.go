package provider

import (
	"context"
	modelv1 "github.com/samzong/modelfs/api/v1"
	"github.com/samzong/modelfs/pkg/ui/api"
)

type Store interface {
	ListModels(ctx context.Context, namespace string) ([]api.ModelSummary, error)
	GetModel(ctx context.Context, namespace, name string) (api.ModelDetail, error)
	ListModelSources(ctx context.Context, namespace string) ([]api.ModelSourceSummary, error)
	GetModelSource(ctx context.Context, namespace, name string) (*modelv1.ModelSource, error)
	ListNamespaces(ctx context.Context) ([]api.NamespaceInfo, error)
	ListErrors(ctx context.Context, namespace string) ([]api.ErrorBanner, error)
	Watch(ctx context.Context, namespace string) (<-chan api.SSEPayload, error)
	DeleteModel(ctx context.Context, namespace, name string) error
	DeleteModelVersion(ctx context.Context, namespace, modelName, versionName string) error
	ToggleVersionShare(ctx context.Context, namespace, modelName, versionName string, enabled bool) error
	TriggerResync(ctx context.Context, namespace, modelName string) error
	CreateModelSource(ctx context.Context, namespace, name string, spec modelv1.ModelSourceSpec) error
	UpdateModelSource(ctx context.Context, namespace, name string, spec modelv1.ModelSourceSpec) error
	DeleteModelSource(ctx context.Context, namespace, name string) error
	CreateModel(ctx context.Context, namespace, name string, spec modelv1.ModelSpec) error
	UpdateModel(ctx context.Context, namespace, name string, spec modelv1.ModelSpec) error
	ValidateSecret(ctx context.Context, namespace, name string) (bool, string, error)
	ListDatasets(ctx context.Context, namespace string) ([]map[string]interface{}, error)
}
