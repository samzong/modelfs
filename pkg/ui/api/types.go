package api

import "time"

type Phase string

const (
	PhaseUnknown    Phase = "UNKNOWN"
	PhaseReady      Phase = "READY"
	PhasePending    Phase = "PENDING"
	PhaseProcessing Phase = "PROCESSING"
	PhaseFailed     Phase = "FAILED"
)

type ModelSummary struct {
	Name             string    `json:"name"`
	Namespace        string    `json:"namespace"`
	SourceRef        string    `json:"sourceRef"`
	Tags             []string  `json:"tags,omitempty"`
	VersionsReady    int       `json:"versionsReady"`
	VersionsTotal    int       `json:"versionsTotal"`
	LastSyncTime     time.Time `json:"lastSyncTime"`
	Status           Phase     `json:"status"`
	ReconcileMessage string    `json:"reconcileMessage,omitempty"`
}

type ModelVersionView struct {
	Name            string `json:"name"`
	Repo            string `json:"repo"`
	Revision        string `json:"revision,omitempty"`
	Precision       string `json:"precision,omitempty"`
	DesiredState    string `json:"desiredState"`
	ShareEnabled    bool   `json:"shareEnabled"`
	NamespacePolicy string `json:"namespacePolicy,omitempty"`
	DatasetPhase    Phase  `json:"datasetPhase"`
	PVCName         string `json:"pvcName,omitempty"`
	ObservedHash    string `json:"observedHash,omitempty"`
}

type ModelDetail struct {
	Summary        ModelSummary       `json:"summary"`
	Description    string             `json:"description,omitempty"`
	LogoURL        string             `json:"logoURL,omitempty"`
	Versions       []ModelVersionView `json:"versions"`
	ShareTargets   []string           `json:"shareTargets,omitempty"`
	ConditionsJSON string             `json:"conditionsJson,omitempty"`
}

type ModelSourceSummary struct {
	Name              string    `json:"name"`
	Namespace         string    `json:"namespace"`
	Type              string    `json:"type"`
	SecretRef         string    `json:"secretRef,omitempty"`
	CredentialsReady  bool      `json:"credentialsReady"`
	CredentialsStatus string    `json:"credentialsStatus,omitempty"`
	ReferencedModels  []string  `json:"referencedModels,omitempty"`
	LastChecked       time.Time `json:"lastChecked"`
}

type NamespaceInfo struct {
	Name string `json:"name"`
}

type ErrorBanner struct {
	Namespace string    `json:"namespace"`
	Message   string    `json:"message"`
	Reason    string    `json:"reason"`
	RetryAt   time.Time `json:"retryAt"`
}

type SSEPayload struct {
	Resource string      `json:"resource"`
	Action   string      `json:"action"`
	Payload  interface{} `json:"payload"`
}
