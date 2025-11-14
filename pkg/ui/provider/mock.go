package provider

import (
	"context"
	modelv1 "github.com/samzong/modelfs/api/v1"
	"github.com/samzong/modelfs/pkg/ui/api"
	"time"
)

type mockStore struct{}

func NewMockStore() Store { return &mockStore{} }

func (m *mockStore) ListModels(ctx context.Context, namespace string) ([]api.ModelSummary, error) {
	return []api.ModelSummary{}, nil
}

func (m *mockStore) GetModel(ctx context.Context, namespace, name string) (api.ModelDetail, error) {
	return api.ModelDetail{Summary: api.ModelSummary{Name: name, Namespace: namespace, SourceRef: "mock", VersionsReady: 0, VersionsTotal: 0, LastSyncTime: time.Now(), Status: api.PhaseUnknown}}, nil
}

func (m *mockStore) ListModelSources(ctx context.Context, namespace string) ([]api.ModelSourceSummary, error) {
	return []api.ModelSourceSummary{}, nil
}

func (m *mockStore) ListNamespaces(ctx context.Context) ([]api.NamespaceInfo, error) {
	return []api.NamespaceInfo{{Name: namespaceFallback(ctx)}}, nil
}

func (m *mockStore) ListErrors(ctx context.Context, namespace string) ([]api.ErrorBanner, error) {
	return []api.ErrorBanner{}, nil
}

func (m *mockStore) Watch(ctx context.Context, namespace string) (<-chan api.SSEPayload, error) {
	ch := make(chan api.SSEPayload)
	go func() { defer close(ch); <-ctx.Done() }()
	return ch, nil
}

func (m *mockStore) DeleteModel(ctx context.Context, namespace, name string) error { return nil }
func (m *mockStore) DeleteModelVersion(ctx context.Context, namespace, modelName, versionName string) error {
	return nil
}
func (m *mockStore) ToggleVersionShare(ctx context.Context, namespace, modelName, versionName string, enabled bool) error {
	return nil
}
func (m *mockStore) TriggerResync(ctx context.Context, namespace, modelName string) error { return nil }
func (m *mockStore) CreateModelSource(ctx context.Context, namespace, name string, spec modelv1.ModelSourceSpec) error {
	return nil
}

func namespaceFallback(ctx context.Context) string { return "model-system" }
