package planner

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/cartomix/cancun/gen/go/common"
	eng "github.com/cartomix/cancun/gen/go/engine"
)

// Options controls how set planning scores transitions.
type Options struct {
	Mode           eng.SetMode
	AllowKeyJumps  bool
	MaxBpmStep     float64
	MustPlayHashes map[string]bool
	BanHashes      map[string]bool
}

// Plan produces an ordering of tracks with per-edge explanations.
func Plan(analyses []*common.TrackAnalysis, opts Options) ([]*common.TrackId, []*common.EdgeExplanation, error) {
	if len(analyses) == 0 {
		return nil, nil, fmt.Errorf("no analyses provided")
	}

	filtered := make([]*common.TrackAnalysis, 0, len(analyses))
	for _, a := range analyses {
		if a == nil || a.GetId() == nil {
			continue
		}
		if opts.BanHashes != nil && opts.BanHashes[a.GetId().GetContentHash()] {
			continue
		}
		filtered = append(filtered, a)
	}

	if len(filtered) == 0 {
		return nil, nil, fmt.Errorf("all tracks were filtered out")
	}

	if len(opts.MustPlayHashes) > 0 {
		for hash := range opts.MustPlayHashes {
			found := false
			for _, a := range filtered {
				if a.GetId().GetContentHash() == hash {
					found = true
					break
				}
			}
			if !found {
				return nil, nil, fmt.Errorf("must-play track %s missing analysis", hash)
			}
		}
	}

	start := chooseStart(filtered, opts.Mode)
	order := []*common.TrackAnalysis{start}
	remaining := map[string]*common.TrackAnalysis{}
	for _, a := range filtered {
		if a.GetId().GetContentHash() == start.GetId().GetContentHash() {
			continue
		}
		remaining[a.GetId().GetContentHash()] = a
	}

	explanations := []*common.EdgeExplanation{}
	current := start

	for len(remaining) > 0 {
		next, explanation := bestNext(current, remaining, opts)
		if next == nil {
			// fall back to arbitrary ordering to keep the plan complete
			for _, leftover := range remaining {
				order = append(order, leftover)
			}
			break
		}
		order = append(order, next)
		explanations = append(explanations, explanation)
		delete(remaining, next.GetId().GetContentHash())
		current = next
	}

	ids := make([]*common.TrackId, 0, len(order))
	for _, t := range order {
		ids = append(ids, t.GetId())
	}

	return ids, explanations, nil
}

func chooseStart(analyses []*common.TrackAnalysis, mode eng.SetMode) *common.TrackAnalysis {
	clone := make([]*common.TrackAnalysis, len(analyses))
	copy(clone, analyses)

	switch mode {
	case eng.SetMode_WARM_UP:
		sort.Slice(clone, func(i, j int) bool {
			return clone[i].GetEnergyGlobal() < clone[j].GetEnergyGlobal()
		})
	case eng.SetMode_PEAK_TIME:
		sort.Slice(clone, func(i, j int) bool {
			return clone[i].GetEnergyGlobal() > clone[j].GetEnergyGlobal()
		})
	default: // OPEN_FORMAT or unspecified
		sort.Slice(clone, func(i, j int) bool {
			return estimateBPM(clone[i]) < estimateBPM(clone[j])
		})
	}

	return clone[0]
}

func bestNext(current *common.TrackAnalysis, remaining map[string]*common.TrackAnalysis, opts Options) (*common.TrackAnalysis, *common.EdgeExplanation) {
	var (
		bestTrack *common.TrackAnalysis
		bestScore = math.Inf(-1)
		bestEdge  *common.EdgeExplanation
	)

	for _, cand := range remaining {
		score, expl := scoreEdge(current, cand, opts)
		if score > bestScore {
			bestScore = score
			bestTrack = cand
			bestEdge = expl
		}
	}

	return bestTrack, bestEdge
}

func scoreEdge(from, to *common.TrackAnalysis, opts Options) (float64, *common.EdgeExplanation) {
	fromBPM := estimateBPM(from)
	toBPM := estimateBPM(to)
	bpmDelta := toBPM - fromBPM

	tempoScore := 4.0 - math.Abs(bpmDelta)/2
	if opts.MaxBpmStep > 0 && math.Abs(bpmDelta) > opts.MaxBpmStep {
		tempoScore -= 4 // heavy penalty for exceeding allowed step
	}

	keyScore, relation := keyCompatibility(from.GetKey().GetValue(), to.GetKey().GetValue(), opts.AllowKeyJumps)

	energyDelta := int(to.GetEnergyGlobal() - from.GetEnergyGlobal())
	energyScore := 2.0 - math.Abs(float64(energyDelta))*0.5

	switch opts.Mode {
	case eng.SetMode_WARM_UP:
		if energyDelta > 0 {
			energyScore += 1
		}
	case eng.SetMode_PEAK_TIME:
		if to.GetEnergyGlobal() >= from.GetEnergyGlobal() {
			energyScore += 1
		}
	}

	window := windowOverlap(from, to)
	windowScore := 0.0
	if window != "" {
		windowScore = 1.0
	}

	total := keyScore + tempoScore + energyScore + windowScore

	expl := &common.EdgeExplanation{
		From:          from.GetId(),
		To:            to.GetId(),
		Score:         float32(total),
		TempoDelta:    float32(bpmDelta),
		EnergyDelta:   int32(energyDelta),
		KeyRelation:   relation,
		WindowOverlap: window,
		Reason:        fmt.Sprintf("%s; Δ%.1f BPM; Δenergy %d", relation, bpmDelta, energyDelta),
	}

	return total, expl
}

func keyCompatibility(from, to string, allowJumps bool) (float64, string) {
	if from == "" || to == "" {
		return -1, "unknown key"
	}

	fromNum, fromMode, okFrom := parseCamelot(from)
	toNum, toMode, okTo := parseCamelot(to)

	if !okFrom || !okTo {
		if allowJumps {
			return 0, "unverified key jump"
		}
		return -3, "key mismatch"
	}

	if fromNum == toNum && fromMode == toMode {
		return 4, "same key"
	}

	if fromMode == toMode && int(math.Abs(float64(fromNum-toNum))) == 1 {
		dir := "+"
		if toNum < fromNum {
			dir = "-"
		}
		return 3, fmt.Sprintf("%s1 Camelot", dir)
	}

	if allowJumps {
		return 1, "permitted key jump"
	}

	return -2, "distant key"
}

func parseCamelot(value string) (int, string, bool) {
	value = strings.TrimSpace(strings.ToUpper(value))
	if value == "" {
		return 0, "", false
	}

	mode := value[len(value)-1:]
	numPart := value[:len(value)-1]
	num, err := strconv.Atoi(numPart)
	if err != nil || num < 1 || num > 12 {
		return 0, "", false
	}

	if mode != "A" && mode != "B" {
		return 0, "", false
	}

	return num, mode, true
}

func windowOverlap(from, to *common.TrackAnalysis) string {
	if len(from.GetTransitionWindows()) == 0 || len(to.GetTransitionWindows()) == 0 {
		return ""
	}

	return fmt.Sprintf("%s → %s", from.GetTransitionWindows()[0].GetTag(), to.GetTransitionWindows()[0].GetTag())
}

func estimateBPM(a *common.TrackAnalysis) float64 {
	if a.GetBeatgrid() == nil {
		return 0
	}
	if tm := a.GetBeatgrid().GetTempoMap(); len(tm) > 0 {
		return tm[0].GetBpm()
	}
	beats := a.GetBeatgrid().GetBeats()
	if len(beats) >= 2 {
		delta := beats[1].GetTime().AsDuration() - beats[0].GetTime().AsDuration()
		if delta > 0 {
			return 60 / delta.Seconds()
		}
	}
	return 0
}
