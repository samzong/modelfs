package api

import "time"

var sampleNS = []NamespaceInfo{{Name: "model-system"}, {Name: "prod"}, {Name: "staging"}}

var sampleModels = []ModelSummary{
	{
		Name:          "qwen3-7b",
		Namespace:     "model-system",
		SourceRef:     "hf-qwen",
		Tags:          []string{"llm", "qwen"},
		VersionsReady: 1,
		VersionsTotal: 2,
		LastSyncTime:  time.Now(),
		Status:        PhaseReady,
	},
	{
		Name:             "llama3-8b",
		Namespace:        "model-system",
		SourceRef:        "hf-llama3",
		Tags:             []string{"llm", "meta"},
		VersionsReady:    0,
		VersionsTotal:    1,
		LastSyncTime:     time.Now(),
		Status:           PhaseProcessing,
		ReconcileMessage: "Syncing weights",
	},
}

var sampleModelDetails = map[string]ModelDetail{
	"model-system/qwen3-7b": {
		Summary:     sampleModels[0],
		Description: "Qwen3 7B base model",
		Versions: []ModelVersionView{
			{Name: "fp16", Repo: "qwen/Qwen3-7B", DesiredState: "PRESENT", ShareEnabled: true, DatasetPhase: PhaseReady, PVCName: "mdl-qwen3-7b-fp16"},
			{Name: "int4", Repo: "qwen/Qwen3-7B", DesiredState: "PRESENT", ShareEnabled: false, DatasetPhase: PhasePending},
		},
	},
}

var sampleSources = []ModelSourceSummary{
	{
		Name:              "hf-qwen",
		Namespace:         "model-system",
		Type:              "HUGGING_FACE",
		SecretRef:         "hf-token",
		CredentialsReady:  true,
		CredentialsStatus: "OK",
		ReferencedModels:  []string{"model-system/qwen3-7b"},
		LastChecked:       time.Now(),
	},
}

func SampleNamespaces() []NamespaceInfo { return append([]NamespaceInfo{}, sampleNS...) }
func SampleModels(ns string) []ModelSummary {
	out := []ModelSummary{}
	for _, m := range sampleModels {
		if m.Namespace == ns {
			out = append(out, m)
		}
	}
	return out
}
func SampleModelDetail(ns, name string) (ModelDetail, bool) {
	d, ok := sampleModelDetails[ns+"/"+name]
	return d, ok
}
func SampleModelSources(ns string) []ModelSourceSummary {
	out := []ModelSourceSummary{}
	for _, s := range sampleSources {
		if s.Namespace == ns {
			out = append(out, s)
		}
	}
	return out
}
