package Predict

import (
	"time"
)

func CombineRatings(rating_data []map[string][2]float64, weights []float64) map[string][2]float64 {
	ratings := make(map[string][2]float64)
	for i, rats := range rating_data {
		for st, rat := range rats {
			ratings[st] = [2]float64{ratings[st][0] + rat[0]*weights[i], ratings[st][1] + rat[1]*weights[i]}
		}
	}
	return ratings
}

// Same as CombineRatings but no weights
func MergeRatings(rating_data []map[string][2]float64) map[string][2]float64 {
	ratings := make(map[string][2]float64)
	for _, rats := range rating_data {
		for st, rat := range rats {
			ratings[st] = [2]float64{ratings[st][0] + rat[0], ratings[st][1] + rat[1]}
		}
	}
	return ratings
}

func MergeRaces(races map[string]struct{}, unchallenged_races [2]map[string]struct{}, ratings map[string]map[string][2]float64, election_date, now time.Time) (r map[string][2]float64, sources map[string]map[string][2]float64) {
	days := election_date.Sub(now).Hours() / 24

	r = map[string][2]float64{}
	sources = map[string]map[string][2]float64{}
	for k := range races {
		sources[k] = map[string][2]float64{}
		alpha := 0.0
		beta := 0.0
		for source, rats := range ratings {
			rat := rats[k]
			alpha += rat[0]
			beta += rat[1]

			sources[k][source] = rat
			if _, ok := unchallenged_races[0][k]; ok {
				sources[k][source] = [2]float64{rat[0], 0}
			} else if _, ok := unchallenged_races[1][k]; ok {
				sources[k][source] = [2]float64{0, rat[1]}
			}
		}
		alpha, beta = AdjustRaceError(alpha+1, beta+1, days)
		if _, ok := unchallenged_races[0][k]; ok {
			beta = 0.0
		} else if _, ok := unchallenged_races[1][k]; ok {
			alpha = 0.0
		}
		if alpha < 0 {
			alpha = 0
			beta = 1
		} else if beta < 0 {
			alpha = 1
			beta = 0
		}
		r[k] = [2]float64{alpha, beta}
	}
	return
}
