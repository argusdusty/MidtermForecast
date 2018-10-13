package Predict

import (
	. "MidtermForecast/Utils"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"time"
)

var (
	// Pre-computed in historical_optimize.go
	INCUMBENT_ADVANTAGE_PVI = 0.016976345793067837 // Average margin advantage to incumbent over what PVI would expect
	FUNDRAISING_MULTIPLIER  = 0.02367143350492486  // Gain to generic ballot per doubling of fundraising over your opponent
	PAST_PVI_WEIGHT         = 0.25473575624092604  // Weight of previous PVI in PVI calculation
	OVERALL_PVI_WEIGHT      = 0.2392411256694573   // Weight of generic ballot average in PVI calculation
	FUNDAMENTALS_WEIGHT     = 22.90325456831519    // Multiplier to beta/dirichlet params for fundamentals margin in forecast. Larger means more confident
)

type Fundamental struct {
	StepName string  `json:"step_name"`
	Value    float64 `json:"value"`
	Weight   float64 `json:"weight"`
}

type Fundamentals []Fundamental

type RaceFundamentals map[string]Fundamentals

func (F Fundamentals) GetText(name string) []string {
	s := make([]string, 0, len(F)+1)
	s = append(s, name+":")
	for i, f := range F {
		if f.Weight == 0.0 {
			continue
		}
		if f.StepName == "Fundamentals Forecast" {
			s = append(s, fmt.Sprintf("Step %d. %s: %.2f%%", i+1, f.StepName, f.Value*100))
		} else {
			s = append(s, fmt.Sprintf("Step %d. %s: %.2f%% (Weight: %.1f%%)", i+1, f.StepName, f.Value*100, f.Weight*100))
		}
	}
	return s
}

func LoadFundamentals(ftype string, F *RaceFundamentals) (error, time.Time) {
	f, t := LoadFileCache("forecast/"+ftype+"_fundamentals.json", func(r io.Reader) interface{} {
		err := json.NewDecoder(r).Decode(F)
		if err != nil {
			panic(err)
		}
		return *F
	})
	*F = f.(RaceFundamentals)
	return nil, t
}

func SaveFundamentals(name string, fundamentals RaceFundamentals) {
	f, err := os.Create(name + ".tmp")
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(fundamentals); err != nil {
		panic(err)
	}
	f.Close()
	if err := os.Rename(name+".tmp", name); err != nil {
		panic(err)
	}
}

func getBetaFromPVI(pvi float64) (alpha, beta float64) {
	if pvi > 1 {
		return 1.0, 0.0
	} else if pvi < -1 {
		return 0.0, 1.0
	}
	return (1 + pvi) / 2, (1 - pvi) / 2
}

func GetBetasFromPVIs(pvis map[string]float64) map[string][2]float64 {
	r := make(map[string][2]float64, len(pvis))
	for k, pvi := range pvis {
		alpha, beta := getBetaFromPVI(pvi)
		r[k] = [2]float64{alpha, beta}
	}
	return r
}

func GetFundamentalsRatings(incumbents map[string]string, fundraising_ratios map[string]float64, pvi_estimates []map[string]float64, pvi_weights []float64, pvi_names []string, overall_pvi float64) (map[string][2]float64, RaceFundamentals) {
	steps := make(RaceFundamentals)
	pvi_est := make(map[string]float64) // Est (D-R)/(D+R)
	pvi_sw := make(map[string]float64)
	for i, pvi_estimate := range pvi_estimates {
		w := pvi_weights[i]
		for k, pvi := range pvi_estimate {
			pvi_est[k] += pvi * w
			pvi_sw[k] += w
		}
	}
	for i, name := range pvi_names {
		w := pvi_weights[i]
		for k, pvi := range pvi_estimates[i] {
			steps[k] = append(steps[k], Fundamental{StepName: name + " PVI Forecast", Value: pvi, Weight: w})
		}
	}
	ratings := make(map[string][2]float64, len(pvi_est))
	for k, pvi := range pvi_est {
		fr, ok := fundraising_ratios[k]
		if !ok {
			fr = 1.0
		}
		pvi += overall_pvi * OVERALL_PVI_WEIGHT
		steps[k] = append(steps[k], Fundamental{StepName: "Generic ballot adjustment", Value: overall_pvi, Weight: OVERALL_PVI_WEIGHT})
		pvi += math.Log(fr) / math.Ln2 * FUNDRAISING_MULTIPLIER
		steps[k] = append(steps[k], Fundamental{StepName: "Fundraising adjustment", Value: math.Log(fr) / math.Ln2 * FUNDRAISING_MULTIPLIER, Weight: 1.0})
		if incumbents[k] == "D" {
			pvi += INCUMBENT_ADVANTAGE_PVI
			steps[k] = append(steps[k], Fundamental{StepName: "Incumbent advantage adjustment", Value: INCUMBENT_ADVANTAGE_PVI, Weight: 1.0})
		} else if incumbents[k] == "R" {
			pvi -= INCUMBENT_ADVANTAGE_PVI
			steps[k] = append(steps[k], Fundamental{StepName: "Incumbent advantage adjustment", Value: -INCUMBENT_ADVANTAGE_PVI, Weight: 1.0})
		}
		steps[k] = append(steps[k], Fundamental{StepName: "Fundamentals Forecast", Value: pvi, Weight: 1.0})
		alpha, beta := getBetaFromPVI(pvi)
		ratings[k] = [2]float64{alpha * FUNDAMENTALS_WEIGHT, beta * FUNDAMENTALS_WEIGHT}
	}
	return ratings, steps
}
