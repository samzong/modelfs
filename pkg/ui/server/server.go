package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"io/fs"
	"path"
	"time"

	"github.com/go-chi/chi/v5"
	modelv1 "github.com/samzong/modelfs/api/v1"
	"github.com/samzong/modelfs/pkg/ui/api"
	"github.com/samzong/modelfs/pkg/ui/static"
	"github.com/samzong/modelfs/pkg/ui/provider"
)

type Config struct {
	DefaultNamespace  string
	ReadHeaderTimeout time.Duration
}

type Server struct {
	cfg   Config
	store provider.Store
}

type Option func(*Server)

func WithProvider(store provider.Store) Option {
	return func(s *Server) { s.store = store }
}

func New(cfg Config, opts ...Option) *Server {
	if cfg.DefaultNamespace == "" {
		cfg.DefaultNamespace = "model-system"
	}
	if cfg.ReadHeaderTimeout == 0 {
		cfg.ReadHeaderTimeout = 15 * time.Second
	}
	s := &Server{cfg: cfg}
	for _, opt := range opts {
		opt(s)
	}
	if s.store == nil {
		s.store = provider.NewMockStore()
	}
	return s
}

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				origin = "*"
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Route("/api", func(apiRouter chi.Router) {
		apiRouter.Get("/models", s.handleModelsList)
		apiRouter.Get("/models/{namespace}/{name}", s.handleModelDetail)
		apiRouter.Post("/models", s.handleModelCreate)
		apiRouter.Put("/models/{namespace}/{name}", s.handleModelUpdate)
		apiRouter.Delete("/models/{namespace}/{name}", s.handleModelDelete)
		apiRouter.Delete("/models/{namespace}/{name}/versions/{version}", s.handleModelVersionDelete)
		apiRouter.Post("/models/{namespace}/{name}/versions/{version}/share", s.handleShareToggle)
		apiRouter.Post("/models/{namespace}/{name}/actions/resync", s.handleResync)
		apiRouter.Get("/modelsources", s.handleModelSources)
		apiRouter.Get("/modelsources/{namespace}/{name}", s.handleModelSourceDetail)
		apiRouter.Post("/modelsources", s.handleModelSourceCreate)
		apiRouter.Put("/modelsources/{namespace}/{name}", s.handleModelSourceUpdate)
		apiRouter.Delete("/modelsources/{namespace}/{name}", s.handleModelSourceDelete)
		apiRouter.Get("/secrets/validate", s.handleSecretValidate)
		apiRouter.Get("/datasets", s.handleDatasets)
		apiRouter.Get("/namespaces", s.handleNamespaces)
		apiRouter.Get("/errors", s.handleErrors)
		apiRouter.Get("/sse", s.handleSSE)
	})

    fsys := static.Files
    dist, _ := fs.Sub(fsys, "dist")
	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if p == "/" {
			b, err := fs.ReadFile(dist, "index.html")
			if err == nil {
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(b)
				return
			}
		}
		name := strings.TrimPrefix(p, "/")
		if exists(dist, name) {
			// serve file bytes with simple content-type
			b, err := fs.ReadFile(dist, name)
			if err == nil {
				ct := contentTypeByExt(path.Ext(name))
				if ct != "" { w.Header().Set("Content-Type", ct) }
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(b)
				return
			}
		}
		// fallback to index.html for SPA
		b, err := fs.ReadFile(dist, "index.html")
		if err == nil {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(b)
			return
		}
		http.NotFound(w, r)
	}))
	return r
}

func (s *Server) handleModelsList(w http.ResponseWriter, r *http.Request) {
	namespace := s.namespaceFromRequest(r)
	items, err := s.store.ListModels(r.Context(), namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, struct {
		Items []api.ModelSummary `json:"items"`
	}{Items: items})
}

func (s *Server) handleModelDetail(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	if namespace == "" {
		namespace = s.namespaceFromRequest(r)
	}
	name := chi.URLParam(r, "name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	payload, err := s.store.GetModel(r.Context(), namespace, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (s *Server) handleModelSources(w http.ResponseWriter, r *http.Request) {
	namespace := s.namespaceFromRequest(r)
	items, err := s.store.ListModelSources(r.Context(), namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, struct {
		Items []api.ModelSourceSummary `json:"items"`
	}{Items: items})
}

func (s *Server) handleModelSourceDetail(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	obj, err := s.store.GetModelSource(r.Context(), namespace, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"name": obj.Name, "namespace": obj.Namespace, "spec": obj.Spec})
}

func (s *Server) handleNamespaces(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.ListNamespaces(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, struct {
		Items []api.NamespaceInfo `json:"items"`
	}{Items: items})
}

func (s *Server) handleErrors(w http.ResponseWriter, r *http.Request) {
	namespace := s.namespaceFromRequest(r)
	items, err := s.store.ListErrors(r.Context(), namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, struct {
		Items []api.ErrorBanner `json:"items"`
	}{Items: items})
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	namespace := s.namespaceFromRequest(r)
	ch, err := s.store.Watch(r.Context(), namespace)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for {
		select {
		case <-r.Context().Done():
			return
		case payload, ok := <-ch:
			if !ok {
				return
			}
			b, err := json.Marshal(payload)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: %s\n", payload.Action)
			fmt.Fprintf(w, "data: %s\n\n", b)
			flusher.Flush()
		}
	}
}

func (s *Server) handleModelDelete(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	if err := s.store.DeleteModel(r.Context(), namespace, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleModelVersionDelete(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")
	if version == "" {
		writeError(w, http.StatusBadRequest, "version is required")
		return
	}
	if err := s.store.DeleteModelVersion(r.Context(), namespace, name, version); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type shareToggleRequest struct {
	Enabled bool `json:"enabled"`
}

func (s *Server) handleShareToggle(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	version := chi.URLParam(r, "version")
	var req shareToggleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if err := s.store.ToggleVersionShare(r.Context(), namespace, name, version, req.Enabled); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleResync(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	if err := s.store.TriggerResync(r.Context(), namespace, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type modelSourceCreateRequest struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Type      string            `json:"type"`
	SecretRef string            `json:"secretRef"`
	Config    map[string]string `json:"config"`
}

func (s *Server) handleModelSourceCreate(w http.ResponseWriter, r *http.Request) {
	var req modelSourceCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	ns := req.Namespace
	if ns == "" {
		ns = s.namespaceFromRequest(r)
	}
	spec := modelv1.ModelSourceSpec{Type: req.Type, SecretRef: req.SecretRef, Config: req.Config}
	if err := s.store.CreateModelSource(r.Context(), ns, req.Name, spec); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleModelSourceUpdate(w http.ResponseWriter, r *http.Request) {
	var req modelSourceCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	ns := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	spec := modelv1.ModelSourceSpec{Type: req.Type, SecretRef: req.SecretRef, Config: req.Config}
	if err := s.store.UpdateModelSource(r.Context(), ns, name, spec); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleModelSourceDelete(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	if err := s.store.DeleteModelSource(r.Context(), ns, name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type modelCreateUpdateRequest struct {
	Name        string              `json:"name"`
	Namespace   string              `json:"namespace"`
	SourceRef   string              `json:"sourceRef"`
	Description string              `json:"description"`
	Tags        []string            `json:"tags"`
	Versions    []modelVersionInput `json:"versions"`
}

type modelVersionInput struct {
	Name         string `json:"name"`
	Repo         string `json:"repo"`
	Revision     string `json:"revision"`
	Precision    string `json:"precision"`
	DesiredState string `json:"desiredState"`
	ShareEnabled bool   `json:"shareEnabled"`
}

func (s *Server) handleModelCreate(w http.ResponseWriter, r *http.Request) {
	var req modelCreateUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	ns := req.Namespace
	if ns == "" {
		ns = s.namespaceFromRequest(r)
	}
	spec := toModelSpec(req)
	if err := s.store.CreateModel(r.Context(), ns, req.Name, spec); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	payload, err := s.store.GetModel(r.Context(), ns, req.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, payload)
}

func (s *Server) handleModelUpdate(w http.ResponseWriter, r *http.Request) {
	var req modelCreateUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	ns := chi.URLParam(r, "namespace")
	if ns == "" {
		ns = s.namespaceFromRequest(r)
	}
	name := chi.URLParam(r, "name")
	spec := toModelSpec(req)
	if err := s.store.UpdateModel(r.Context(), ns, name, spec); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	payload, err := s.store.GetModel(r.Context(), ns, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func toModelSpec(req modelCreateUpdateRequest) modelv1.ModelSpec {
	versions := make([]modelv1.ModelVersion, 0, len(req.Versions))
	for _, v := range req.Versions {
		mv := modelv1.ModelVersion{
			Name:      v.Name,
			Repo:      v.Repo,
			Revision:  v.Revision,
			Precision: v.Precision,
		}
		if v.DesiredState != "" {
			mv.State = modelv1.ModelVersionState(v.DesiredState)
		}
		if v.ShareEnabled {
			mv.Share = &modelv1.ShareSpec{Enabled: true}
		}
		versions = append(versions, mv)
	}
	disp := &modelv1.DisplaySpec{Description: req.Description, Tags: req.Tags}
	return modelv1.ModelSpec{SourceRef: req.SourceRef, Display: disp, Versions: versions}
}

func (s *Server) handleSecretValidate(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("namespace")
	if ns == "" {
		ns = s.namespaceFromRequest(r)
	}
	name := r.URL.Query().Get("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	ready, msg, err := s.store.ValidateSecret(r.Context(), ns, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ready": ready, "message": msg})
}

func (s *Server) handleDatasets(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("namespace")
	if ns == "" {
		ns = s.namespaceFromRequest(r)
	}
	items, err := s.store.ListDatasets(r.Context(), ns)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

func (s *Server) namespaceFromRequest(r *http.Request) string {
	ns := r.URL.Query().Get("namespace")
	if ns != "" {
		return ns
	}
	return s.cfg.DefaultNamespace
}

func writeJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, map[string]string{"error": message})
}

func exists(fsys fs.FS, name string) bool {
	f, err := fsys.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func contentTypeByExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".html":
		return "text/html"
	case ".js":
		return "text/javascript"
	case ".css":
		return "text/css"
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".ico":
		return "image/x-icon"
	default:
		return "application/octet-stream"
	}
}
