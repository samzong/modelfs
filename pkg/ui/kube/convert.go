package kube

import (
	modelv1 "github.com/samzong/modelfs/api/v1"
	"github.com/samzong/modelfs/pkg/ui/api"
	"time"
)

func summarizeModel(m *modelv1.Model) api.ModelSummary {
	total := len(m.Spec.Versions)
	ready := 0
	phases := make([]api.Phase, 0, len(m.Status.SyncedVersions))
	var last time.Time
	for _, sv := range m.Status.SyncedVersions {
		if sv.Phase == "Ready" || sv.Phase == "READY" {
			ready++
		}
		phases = append(phases, toPhase(sv.Phase))
		if sv.LastSyncTime != nil && sv.LastSyncTime.Time.After(last) {
			last = sv.LastSyncTime.Time
		}
	}
	var tags []string
	if m.Spec.Display != nil {
		tags = append(tags, m.Spec.Display.Tags...)
	}
	return api.ModelSummary{
		Name:          m.Name,
		Namespace:     m.Namespace,
		SourceRef:     m.Spec.SourceRef,
		Tags:          tags,
		VersionsReady: ready,
		VersionsTotal: total,
		LastSyncTime:  last,
		Status:        api.AggregatePhase(phases),
	}
}

func modelDetail(m *modelv1.Model) api.ModelDetail {
	summary := summarizeModel(m)
	versions := make([]api.ModelVersionView, 0, len(m.Spec.Versions))
	for _, v := range m.Spec.Versions {
		vv := api.ModelVersionView{
			Name:         v.Name,
			Repo:         v.Repo,
			Revision:     v.Revision,
			Precision:    v.Precision,
			DesiredState: string(v.State),
			ShareEnabled: v.Share != nil && v.Share.Enabled,
			DatasetPhase: api.PhaseUnknown,
		}
		for _, sv := range m.Status.SyncedVersions {
			if sv.Name == v.Name {
				vv.DatasetPhase = toPhase(sv.Phase)
				vv.PVCName = sv.PVCName
				vv.ObservedHash = sv.ObservedVersionHash
				break
			}
		}
		versions = append(versions, vv)
	}
	desc := ""
	if m.Spec.Display != nil {
		desc = m.Spec.Display.Description
	}
	return api.ModelDetail{Summary: summary, Description: desc, Versions: versions}
}

func toPhase(p string) api.Phase {
	switch p {
	case "READY", "Ready":
		return api.PhaseReady
	case "PENDING", "Pending":
		return api.PhasePending
	case "PROCESSING", "Processing":
		return api.PhaseProcessing
	case "FAILED", "Failed":
		return api.PhaseFailed
	default:
		return api.PhaseUnknown
	}
}
