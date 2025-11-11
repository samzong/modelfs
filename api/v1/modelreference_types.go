package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ModelReference links consumers to a concrete model version.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=mref
// +kubebuilder:printcolumn:name="Model",type=string,JSONPath=`.spec.modelName`
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.spec.version`
type ModelReference struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModelReferenceSpec   `json:"spec,omitempty"`
	Status ModelReferenceStatus `json:"status,omitempty"`
}

// ModelReferenceSpec defines lookup details for a model reference.
type ModelReferenceSpec struct {
	// +kubebuilder:validation:Required
	ModelName string `json:"modelName"`
	// +kubebuilder:validation:Required
	Version string `json:"version"`
	// +kubebuilder:validation:Optional
	Alias string `json:"alias,omitempty"`
}

// ModelReferenceStatus reflects the resolved model revision.
type ModelReferenceStatus struct {
	// +kubebuilder:validation:Optional
	ResolvedVersion string `json:"resolvedVersion,omitempty"`
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// ModelReferenceList lists ModelReference items.
type ModelReferenceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ModelReference `json:"items"`
}
