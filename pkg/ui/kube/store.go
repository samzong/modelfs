package kube

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	modelv1 "github.com/samzong/modelfs/api/v1"
	"github.com/samzong/modelfs/pkg/ui/api"
	"github.com/samzong/modelfs/pkg/ui/provider"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ provider.Store = (*Store)(nil)

type Store struct{ client client.WithWatch }

func NewStore(cfg *rest.Config) (*Store, error) {
	cl, err := client.NewWithWatch(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}
	return &Store{client: cl}, nil
}

func (s *Store) ListModels(ctx context.Context, namespace string) ([]api.ModelSummary, error) {
	models := &modelv1.ModelList{}
	if err := s.client.List(ctx, models, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	items := make([]api.ModelSummary, 0, len(models.Items))
	for i := range models.Items {
		items = append(items, summarizeModel(&models.Items[i]))
	}
	return items, nil
}

func (s *Store) GetModel(ctx context.Context, namespace, name string) (api.ModelDetail, error) {
	model := &modelv1.Model{}
	if err := s.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, model); err != nil {
		return api.ModelDetail{}, err
	}
	return modelDetail(model), nil
}

func (s *Store) ListModelSources(ctx context.Context, namespace string) ([]api.ModelSourceSummary, error) {
	list := &modelv1.ModelSourceList{}
	if err := s.client.List(ctx, list, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	items := make([]api.ModelSourceSummary, 0, len(list.Items))
	for i := range list.Items {
		items = append(items, summarizeModelSource(&list.Items[i]))
	}
	return items, nil
}

func (s *Store) ListNamespaces(ctx context.Context) ([]api.NamespaceInfo, error) {
	list := &corev1.NamespaceList{}
	if err := s.client.List(ctx, list); err != nil {
		return nil, err
	}
	items := make([]api.NamespaceInfo, 0, len(list.Items))
	for i := range list.Items {
		items = append(items, api.NamespaceInfo{Name: list.Items[i].Name})
	}
	return items, nil
}

func (s *Store) ListErrors(ctx context.Context, namespace string) ([]api.ErrorBanner, error) {
	models := &modelv1.ModelList{}
	if err := s.client.List(ctx, models, client.InNamespace(namespace)); err != nil {
		return nil, err
	}
	var banners []api.ErrorBanner
	for i := range models.Items {
		if message, reason, retry := reconcileError(&models.Items[i]); message != "" {
			banners = append(banners, api.ErrorBanner{Namespace: namespace, Message: message, Reason: reason, RetryAt: retry})
		}
	}
	return banners, nil
}

func (s *Store) Watch(ctx context.Context, namespace string) (<-chan api.SSEPayload, error) {
	modelWatcher, err := s.client.Watch(ctx, &modelv1.ModelList{}, client.InNamespace(namespace))
	if err != nil {
		return nil, err
	}
	sourceWatcher, err := s.client.Watch(ctx, &modelv1.ModelSourceList{}, client.InNamespace(namespace))
	if err != nil {
		modelWatcher.Stop()
		return nil, err
	}
	out := make(chan api.SSEPayload)
	go func() {
		defer close(out)
		defer modelWatcher.Stop()
		defer sourceWatcher.Stop()
		var wg sync.WaitGroup
		wg.Add(2)
		go s.pipeWatch(ctx, &wg, "models", modelWatcher, out)
		go s.pipeWatch(ctx, &wg, "modelsources", sourceWatcher, out)
		wg.Wait()
	}()
	return out, nil
}

func (s *Store) pipeWatch(ctx context.Context, wg *sync.WaitGroup, resource string, watcher watch.Interface, out chan<- api.SSEPayload) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-watcher.ResultChan():
			if !ok {
				return
			}
			payload, ok := s.payloadForEvent(resource, evt)
			if !ok {
				continue
			}
			select {
			case out <- payload:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (s *Store) payloadForEvent(resource string, evt watch.Event) (api.SSEPayload, bool) {
	action := strings.ToLower(string(evt.Type))
	switch obj := evt.Object.(type) {
	case *modelv1.Model:
		summary := summarizeModel(obj)
		return api.SSEPayload{Resource: resource, Action: action, Payload: summary}, true
	case *modelv1.ModelSource:
		summary := summarizeModelSource(obj)
		return api.SSEPayload{Resource: resource, Action: action, Payload: summary}, true
	default:
		return api.SSEPayload{}, false
	}
}

func summarizeModelSource(src *modelv1.ModelSource) api.ModelSourceSummary {
	ready, msg := credentialsStatus(src.Status.Conditions)
	return api.ModelSourceSummary{
		Name: src.Name, Namespace: src.Namespace, Type: src.Spec.Type, SecretRef: src.Spec.SecretRef,
		CredentialsReady: ready, CredentialsStatus: msg, ReferencedModels: append([]string{}, src.Status.ReferencedBy...), LastChecked: lastConditionTime(src.Status.Conditions),
	}
}

func credentialsStatus(conditions []metav1.Condition) (bool, string) {
	for _, c := range conditions {
		if c.Type == "CredentialsReady" {
			return c.Status == metav1.ConditionTrue, c.Message
		}
	}
	return false, "unknown"
}

func lastConditionTime(conditions []metav1.Condition) time.Time {
	var latest time.Time
	for _, c := range conditions {
		if c.LastTransitionTime.After(latest) {
			latest = c.LastTransitionTime.Time
		}
	}
	return latest
}

func reconcileError(model *modelv1.Model) (message string, reason string, retryAt time.Time) {
	for _, c := range model.Status.Conditions {
		if c.Type == "ReconcileError" && c.Status == metav1.ConditionFalse {
			return c.Message, c.Reason, c.LastTransitionTime.Time.Add(1 * time.Minute)
		}
	}
	return "", "", time.Time{}
}

func (s *Store) DeleteModel(ctx context.Context, namespace, name string) error {
	model := &modelv1.Model{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name}}
	return s.client.Delete(ctx, model)
}

func (s *Store) DeleteModelVersion(ctx context.Context, namespace, modelName, versionName string) error {
	model := &modelv1.Model{}
	if err := s.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: modelName}, model); err != nil {
		return err
	}
	updated := false
	for i := range model.Spec.Versions {
		if model.Spec.Versions[i].Name == versionName {
			state := model.Spec.Versions[i].State
			if state == "" || state == modelv1.ModelVersionStatePresent {
				model.Spec.Versions[i].State = modelv1.ModelVersionStateAbsent
			}
			updated = true
			break
		}
	}
	if !updated {
		return fmt.Errorf("version %s not found on model %s", versionName, modelName)
	}
	return s.client.Update(ctx, model)
}

func (s *Store) ToggleVersionShare(ctx context.Context, namespace, modelName, versionName string, enabled bool) error {
	model := &modelv1.Model{}
	if err := s.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: modelName}, model); err != nil {
		return err
	}
	found := false
	for i := range model.Spec.Versions {
		if model.Spec.Versions[i].Name == versionName {
			found = true
			if model.Spec.Versions[i].Share == nil {
				model.Spec.Versions[i].Share = &modelv1.ShareSpec{}
			}
			model.Spec.Versions[i].Share.Enabled = enabled
			break
		}
	}
	if !found {
		return fmt.Errorf("version %s not found on model %s", versionName, modelName)
	}
	return s.client.Update(ctx, model)
}

func (s *Store) TriggerResync(ctx context.Context, namespace, modelName string) error {
	model := &modelv1.Model{}
	if err := s.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: modelName}, model); err != nil {
		return err
	}
	if model.Annotations == nil {
		model.Annotations = map[string]string{}
	}
	model.Annotations["modelfs.samzong.dev/resyncAt"] = time.Now().UTC().Format(time.RFC3339Nano)
	return s.client.Update(ctx, model)
}
