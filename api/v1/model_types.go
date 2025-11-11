package v1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Model describes a machine learning model instance tracked by the system.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=model
// +kubebuilder:printcolumn:name="Versions",type=integer,JSONPath=`.status.syncedVersions[*].name`
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
	// SourceRef references a ModelSource in the same namespace.
	SourceRef string `json:"sourceRef"`
	// +kubebuilder:validation:Optional
	// Display contains display metadata for catalog purposes.
	Display *DisplaySpec `json:"display,omitempty"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	// Versions defines all model versions and their configurations.
	Versions []ModelVersion `json:"versions"`
}

// DisplaySpec contains display metadata for a model.
type DisplaySpec struct {
	// +kubebuilder:validation:Optional
	Description string `json:"description,omitempty"`
	// +kubebuilder:validation:Optional
	Tags []string `json:"tags,omitempty"`
	// +kubebuilder:validation:Optional
	LogoURL string `json:"logoURL,omitempty"`
}

// ModelVersion defines a specific version of a model.
type ModelVersion struct {
	// +kubebuilder:validation:Required
	// Name is the unique version identifier (e.g., "fp16", "q4", "v2.0.0").
	Name string `json:"name"`
	// +kubebuilder:validation:Required
	// Repo specifies the repository/path for this version. Format depends on the source type.
	// For HUGGING_FACE: format is "username/model-name" (e.g., "qwen/Qwen3-7B")
	// For S3: format is "bucket/path"
	// For GIT: format is the repository URL
	Repo string `json:"repo"`
	// +kubebuilder:validation:Optional
	// Revision specifies the revision/branch/tag to use (default: "main").
	Revision string `json:"revision,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=FP16;INT4;INT8
	// Precision specifies the model precision.
	Precision string `json:"precision,omitempty"`
	// +kubebuilder:validation:Optional
	// Metadata contains version-specific metadata (e.g., tags).
	Metadata map[string]string `json:"metadata,omitempty"`
	// +kubebuilder:validation:Optional
	// Storage defines the PVC specification for this version.
	Storage *ModelVolumeSpec `json:"storage,omitempty"`
	// +kubebuilder:validation:Optional
	// State specifies whether this version should be present or absent (default: PRESENT).
	// +kubebuilder:default=PRESENT
	// +kubebuilder:validation:Enum=PRESENT;ABSENT
	State ModelVersionState `json:"state,omitempty"`
	// +kubebuilder:validation:Optional
	// Share defines sharing configuration for this version.
	Share *ShareSpec `json:"share,omitempty"`
}

// ModelVersionState represents the desired state of a model version.
// +kubebuilder:validation:Enum=PRESENT;ABSENT
type ModelVersionState string

const (
	// ModelVersionStatePresent indicates the version should be synced and available.
	ModelVersionStatePresent ModelVersionState = "PRESENT"
	// ModelVersionStateAbsent indicates the version should be removed.
	ModelVersionStateAbsent ModelVersionState = "ABSENT"
)

// ModelVolumeSpec defines the PVC specification for a model version.
type ModelVolumeSpec struct {
	// +kubebuilder:validation:Optional
	// AccessModes defines the access modes for the PVC (default: ReadWriteMany).
	AccessModes []corev1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`
	// +kubebuilder:validation:Required
	// Resources defines the storage resource requirements.
	Resources corev1.ResourceRequirements `json:"resources"`
	// +kubebuilder:validation:Optional
	// StorageClassName specifies the storage class to use.
	StorageClassName *string `json:"storageClassName,omitempty"`
}

// ShareSpec defines sharing configuration for a model version.
type ShareSpec struct {
	// +kubebuilder:validation:Required
	// Enabled indicates whether sharing is enabled for this version.
	Enabled bool `json:"enabled"`
	// +kubebuilder:validation:Optional
	// NamespaceSelector selects namespaces that can receive shared datasets.
	NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty"`
	// +kubebuilder:validation:Optional
	// RequireOptInLabel specifies a label key-value pair that namespaces must have to receive shares.
	// Format: "key=value" or just "key" (value defaults to "true").
	RequireOptInLabel string `json:"requireOptInLabel,omitempty"`
}

// ModelStatus captures observed details about a model.
type ModelStatus struct {
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// +kubebuilder:validation:Optional
	// SyncedVersions lists all versions and their observed states.
	SyncedVersions []SyncedVersion `json:"syncedVersions,omitempty"`
	// +kubebuilder:validation:Optional
	// ObservedGeneration tracks the generation of the Model spec that was last reconciled.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// SyncedVersion represents the observed state of a model version.
type SyncedVersion struct {
	// Name is the version name from spec.
	Name string `json:"name"`
	// Phase is the Dataset phase (Pending/Processing/Ready/Failed).
	Phase string `json:"phase,omitempty"`
	// PVCName is the name of the PVC created for this version.
	PVCName string `json:"pvcName,omitempty"`
	// ActiveDataset is the name of the currently active Dataset CR.
	ActiveDataset string `json:"activeDataset,omitempty"`
	// LastSyncTime is the last sync time from the Dataset status.
	LastSyncTime *metav1.Time `json:"lastSyncTime,omitempty"`
	// Conditions contains conditions from the Dataset status.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// ObservedState is the observed state (PRESENT/ABSENT).
	ObservedState ModelVersionState `json:"observedState,omitempty"`
	// ObservedStorage is the observed storage capacity.
	ObservedStorage *resource.Quantity `json:"observedStorage,omitempty"`
	// ObservedVersionHash is a hash of the version spec for change detection.
	ObservedVersionHash string `json:"observedVersionHash,omitempty"`
}

// +kubebuilder:object:root=true

// ModelList lists multiple models.
type ModelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Model `json:"items"`
}
