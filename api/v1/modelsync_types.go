package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ModelSync orchestrates synchronization from a source into models.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=msync
// +kubebuilder:printcolumn:name="Model",type=string,JSONPath=`.spec.modelRef`
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`
type ModelSync struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModelSyncSpec   `json:"spec,omitempty"`
	Status ModelSyncStatus `json:"status,omitempty"`
}

// ModelSyncSpec defines the desired sync configuration.
type ModelSyncSpec struct {
	// +kubebuilder:validation:Required
	ModelRef string `json:"modelRef"`
	// +kubebuilder:validation:Required
	// Version specifies which version from the Model to sync. Must be one of the versions defined in Model.spec.versionConfigs.
	Version string `json:"version"`
	// +kubebuilder:validation:Optional
	Schedule string `json:"schedule,omitempty"`
	// +kubebuilder:validation:Optional
	RetentionCount int `json:"retentionCount,omitempty"`
}

// ModelSyncStatus captures the last sync outcome.
type ModelSyncStatus struct {
	// +kubebuilder:validation:Optional
	LastSyncedAt metav1.Time `json:"lastSyncedAt,omitempty"`
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// ModelSyncList lists ModelSync objects.
type ModelSyncList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ModelSync `json:"items"`
}
