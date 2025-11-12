package controllers

import "fmt"

// containsString checks if a string slice contains a specific string.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// removeString removes a string from a slice and returns the new slice.
func removeString(slice []string, s string) []string {
	var result []string
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}

// formatNamespacedName formats namespace and name into "namespace/name" format.
func formatNamespacedName(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

// buildModelLabels builds a label map for Model resources.
func buildModelLabels(namespace, modelName, versionName string) map[string]string {
	return map[string]string{
		"modelfs.samzong.dev/model":   formatNamespacedName(namespace, modelName),
		"modelfs.samzong.dev/version": versionName,
	}
}
