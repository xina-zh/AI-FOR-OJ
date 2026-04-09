package service

type VerdictDistribution struct {
	ACCount          int `json:"ac_count"`
	WACount          int `json:"wa_count"`
	CECount          int `json:"ce_count"`
	RECount          int `json:"re_count"`
	TLECount         int `json:"tle_count"`
	UnjudgeableCount int `json:"unjudgeable_count"`
	UnknownCount     int `json:"unknown_count"`
}

func BuildVerdictDistribution(verdicts []string) VerdictDistribution {
	var dist VerdictDistribution
	for _, verdict := range verdicts {
		dist.Add(verdict)
	}
	return dist
}

func (d *VerdictDistribution) Add(verdict string) {
	switch verdict {
	case "AC":
		d.ACCount++
	case "WA":
		d.WACount++
	case "CE":
		d.CECount++
	case "RE":
		d.RECount++
	case "TLE":
		d.TLECount++
	case "UNJUDGEABLE":
		d.UnjudgeableCount++
	default:
		d.UnknownCount++
	}
}

func DiffVerdictDistribution(candidate, baseline VerdictDistribution) VerdictDistribution {
	return VerdictDistribution{
		ACCount:          candidate.ACCount - baseline.ACCount,
		WACount:          candidate.WACount - baseline.WACount,
		CECount:          candidate.CECount - baseline.CECount,
		RECount:          candidate.RECount - baseline.RECount,
		TLECount:         candidate.TLECount - baseline.TLECount,
		UnjudgeableCount: candidate.UnjudgeableCount - baseline.UnjudgeableCount,
		UnknownCount:     candidate.UnknownCount - baseline.UnknownCount,
	}
}
