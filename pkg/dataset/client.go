package dataset

import (
        "bytes"
        "context"
        "encoding/json"
        "errors"
        "fmt"
        "net/http"
        "net/url"
        "strings"
        "time"

        modelv1 "github.com/samzong/modelfs/api/v1"
)

// Client exposes the minimal surface required by controllers to integrate with BaizeAI/dataset.
type Client interface {
        EnsureSource(ctx context.Context, src modelv1.ModelSource) error
        TriggerSync(ctx context.Context, input SyncInput) error
}

// SyncInput captures the data the dataset API expects for a synchronization request.
type SyncInput struct {
        Namespace      string            `json:"namespace"`
        ModelName      string            `json:"modelName"`
        ModelVersion   string            `json:"modelVersion"`
        SourceName     string            `json:"sourceName"`
        SourceType     string            `json:"sourceType"`
        SourceConfig   map[string]string `json:"sourceConfig"`
        Schedule       string            `json:"schedule,omitempty"`
        RetentionCount int               `json:"retentionCount,omitempty"`
}

// HTTPClient talks to a dataset control-plane endpoint over HTTP.
type HTTPClient struct {
        baseURL    *url.URL
        httpClient *http.Client
}

// Option customises HTTPClient configuration.
type Option func(*HTTPClient)

// WithHTTPClient overrides the default HTTP client implementation.
func WithHTTPClient(client *http.Client) Option {
        return func(c *HTTPClient) {
                if client != nil {
                        c.httpClient = client
                }
        }
}

// NewHTTPClient creates an HTTP-backed dataset client.
func NewHTTPClient(endpoint string, opts ...Option) (*HTTPClient, error) {
        if endpoint == "" {
                return nil, errors.New("dataset endpoint must be provided")
        }
        parsed, err := url.Parse(endpoint)
        if err != nil {
                return nil, fmt.Errorf("invalid dataset endpoint: %w", err)
        }
        client := &HTTPClient{
                baseURL: parsed,
                httpClient: &http.Client{
                        Timeout: 30 * time.Second,
                },
        }
        for _, opt := range opts {
                opt(client)
        }
        return client, nil
}

// EnsureSource registers or updates a source definition inside dataset.
func (c *HTTPClient) EnsureSource(ctx context.Context, src modelv1.ModelSource) error {
        if err := validateSource(src); err != nil {
                return err
        }
        payload := map[string]any{
                "name":      src.Metadata.Name,
                "namespace": src.Metadata.Namespace,
                "type":      src.Spec.Type,
                "config":    src.Spec.Config,
        }
        return c.post(ctx, "/api/v1/sources", payload)
}

// TriggerSync instructs dataset to synchronise a model revision.
func (c *HTTPClient) TriggerSync(ctx context.Context, input SyncInput) error {
        if err := validateSync(input); err != nil {
                return err
        }
        return c.post(ctx, "/api/v1/syncs", input)
}

func (c *HTTPClient) post(ctx context.Context, p string, payload any) error {
        target := *c.baseURL
        target.Path = strings.TrimSuffix(c.baseURL.Path, "/") + p
        body, err := json.Marshal(payload)
        if err != nil {
                return fmt.Errorf("marshal payload: %w", err)
        }
        req, err := http.NewRequestWithContext(ctx, http.MethodPost, target.String(), bytes.NewReader(body))
        if err != nil {
                return fmt.Errorf("construct request: %w", err)
        }
        req.Header.Set("Content-Type", "application/json")
        resp, err := c.httpClient.Do(req)
        if err != nil {
                return fmt.Errorf("post to dataset: %w", err)
        }
        defer resp.Body.Close()
        if resp.StatusCode >= 300 {
                return fmt.Errorf("dataset returned status %d", resp.StatusCode)
        }
        return nil
}

func validateSource(src modelv1.ModelSource) error {
        if src.Metadata.Name == "" {
                return errors.New("model source name must be set")
        }
        if src.Metadata.Namespace == "" {
                return errors.New("model source namespace must be set")
        }
        if src.Spec.Type == "" {
                return errors.New("model source type must be set")
        }
        return nil
}

func validateSync(input SyncInput) error {
        if input.Namespace == "" {
                return errors.New("sync namespace must be set")
        }
        if input.ModelName == "" {
                return errors.New("sync model name must be set")
        }
        if input.ModelVersion == "" {
                return errors.New("sync model version must be set")
        }
        if input.SourceName == "" {
                return errors.New("sync source name must be set")
        }
        if input.SourceType == "" {
                return errors.New("sync source type must be set")
        }
        return nil
}
