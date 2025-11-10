package v1

// ModelSync orchestrates synchronization from a source into models.
type ModelSync struct {
	Metadata ObjectMeta      `json:"metadata"`
	Spec     ModelSyncSpec   `json:"spec"`
	Status   ModelSyncStatus `json:"status"`
}

// ModelSyncSpec defines the desired sync configuration.
type ModelSyncSpec struct {
	ModelRef       string `json:"modelRef"`
	SourceRef      string `json:"sourceRef"`
	Schedule       string `json:"schedule"`
	RetentionCount int    `json:"retentionCount"`
}

// ModelSyncStatus captures the last sync outcome.
type ModelSyncStatus struct {
	LastSyncedAt string      `json:"lastSyncedAt"`
	Conditions   []Condition `json:"conditions"`
}

// ModelSyncList lists ModelSync objects.
type ModelSyncList struct {
	Items []ModelSync `json:"items"`
}
