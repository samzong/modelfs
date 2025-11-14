package provider

import (
	"context"
	modelv1 "github.com/samzong/modelfs/api/v1"
	"github.com/samzong/modelfs/pkg/ui/api"
	"time"
)

type mockStore struct {
	models  map[string]api.ModelDetail
	sources map[string]api.ModelSourceSummary
}

func NewMockStore() Store {
	return &mockStore{models: map[string]api.ModelDetail{}, sources: map[string]api.ModelSourceSummary{}}
}

func (m *mockStore) ListModels(ctx context.Context, namespace string) ([]api.ModelSummary, error) {
	var out []api.ModelSummary
	for _, d := range m.models {
		if d.Summary.Namespace == namespace {
			out = append(out, d.Summary)
		}
	}
	return out, nil
}

func (m *mockStore) GetModel(ctx context.Context, namespace, name string) (api.ModelDetail, error) {
	key := namespace + "/" + name
	if d, ok := m.models[key]; ok {
		return d, nil
	}
	return api.ModelDetail{Summary: api.ModelSummary{Name: name, Namespace: namespace, SourceRef: "mock", VersionsReady: 0, VersionsTotal: 0, LastSyncTime: time.Now(), Status: api.PhaseUnknown}}, nil
}

func (m *mockStore) ListModelSources(ctx context.Context, namespace string) ([]api.ModelSourceSummary, error) {
	var out []api.ModelSourceSummary
	for _, s := range m.sources {
		if s.Namespace == namespace {
			out = append(out, s)
		}
	}
	return out, nil
}

func (m *mockStore) GetModelSource(ctx context.Context, namespace, name string) (*modelv1.ModelSource, error) {
	return &modelv1.ModelSource{ObjectMeta: modelv1.ModelSource{}.ObjectMeta, Spec: modelv1.ModelSourceSpec{Type: "HUGGING_FACE"}}, nil
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
	s := api.ModelSourceSummary{Name: name, Namespace: namespace, Type: spec.Type, SecretRef: spec.SecretRef, CredentialsReady: true, LastChecked: time.Now()}
	m.sources[namespace+"/"+name] = s
	return nil
}

func (m *mockStore) UpdateModelSource(ctx context.Context, namespace, name string, spec modelv1.ModelSourceSpec) error {
	s := api.ModelSourceSummary{Name: name, Namespace: namespace, Type: spec.Type, SecretRef: spec.SecretRef, CredentialsReady: true, LastChecked: time.Now()}
	m.sources[namespace+"/"+name] = s
	return nil
}

func (m *mockStore) DeleteModelSource(ctx context.Context, namespace, name string) error {
	delete(m.sources, namespace+"/"+name)
	return nil
}

func (m *mockStore) CreateModel(ctx context.Context, namespace, name string, spec modelv1.ModelSpec) error {
	d := mockDetailFromSpec(namespace, name, spec)
	m.models[namespace+"/"+name] = d
	return nil
}

func (m *mockStore) UpdateModel(ctx context.Context, namespace, name string, spec modelv1.ModelSpec) error {
	d := mockDetailFromSpec(namespace, name, spec)
	m.models[namespace+"/"+name] = d
	return nil
}

func (m *mockStore) ValidateSecret(ctx context.Context, namespace, name string) (bool, string, error) {
	return true, "ok", nil
}

func (m *mockStore) ListDatasets(ctx context.Context, namespace string) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func mockDetailFromSpec(ns, name string, spec modelv1.ModelSpec) api.ModelDetail {
	versions := make([]api.ModelVersionView, 0, len(spec.Versions))
	for _, v := range spec.Versions {
		versions = append(versions, api.ModelVersionView{Name: v.Name, Repo: v.Repo, Revision: v.Revision, Precision: v.Precision, DesiredState: string(v.State), ShareEnabled: v.Share != nil && v.Share.Enabled, DatasetPhase: api.PhasePending})
	}
	var tags []string
	var desc string
	if spec.Display != nil {
		tags = append(tags, spec.Display.Tags...)
		desc = spec.Display.Description
	}
	summary := api.ModelSummary{Name: name, Namespace: ns, SourceRef: spec.SourceRef, Tags: tags, VersionsReady: 0, VersionsTotal: len(versions), LastSyncTime: time.Now(), Status: api.PhasePending}
	return api.ModelDetail{Summary: summary, Description: desc, Versions: versions}
}

func namespaceFallback(ctx context.Context) string { return "model-system" }
