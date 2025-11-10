package v1

// ModelSource defines where model artifacts can be fetched.
type ModelSource struct {
	Metadata ObjectMeta        `json:"metadata"`
	Spec     ModelSourceSpec   `json:"spec"`
	Status   ModelSourceStatus `json:"status"`
}

// ModelSourceSpec describes access details for a model source.
type ModelSourceSpec struct {
	Type   string            `json:"type"`
	Config map[string]string `json:"config"`
}

// ModelSourceStatus tracks availability of a source.
type ModelSourceStatus struct {
	Conditions []Condition `json:"conditions"`
}

// ModelSourceList is a list of sources.
type ModelSourceList struct {
	Items []ModelSource `json:"items"`
}
