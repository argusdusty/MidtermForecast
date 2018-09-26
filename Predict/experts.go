package Predict

import (
	. "MidtermForecast/Utils"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"time"
)

type Expert struct {
	Expert              string             `json:"expert"`
	ConcentrationParams map[string]float64 `json:"concentration_params"`
	Weight              float64            `json:"weight"`
}

func (A Expert) Compare(B Expert) bool {
	if A.Weight > B.Weight {
		return true
	} else if A.Weight < B.Weight {
		return false
	}
	return A.Expert > B.Expert
}

type Experts []Expert

func (E Experts) Len() int           { return len(E) }
func (E Experts) Less(i, j int) bool { return E[i].Compare(E[j]) }
func (E Experts) Swap(i, j int)      { E[i], E[j] = E[j], E[i] }

type MapExperts map[string]Expert
type RaceMapExperts map[string]MapExperts

func (E MapExperts) GetText(name string) []string {
	e := make(Experts, 0, len(E))
	for _, expert := range E {
		e = append(e, expert)
	}
	return e.GetText(name)
}

func (E Experts) GetText(name string) []string {
	sort.Sort(E)
	s := make([]string, 0, len(E)+1)
	s = append(s, name+":")
	var sw float64
	for _, e := range E {
		sw += e.Weight
	}
	for _, e := range E {
		if e.Weight == 0.0 {
			continue
		}
		s = append(s, fmt.Sprintf("%s Forecast: Dem expected vote margin: %.2f%% (Weight: %.1f%%)", e.Expert, (e.ConcentrationParams["D"]-e.ConcentrationParams["R"])/(e.ConcentrationParams["D"]+e.ConcentrationParams["R"])*100, e.Weight/sw*100))
	}
	return s
}

func LoadExperts(ftype string, E *RaceMapExperts) (error, time.Time) {
	e, t := LoadFileCache("forecast/"+ftype+"_experts.json", func(r io.Reader) interface{} {
		err := json.NewDecoder(r).Decode(E)
		if err != nil {
			panic(err)
		}
		return *E
	})
	*E = e.(RaceMapExperts)
	return nil, t
}

func SaveExperts(name string, experts RaceMapExperts) {
	f, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(experts); err != nil {
		panic(err)
	}
}

func CombineExpertRatings(rating_data []map[string][2]float64, weights []float64, names []string) (map[string][2]float64, RaceMapExperts) {
	raceExperts := make(RaceMapExperts)
	ratings := make(map[string][2]float64)
	for i, rats := range rating_data {
		for st, rat := range rats {
			ratings[st] = [2]float64{ratings[st][0] + rat[0]*weights[i], ratings[st][1] + rat[1]*weights[i]}
			if _, ok := raceExperts[st]; !ok {
				raceExperts[st] = make(MapExperts)
			}
			raceExperts[st][names[i]] = Expert{Expert: names[i], ConcentrationParams: map[string]float64{"D": rat[0], "R": rat[1]}, Weight: weights[i]}
		}
	}
	return ratings, raceExperts
}
