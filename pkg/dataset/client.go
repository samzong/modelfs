package dataset

import (
	"context"
	"fmt"
	"net/url"

	datasetv1alpha1 "github.com/BaizeAI/dataset/api/dataset/v1alpha1"
	modelv1 "github.com/samzong/modelfs/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Client exposes the minimal surface required by controllers to integrate with BaizeAI/dataset.
type Client interface {
	EnsureDataset(ctx context.Context, model modelv1.Model, source modelv1.ModelSource, sync modelv1.ModelSync) error
	GetDatasetName(model modelv1.Model, sync modelv1.ModelSync) string
}

// KubernetesClient talks to dataset CRD via Kubernetes API.
type KubernetesClient struct {
	client client.Client
}

// NewKubernetesClient creates a Kubernetes-backed dataset client.
func NewKubernetesClient(c client.Client) *KubernetesClient {
	return &KubernetesClient{client: c}
}

// GetDatasetName returns the name of the Dataset CR for a given Model and ModelSync.
// Uses the version specified in ModelSync.
func (c *KubernetesClient) GetDatasetName(model modelv1.Model, sync modelv1.ModelSync) string {
	return fmt.Sprintf("%s-%s", model.Name, sync.Spec.Version)
}

// EnsureDataset creates or updates a Dataset CR that corresponds to the ModelSync.
func (c *KubernetesClient) EnsureDataset(ctx context.Context, model modelv1.Model, source modelv1.ModelSource, sync modelv1.ModelSync) error {
	version := sync.Spec.Version

	// Convert ModelSource to DatasetSource (with version-specific config)
	datasetSource, err := convertModelSourceToDatasetSource(source, model, version)
	if err != nil {
		return fmt.Errorf("convert model source: %w", err)
	}

	// Build URI from Model repo (version-specific or default) and ModelSource config
	uri, err := buildDatasetURI(source, model, version)
	if err != nil {
		return fmt.Errorf("build dataset URI: %w", err)
	}

	datasetSource.URI = uri

	// Create or update Dataset CR
	gvk := modelv1.GroupVersion.WithKind("ModelSync")
	dataset := &datasetv1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", model.Name, version),
			Namespace: sync.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(&sync, gvk),
			},
		},
		Spec: datasetv1alpha1.DatasetSpec{
			Source: *datasetSource,
			VolumeClaimTemplate: corev1.PersistentVolumeClaim{
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteMany,
					},
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("100Ti"),
						},
					},
				},
			},
		},
	}

	existing := &datasetv1alpha1.Dataset{}
	err = c.client.Get(ctx, client.ObjectKeyFromObject(dataset), existing)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			// Create new dataset
			return c.client.Create(ctx, dataset)
		}
		return err
	}

	// Update existing dataset
	existing.Spec = dataset.Spec
	return c.client.Update(ctx, existing)
}

func convertModelSourceToDatasetSource(source modelv1.ModelSource, model modelv1.Model, version string) (*datasetv1alpha1.DatasetSource, error) {
	datasetType, err := convertSourceType(source.Spec.Type)
	if err != nil {
		return nil, err
	}

	// Get version-specific config (must exist)
	versionConfig, exists := model.Spec.VersionConfigs[version]
	if !exists {
		return nil, fmt.Errorf("version %s not found in model versionConfigs", version)
	}

	// Merge ModelSource config (connection/auth info + common config like include/exclude) with Model repoConfig
	options := make(map[string]string)

	// First, copy ModelSource config (connection/auth info + common config)
	for k, v := range source.Spec.Config {
		options[k] = v
	}

	// Then, merge Model versionConfig.repoConfig (repo-specific config like revision)
	// Note: Model.repoConfig can override ModelSource config if needed
	if versionConfig.RepoConfig != nil {
		for k, v := range versionConfig.RepoConfig {
			options[k] = v
		}
	}

	return &datasetv1alpha1.DatasetSource{
		Type:    datasetType,
		Options: options,
	}, nil
}

func convertSourceType(modelType string) (datasetv1alpha1.DatasetType, error) {
	switch modelType {
	case "GIT":
		return datasetv1alpha1.DatasetTypeGit, nil
	case "S3":
		return datasetv1alpha1.DatasetTypeS3, nil
	case "HTTP":
		return datasetv1alpha1.DatasetTypeHTTP, nil
	case "PVC":
		return datasetv1alpha1.DatasetTypePVC, nil
	case "NFS":
		return datasetv1alpha1.DatasetTypeNFS, nil
	case "CONDA":
		return datasetv1alpha1.DatasetTypeConda, nil
	case "REFERENCE":
		return datasetv1alpha1.DatasetTypeReference, nil
	case "HUGGING_FACE":
		return datasetv1alpha1.DatasetTypeHuggingFace, nil
	case "MODEL_SCOPE":
		return datasetv1alpha1.DatasetTypeModelScope, nil
	default:
		return "", fmt.Errorf("unsupported source type: %s", modelType)
	}
}

func buildDatasetURI(source modelv1.ModelSource, model modelv1.Model, version string) (string, error) {
	// Get version-specific config (must exist)
	versionConfig, exists := model.Spec.VersionConfigs[version]
	if !exists {
		return "", fmt.Errorf("version %s not found in model versionConfigs", version)
	}

	repo := versionConfig.Repo
	repoConfig := versionConfig.RepoConfig

	// Extract URI from version-specific repoConfig if provided
	if repoConfig != nil {
		if uri, ok := repoConfig["uri"]; ok {
			return uri, nil
		}
	}

	// Extract URI from ModelSource config if provided
	if uri, ok := source.Spec.Config["uri"]; ok {
		return uri, nil
	}

	// Build URI based on type, using version-specific repo
	switch source.Spec.Type {
	case "HTTP", "S3", "GIT":
		// For these types, repo can be the URL
		if repo != "" {
			return repo, nil
		}
		// Fallback to config url
		if urlStr, ok := source.Spec.Config["url"]; ok {
			return urlStr, nil
		}
		return "", fmt.Errorf("missing repo in Model versionConfig[%s] or url in ModelSource config for type %s", version, source.Spec.Type)
	case "HUGGING_FACE":
		if repo == "" {
			return "", fmt.Errorf("missing repo in Model versionConfig[%s] for HuggingFace", version)
		}
		return fmt.Sprintf("huggingface://%s", repo), nil
	case "MODEL_SCOPE":
		if repo == "" {
			return "", fmt.Errorf("missing repo in Model versionConfig[%s] for ModelScope", version)
		}
		// ModelScope repo format: namespace/model
		return fmt.Sprintf("modelscope://%s", repo), nil
	case "PVC":
		if pvcName, ok := source.Spec.Config["pvcName"]; ok {
			path := repo
			if path == "" {
				path = source.Spec.Config["path"]
			}
			if path == "" {
				path = "/"
			}
			return fmt.Sprintf("pvc://%s%s", pvcName, path), nil
		}
		return "", fmt.Errorf("missing pvcName in ModelSource config for PVC")
	case "NFS":
		if server, ok := source.Spec.Config["server"]; ok {
			path := repo
			if path == "" {
				path = source.Spec.Config["path"]
			}
			if path == "" {
				path = "/"
			}
			u, err := url.Parse(fmt.Sprintf("nfs://%s%s", server, path))
			if err != nil {
				return "", err
			}
			return u.String(), nil
		}
		return "", fmt.Errorf("missing server in ModelSource config for NFS")
	default:
		return "", fmt.Errorf("cannot build URI for type %s", source.Spec.Type)
	}
}
