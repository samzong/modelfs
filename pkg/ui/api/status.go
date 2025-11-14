package api

var phasePriority = map[Phase]int{
	PhaseFailed:     4,
	PhaseProcessing: 3,
	PhasePending:    2,
	PhaseReady:      1,
	PhaseUnknown:    0,
}

func AggregatePhase(phases []Phase) Phase {
	if len(phases) == 0 {
		return PhaseUnknown
	}
	worst := PhaseUnknown
	worstScore := phasePriority[PhaseUnknown]
	for _, p := range phases {
		score, ok := phasePriority[p]
		if !ok {
			p = PhaseUnknown
			score = phasePriority[p]
		}
		if score > worstScore {
			worst = p
			worstScore = score
		}
	}
	return worst
}
