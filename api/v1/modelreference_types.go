package v1

// ModelReference links consumers to a concrete model version.
type ModelReference struct {
	Metadata ObjectMeta           `json:"metadata"`
	Spec     ModelReferenceSpec   `json:"spec"`
	Status   ModelReferenceStatus `json:"status"`
}

// ModelReferenceSpec defines lookup details for a model reference.
type ModelReferenceSpec struct {
	ModelName string `json:"modelName"`
	Version   string `json:"version"`
	Alias     string `json:"alias"`
}

// ModelReferenceStatus reflects the resolved model revision.
type ModelReferenceStatus struct {
	ResolvedVersion string      `json:"resolvedVersion"`
	Conditions      []Condition `json:"conditions"`
}

// ModelReferenceList lists ModelReference items.
type ModelReferenceList struct {
	Items []ModelReference `json:"items"`
}
