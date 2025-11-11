package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ModelSource defines where model artifacts can be fetched.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=msrc
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
type ModelSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModelSourceSpec   `json:"spec,omitempty"`
	Status ModelSourceStatus `json:"status,omitempty"`
}

// ModelSourceSpec describes access details for a model source.
type ModelSourceSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=GIT;S3;HTTP;PVC;NFS;CONDA;REFERENCE;HUGGING_FACE;MODEL_SCOPE
	Type string `json:"type"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Config map[string]string `json:"config,omitempty"`
}

// ModelSourceStatus tracks availability of a source.
type ModelSourceStatus struct {
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// ModelSourceList is a list of sources.
type ModelSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ModelSource `json:"items"`
}
