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
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// BuildDatasetSpec builds a DatasetSpec from a ModelVersion and ModelSource.
// It reads the Secret referenced by ModelSource.secretRef and merges options (if secretRef is provided).
func BuildDatasetSpec(ctx context.Context, c client.Client, version modelv1.ModelVersion, source modelv1.ModelSource, namespace string) (*datasetv1alpha1.DatasetSpec, error) {
	// Convert source type
	datasetType, err := convertSourceType(source.Spec.Type)
	if err != nil {
		return nil, err
	}

	// Merge options: ModelSource config + Secret data (if secretRef is provided)
	options := make(map[string]string)
	for k, v := range source.Spec.Config {
		options[k] = v
	}
	// Add Secret data as options (if secretRef is provided)
	if source.Spec.SecretRef != "" {
		secret := &corev1.Secret{}
		secretKey := types.NamespacedName{Namespace: namespace, Name: source.Spec.SecretRef}
		if err := c.Get(ctx, secretKey, secret); err != nil {
			return nil, fmt.Errorf("get secret %s: %w", source.Spec.SecretRef, err)
		}
		for k, v := range secret.Data {
			options[k] = string(v)
		}
	}

	// Build URI from repo and revision
	uri, err := buildDatasetURI(source, version)
	if err != nil {
		return nil, fmt.Errorf("build URI: %w", err)
	}

	// Build DatasetSource
	datasetSource := datasetv1alpha1.DatasetSource{
		Type:    datasetType,
		URI:     uri,
		Options: options,
	}

	// Build VolumeClaimTemplate from ModelVolumeSpec
	var volumeClaimTemplate corev1.PersistentVolumeClaim
	if version.Storage != nil {
		volumeClaimTemplate = corev1.PersistentVolumeClaim{
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: version.Storage.AccessModes,
				Resources: corev1.VolumeResourceRequirements{
					Requests: version.Storage.Resources.Requests,
					Limits:   version.Storage.Resources.Limits,
				},
				StorageClassName: version.Storage.StorageClassName,
			},
		}
	} else {
		// Default: ReadWriteMany, 100Ti
		volumeClaimTemplate = corev1.PersistentVolumeClaim{
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteMany,
				},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: mustParseResourceQuantity("100Ti"),
					},
				},
			},
		}
	}

	return &datasetv1alpha1.DatasetSpec{
		Source:              datasetSource,
		VolumeClaimTemplate: volumeClaimTemplate,
	}, nil
}

// EnsureDataset creates or updates a Dataset CR.
func EnsureDataset(ctx context.Context, c client.Client, name, namespace string, spec *datasetv1alpha1.DatasetSpec, ownerRef *metav1.OwnerReference) error {
	dataset := &datasetv1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: *spec,
	}
	if ownerRef != nil {
		dataset.OwnerReferences = []metav1.OwnerReference{*ownerRef}
	}

	existing := &datasetv1alpha1.Dataset{}
	key := types.NamespacedName{Name: name, Namespace: namespace}
	err := c.Get(ctx, key, existing)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			// Create new dataset
			return c.Create(ctx, dataset)
		}
		return err
	}

	// Update existing dataset (patch spec)
	existing.Spec = *spec
	return c.Update(ctx, existing)
}

// EnsureReferenceDataset creates or updates a REFERENCE Dataset in the target namespace.
func EnsureReferenceDataset(ctx context.Context, c client.Client, sourceNs, sourceDatasetName, targetNs, targetName string, labels map[string]string) error {
	uri := fmt.Sprintf("dataset://%s/%s", sourceNs, sourceDatasetName)
	dataset := &datasetv1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      targetName,
			Namespace: targetNs,
			Labels:    labels,
		},
		Spec: datasetv1alpha1.DatasetSpec{
			Source: datasetv1alpha1.DatasetSource{
				Type: datasetv1alpha1.DatasetTypeReference,
				URI:  uri,
			},
		},
	}

	existing := &datasetv1alpha1.Dataset{}
	key := types.NamespacedName{Name: targetName, Namespace: targetNs}
	err := c.Get(ctx, key, existing)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			return c.Create(ctx, dataset)
		}
		return err
	}

	// Update existing reference dataset
	existing.Spec = dataset.Spec
	if existing.Labels == nil {
		existing.Labels = make(map[string]string)
	}
	for k, v := range labels {
		existing.Labels[k] = v
	}
	return c.Update(ctx, existing)
}

// GetDatasetName returns the Dataset name for a model version.
// Format: mdl-<model>-<version>
func GetDatasetName(modelName, versionName string) string {
	return fmt.Sprintf("mdl-%s-%s", modelName, versionName)
}

// GetReferenceDatasetName returns the REFERENCE Dataset name for sharing.
// Format: share-<source-ns>-<model>-<version>
func GetReferenceDatasetName(sourceNs, modelName, versionName string) string {
	return fmt.Sprintf("share-%s-%s-%s", sourceNs, modelName, versionName)
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

func buildDatasetURI(source modelv1.ModelSource, version modelv1.ModelVersion) (string, error) {
	repo := version.Repo
	revision := version.Revision
	if revision == "" {
		revision = "main"
	}

	// Extract URI from ModelSource config if provided
	if uri, ok := source.Spec.Config["uri"]; ok {
		return uri, nil
	}

	// Build URI based on type
	switch source.Spec.Type {
	case "HTTP", "S3", "GIT":
		if repo != "" {
			return repo, nil
		}
		if urlStr, ok := source.Spec.Config["url"]; ok {
			return urlStr, nil
		}
		return "", fmt.Errorf("missing repo in ModelVersion or url in ModelSource config for type %s", source.Spec.Type)
	case "HUGGING_FACE":
		if repo == "" {
			return "", fmt.Errorf("missing repo in ModelVersion for HuggingFace")
		}
		uri := fmt.Sprintf("huggingface://%s", repo)
		if revision != "main" {
			uri = fmt.Sprintf("%s@%s", uri, revision)
		}
		return uri, nil
	case "MODEL_SCOPE":
		if repo == "" {
			return "", fmt.Errorf("missing repo in ModelVersion for ModelScope")
		}
		uri := fmt.Sprintf("modelscope://%s", repo)
		if revision != "main" {
			uri = fmt.Sprintf("%s@%s", uri, revision)
		}
		return uri, nil
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

func mustParseResourceQuantity(s string) resource.Quantity {
	q, err := resource.ParseQuantity(s)
	if err != nil {
		panic(fmt.Sprintf("invalid quantity %s: %v", s, err))
	}
	return q
}
