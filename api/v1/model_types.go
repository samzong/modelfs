package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Model describes a machine learning model instance tracked by the system.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=model
// +kubebuilder:printcolumn:name="Versions",type=string,JSONPath=`.spec.versionConfigs`
// +kubebuilder:printcolumn:name="Source",type=string,JSONPath=`.spec.sourceRef`
type Model struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModelSpec   `json:"spec,omitempty"`
	Status ModelStatus `json:"status,omitempty"`
}

// ModelSpec contains desired attributes for a model.
type ModelSpec struct {
	// +kubebuilder:validation:Required
	SourceRef string `json:"sourceRef"`
	// +kubebuilder:validation:Required
	// VersionConfigs defines all versions and their configurations.
	// The keys of this map are the version names (e.g., "v2.0.0", "v2.5.0").
	// Each version must have its own repo and repoConfig.
	VersionConfigs map[string]VersionConfig `json:"versionConfigs"`
}

// VersionConfig provides configuration for a specific version.
type VersionConfig struct {
	// +kubebuilder:validation:Required
	// Repo specifies the repository/path for this version. Format depends on the source type.
	// For HUGGING_FACE: format is "username/model-name" (e.g., "qwen/Qwen2.5-7B-Instruct")
	// For S3: format is "bucket/path"
	// For GIT: format is the repository URL
	Repo string `json:"repo"`
	// +kubebuilder:validation:Optional
	// RepoConfig provides repository-specific configuration for this version.
	// Common configs like include/exclude should be in ModelSource config.
	// Only repo-specific configs (e.g., revision) should be here.
	RepoConfig map[string]string `json:"repoConfig,omitempty"`
}

// ModelStatus captures observed details about a model.
type ModelStatus struct {
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// +kubebuilder:validation:Optional
	// SyncedVersions lists all versions that have active ModelSync instances.
	SyncedVersions []SyncedVersion `json:"syncedVersions,omitempty"`
}

// SyncedVersion represents a version that is being synced.
type SyncedVersion struct {
	// Version is the version string.
	Version string `json:"version"`
	// ModelSyncName is the name of the ModelSync that syncs this version.
	ModelSyncName string `json:"modelSyncName"`
	// LastSyncedAt is the last sync time from the ModelSync status.
	LastSyncedAt metav1.Time `json:"lastSyncedAt,omitempty"`
	// Ready indicates whether this version is ready (synced successfully).
	Ready bool `json:"ready,omitempty"`
}

// +kubebuilder:object:root=true

// ModelList lists multiple models.
type ModelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Model `json:"items"`
}
