package v1

// Model describes a machine learning model instance tracked by the system.
type Model struct {
	Metadata ObjectMeta  `json:"metadata"`
	Spec     ModelSpec   `json:"spec"`
	Status   ModelStatus `json:"status"`
}

// ModelSpec contains desired attributes for a model.
type ModelSpec struct {
	Version   string `json:"version"`
	SourceRef string `json:"sourceRef"`
}

// ModelStatus captures observed details about a model.
type ModelStatus struct {
	Conditions []Condition `json:"conditions"`
}

// ModelList lists multiple models.
type ModelList struct {
	Items []Model `json:"items"`
}
