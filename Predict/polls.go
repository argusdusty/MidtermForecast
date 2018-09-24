package Predict

import (
	. "MidtermForecast/APIs"
	. "MidtermForecast/Utils"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"strconv"
	"time"
)

var (
	// Pre-computed in historical_optimize.go
	RAND_POLL_SHIFT  = 0.0184622107401895   // Base per-poll error
	DAILY_POLL_SHIFT = 0.002574543816481247 // Per-day per-poll average drift
	POLLING_WEIGHT   = 0.749368253464103    // Multiplier to beta/dirichlet params for polling average in forecast. Larger means more confident
)

func init() {
	LoadPollsWeightBias()
}

func GetPollWeights(polls []Poll, parties map[string]string, electionDate time.Time) []float64 {
	used := map[string]int{}
	for _, poll := range polls {
		used[poll.Pollster]++
	}
	ws := make([]float64, len(polls))
	for i, poll := range polls {
		w := 0.25
		if wb, ok := PollsterBiases[poll.Pollster]; ok {
			w = wb[0]
		}
		w /= float64(used[poll.Pollster])

		// Expected error
		err := math.Pow(poll.Number, -0.5)
		serr := RAND_POLL_SHIFT

		// Model time-based error shift
		d := poll.EndDate.Sub(poll.StartDate).Hours() / 24.0
		ds := electionDate.Sub(poll.EndDate).Hours() / 24.0
		if ds < 0 {
			ds = 0
		}
		if d != 0 {
			serr += DAILY_POLL_SHIFT * (math.Pow(d+ds+1, 1.5) - math.Pow(ds+1, 1.5)) * 2 / (3 * d) // Average of integral of same-day case below
		} else {
			serr += DAILY_POLL_SHIFT * math.Sqrt(ds+1)
		}
		err += math.Abs(serr)

		w *= math.Pow(err, -2.0) / poll.Number

		// Get subpopulation weight (LV best)
		switch poll.Subpopulation {
		case "LV":
			w *= 1.0
		case "RV":
			w *= 0.5
		case "A", "V":
			w *= 0.25
		default:
			w *= 0.1
		}
		ws[i] = w
	}
	return ws
}

type Polls []Poll

func (P Polls) Len() int           { return len(P) }
func (P Polls) Less(i, j int) bool { return P[i].Compare(P[j]) }
func (P Polls) Swap(i, j int)      { P[i], P[j] = P[j], P[i] }

type RacePoll struct {
	Poll
	Race string `json:"race"`
}

type RacePolls []RacePoll

func (P RacePolls) Len() int           { return len(P) }
func (P RacePolls) Less(i, j int) bool { return P[i].Compare(P[j].Poll) }
func (P RacePolls) Swap(i, j int)      { P[i], P[j] = P[j], P[i] }

type RaceMapPolls map[string][]Poll

func (P RaceMapPolls) GetText(name string) []string {
	p := make(RacePolls, 0)
	for race, polls := range P {
		for _, poll := range polls {
			p = append(p, RacePoll{Poll: poll, Race: race})
		}
	}
	return p.GetText(name)
}

func (P RacePolls) GetText(name string) []string {
	sort.Sort(P)
	s := make([]string, len(P)+1)
	s[0] = name + ":"
	for i, p := range P {
		bias := PollsterBiases[p.Pollster][1]
		var bstr string
		if bias >= 0 {
			bstr = "D+" + fmt.Sprintf("%.2f%%", bias*100)
		} else if bias < 0 {
			bstr = "R+" + fmt.Sprintf("%.2f%%", -bias*100)
		}
		if p.URL != "" {
			s[i+1] = fmt.Sprintf("<a href=\"%s\">%s</a> - <a href=\"%s\">%s</a> (Bias: %s): %s-%s - Sample Size: %d (%s) - Dem: %.1f%%, GOP: %.1f%%", p.Race, p.Race, p.URL, p.Pollster, bstr, p.StartDate.Format("2006 Jan 02"), p.EndDate.Format("2006 Jan 02"), int(p.Number), p.Subpopulation, p.Candidates["D"]*100, p.Candidates["R"]*100)
		} else {
			s[i+1] = fmt.Sprintf("<a href=\"%s\">%s</a>  - %s (Bias: %s): %s-%s - Sample Size: %d (%s) - Dem: %.1f%%, GOP: %.1f%%", p.Race, p.Race, p.Pollster, bstr, p.StartDate.Format("2006 Jan 02"), p.EndDate.Format("2006 Jan 02"), int(p.Number), p.Subpopulation, p.Candidates["D"]*100, p.Candidates["R"]*100)
		}
	}
	return s
}

func (P Polls) GetText(name string) []string {
	sort.Sort(P)
	weights := GetPollWeights(P, map[string]string{"D": "D", "R": "R"}, time.Date(2018, 11, 7, 0, 0, 0, 0, time.UTC))
	var sw float64
	for _, w := range weights {
		sw += w
	}
	s := make([]string, len(P)+1)
	s[0] = name + ":"
	for i, p := range P {
		bias := PollsterBiases[p.Pollster][1]
		var bstr string
		if bias >= 0 {
			bstr = "D+" + fmt.Sprintf("%.1f%%", bias*100)
		} else if bias < 0 {
			bstr = "R+" + fmt.Sprintf("%.1f%%", -bias*100)
		}
		if p.URL != "" {
			s[i+1] = fmt.Sprintf("<a href=\"%s\">%s</a> (Weight: %.2f%% Bias: %s): %s-%s - Sample Size: %d (%s) - Dem: %.1f%%, GOP: %.1f%%", p.URL, p.Pollster, weights[i]/sw*100, bstr, p.StartDate.Format("2006 Jan 02"), p.EndDate.Format("2006 Jan 02"), int(p.Number), p.Subpopulation, p.Candidates["D"]*100, p.Candidates["R"]*100)
		} else {
			s[i+1] = fmt.Sprintf("%s (Weight: %.2f%% Bias: %s): %s-%s - Sample Size: %d (%s) - Dem: %.1f%%, GOP: %.1f%%", p.Pollster, weights[i]/sw*100, bstr, p.StartDate.Format("2006 Jan 02"), p.EndDate.Format("2006 Jan 02"), int(p.Number), p.Subpopulation, p.Candidates["D"]*100, p.Candidates["R"]*100)
		}
	}
	return s
}

func LoadPolls(ftype string, P *RaceMapPolls) (error, time.Time) {
	p, t := LoadFileCache("forecast/"+ftype+"_polls.json", func(r io.Reader) interface{} {
		err := json.NewDecoder(r).Decode(P)
		if err != nil {
			panic(err)
		}
		return *P
	})
	*P = p.(RaceMapPolls)
	return nil, t
}

func SavePolls(name string, polls RaceMapPolls) {
	f, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(polls); err != nil {
		panic(err)
	}
}

func LoadPollsWeightBias() {
	r := LoadCache("https://raw.githubusercontent.com/fivethirtyeight/data/master/pollster-ratings/pollster-ratings.csv", "cache/pollster-ratings.csv", -1, func(r io.Reader) interface{} {
		if b, err := ioutil.ReadAll(r); err == nil {
			return b
		} else {
			panic(err)
		}
	}).([]byte)
	reader := csv.NewReader(bytes.NewReader(r))
	reader.Read()
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		pollster := record[0]
		simpleexpectederr, err := strconv.ParseFloat(record[13], 64)
		if err != nil {
			panic(err)
		}
		predplusminus, err := strconv.ParseFloat(record[7], 64)
		if err != nil {
			panic(err)
		}
		weight := math.Pow(1.0-predplusminus/simpleexpectederr, 2.0)
		weight *= weight
		var bias float64
		if record[9] != "" {
			bias, err = strconv.ParseFloat(record[9][2:], 64)
			bias /= 100.0
		}
		PollsterBiases[pollster] = [2]float64{weight, bias}
	}
}

// Not used
func LoadPollsWeightBiasCustom() {
	LoadPollsWeightBias() // Initial guesses
	r := LoadCache("https://raw.githubusercontent.com/fivethirtyeight/data/master/pollster-ratings/raw-polls.csv", "cache/raw-polls.csv", -1, func(r io.Reader) interface{} {
		if b, err := ioutil.ReadAll(r); err == nil {
			return b
		} else {
			panic(err)
		}
	}).([]byte)
	reader := csv.NewReader(bytes.NewReader(r))
	reader.Read()
	var data_map = map[string][][4]float64{} // sample_size, guess, result, bias_dir for each poll
	var bias_map = map[string]float64{}      // Absolute sum of error per pollster
	var bias_n_map = map[string]float64{}    // Number of D vs R polls per pollster
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		pollster := record[6]
		sample_size_raw := record[9]
		sample_size, err := strconv.ParseFloat(sample_size_raw, 64)
		if err != nil {
			panic(err)
		}
		cand1_pct_raw := record[11]
		cand1_pct, err := strconv.ParseFloat(cand1_pct_raw, 64)
		if err != nil {
			panic(err)
		}
		cand2_pct_raw := record[13]
		cand2_pct, err := strconv.ParseFloat(cand2_pct_raw, 64)
		if err != nil {
			panic(err)
		}
		cand1_actual_raw := record[17]
		cand1_actual, err := strconv.ParseFloat(cand1_actual_raw, 64)
		if err != nil {
			panic(err)
		}
		cand2_actual_raw := record[18]
		cand2_actual, err := strconv.ParseFloat(cand2_actual_raw, 64)
		if err != nil {
			panic(err)
		}
		perr := cand1_pct/(cand1_pct+cand2_pct) - cand1_actual/(cand1_actual+cand2_actual)
		var bias_dir float64
		if record[10] == "Democrat" && record[12] == "Republican" {
			bias_map[pollster] += perr
			bias_n_map[pollster]++
			bias_dir = 1
		} else if record[10] == "Republican" && record[12] == "Democrat" {
			bias_map[pollster] -= perr
			bias_n_map[pollster]++
			bias_dir = -1
		}
		if _, ok := data_map[pollster]; !ok {
			data_map[pollster] = make([][4]float64, 0)
		}
		data_map[pollster] = append(data_map[pollster], [4]float64{sample_size, cand1_pct / (cand1_pct + cand2_pct), cand1_actual / (cand1_actual + cand2_actual), bias_dir})
	}
	for pollster, data := range data_map {
		if bias_n_map[pollster] < 20 {
			// Bad sample size
			continue
		}
		var mult float64
		bias := bias_map[pollster] / bias_n_map[pollster]
		//println(pollster, bias_map[pollster], bias_n_map[pollster])
		for _, r := range data {
			real_err := math.Pow(r[1]-r[2]-r[3]*bias, 2)
			exp_err := 1 / r[0] //r[1] * (1 - r[2]) / r[0]
			mult += real_err / exp_err
			//println(pollster, r[0], r[1], r[2], r[3], bias, real_err, exp_err, mult)
		}
		mult /= float64(len(data) - 1)
		PollsterBiases[pollster] = [2]float64{1 / mult, bias}
	}
}

func GetConcentrationParams(polls []Poll, parties map[string]string, electionDate time.Time) map[string]float64 {
	var c_scores = make(map[string]float64)
	sort.Sort(Polls(polls))
	used := map[string]int{}
	for _, poll := range polls {
		used[poll.Pollster]++
	}
	for _, poll := range polls {
		// Get 538 weight
		w := 0.25
		b := 0.0
		if wb, ok := PollsterBiases[poll.Pollster]; ok {
			w = wb[0]
			b = wb[1]
		}
		w /= float64(used[poll.Pollster])

		// Expected error
		err := math.Pow(poll.Number, -0.5)
		serr := RAND_POLL_SHIFT

		// Model time-based error shift
		d := poll.EndDate.Sub(poll.StartDate).Hours() / 24.0
		ds := electionDate.Sub(poll.EndDate).Hours() / 24.0
		if ds < 0 {
			ds = 0
		}
		if d != 0 {
			serr += DAILY_POLL_SHIFT * (math.Pow(d+ds+1, 1.5) - math.Pow(ds+1, 1.5)) * 2 / (3 * d) // Average of integral of same-day case below
		} else {
			serr += DAILY_POLL_SHIFT * math.Sqrt(ds+1)
		}
		err += math.Abs(serr)

		w *= math.Pow(err, -2.0) / poll.Number

		// Get subpopulation weight (LV best)
		switch poll.Subpopulation {
		case "LV":
			w *= 1.0
		case "RV":
			w *= 0.5
		case "A", "V":
			w *= 0.25
		default:
			w *= 0.1
		}

		var d_score, r_score, total_score float64
		for c, v := range poll.Candidates {
			if party, ok := parties[c]; ok {
				if party == "D" {
					d_score += v
				} else if party == "R" {
					r_score += v
				}
			}
			total_score += v
		}
		for c, v := range poll.Candidates {
			if v == 0 {
				continue
			}
			if party, ok := parties[c]; ok {
				if party == "D" {
					c_scores[c] += v * w * (1.0 - b*v/d_score) * poll.Number //* (v / total_score) * (1 - v/total_score)
				} else if party == "R" {
					c_scores[c] += v * w * (1.0 + b*v/r_score) * poll.Number //* (v / total_score) * (1 - v/total_score)
				}
			} else {
				c_scores[c] += v * w * poll.Number
			}
		}
	}
	return c_scores
}

func CombinePolls(polling_data []map[string][]Poll) map[string][]Poll {
	polls := make(map[string][]Poll)
	for _, poll := range polling_data {
		for k, v := range poll {
			if _, ok := polls[k]; !ok {
				polls[k] = v
			} else {
				for _, p := range v {
					f := false
					for _, p2 := range polls[k] {
						if p.Pollster == p2.Pollster {
							if (p.EndDate.After(p2.EndDate) && p.EndDate.Sub(p2.EndDate) < 48*time.Hour) || (p2.EndDate.After(p.EndDate) && p2.EndDate.Sub(p.EndDate) < 48*time.Hour) {
								f = true
								break
							}
						}
					}
					if !f {
						polls[k] = append(polls[k], p)
					}
				}
			}
		}
	}
	return polls
}

func GetPollingRatings(polls map[string][]Poll, election_date time.Time) map[string][2]float64 {
	ratings := make(map[string][2]float64)
	for k, v := range polls {
		params := GetConcentrationParams(v, map[string]string{"D": "D", "R": "R"}, election_date)
		ratings[k] = [2]float64{params["D"] * POLLING_WEIGHT, params["R"] * POLLING_WEIGHT}
	}
	return ratings
}
