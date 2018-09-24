package Predict

import (
	. "MidtermForecast/Utils"
	"bytes"
	"encoding/json"
	"fmt"
	"gonum.org/v1/gonum/mathext"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"
)

type ShortForecast struct {
	Date                   time.Time `json:"date"`
	DemMajorityProbability float64   `json:"dem_maj_prob"`
	DemSeatsExpected       float64   `json:"dem_seats_exp"`
}

type Forecast struct {
	ShortForecast
	DemSeatsLow       int                        `json:"dem_seats_lo"`
	DemSeatsHigh      int                        `json:"dem_seats_hi"`
	RaceProbabilities map[string]RaceProbability `json:"race_probabilities"`
	SeatProbabilities []SeatProbability          `json:"seat_probabilities"`
	Past              []ShortForecast            `json:"past"`
}

type ShortRaceProbability struct {
	Date                time.Time          `json:"date"`
	DemWinProbability   float64            `json:"dem_win_prob"`
	DemVoteExpected     float64            `json:"dem_vote_exp"`
	DemVoteLow          float64            `json:"dem_vote_lo"`
	DemVoteHigh         float64            `json:"dem_vote_hi"`
	ConcentrationParams map[string]float64 `json:"concentration_params"`
}

type RaceProbability struct {
	ShortRaceProbability
	Race    string                 `json:"race"`
	Sources map[string][2]float64  `json:"sources"`
	Past    []ShortRaceProbability `json:"past"`
}

type SeatProbability struct {
	DemSeats    int     `json:"dem_seats"`
	Probability float64 `json:"probability"`
}

type RaceProbabilities []RaceProbability

func (R RaceProbabilities) Len() int { return len(R) }
func (R RaceProbabilities) Less(i, j int) bool {
	return R[i].DemWinProbability < R[j].DemWinProbability
}
func (R RaceProbabilities) Swap(i, j int) {
	R[i], R[j] = R[j], R[i]
}

type SeatProbabilities []SeatProbability

func (S SeatProbabilities) Len() int { return len(S) }
func (S SeatProbabilities) Less(i, j int) bool {
	return S[i].Probability < S[j].Probability
}
func (S SeatProbabilities) Swap(i, j int) {
	S[i], S[j] = S[j], S[i]
}

func color(p float64) (r, g, b int) {
	r = int(256 * (1 - p))
	if r > 255 {
		r = 255
	}
	g = 0
	b = int(256 * p)
	if b > 255 {
		b = 255
	}
	return
}

func (R *RaceProbability) Fix(sources map[string][2]float64) {
	R.DemVoteExpected = R.ConcentrationParams["D"] / (R.ConcentrationParams["D"] + R.ConcentrationParams["R"])
	if R.ConcentrationParams["D"] == 0.0 {
		R.DemVoteLow = 0.0
		R.DemVoteHigh = 0.0
	} else if R.ConcentrationParams["R"] == 0.0 {
		R.DemVoteLow = 1.0
		R.DemVoteHigh = 1.0
	} else {
		R.DemVoteLow = mathext.InvRegIncBeta(R.ConcentrationParams["D"], R.ConcentrationParams["R"], 0.05)
		R.DemVoteHigh = mathext.InvRegIncBeta(R.ConcentrationParams["D"], R.ConcentrationParams["R"], 0.95)
	}
	if sources != nil {
		R.Sources = sources
	}
}

func (R RaceProbability) GetText(name string) []string {
	r, g, b := color(R.DemWinProbability)
	sources := make([]string, 0, len(R.Sources))
	var sv float64
	for k, s := range R.Sources {
		sources = append(sources, k)
		for _, v := range s {
			sv += v
		}
	}
	sort.Strings(sources)
	sourceText := make([]string, 0, len(sources))
	for _, k := range sources {
		s := R.Sources[k]
		var ss float64
		for _, v := range s {
			ss += v
		}
		if ss == 0 {
			continue
		}
		sourceText = append(sourceText, fmt.Sprintf("<a href=\"/%s/%s/%s\">%s Forecast</a>: Dem expected vote margin: %.2f%% (Weight: %.1f%%)", name, R.Race, k, strings.Title(k), (s[0]-s[1])/ss*100.0, ss/sv*100.0))
	}
	return append([]string{fmt.Sprintf("<a href=\"/%s\">Full %s results</a>", name, name), fmt.Sprintf("<a href=\"/%s/%s\">%s</a> Dem win probability: <span style=\"color:rgb(%d,%d,%d)\">%.1f%%</span> Dem expected vote margin: %.2f%% (95%% Confidence Interval: %.2f%% <-> %.2f%%)", name, R.Race, R.Race, r, g, b, R.DemWinProbability*100.0, (2*R.DemVoteExpected-1)*100.0, R.DemVoteLow*100.0, R.DemVoteHigh*100.0)}, sourceText...)
}

func (F Forecast) GetText(name string) []string {
	var s []string
	s = append(s, fmt.Sprintf("<a href=\"/\">All Forecasts</a>"))
	// Step 1. Overall Dem win prob/expected seats
	if _, ok := F.RaceProbabilities["KS"]; !ok {
		// Not Governors
		s = append(s, fmt.Sprintf("Dem win probability: %.2f%%", F.DemMajorityProbability*100.0))
		s = append(s, fmt.Sprintf("Dem expected number of seats: %.2f", F.DemSeatsExpected))
	}
	// Step 2. Close races (10% <-> 90%) Dem win prob
	var counts [7][]string
	//sort.Sort(RaceProbabilities(F.RaceProbabilities))
	for _, r := range F.RaceProbabilities {
		if r.DemWinProbability > 0.05 && r.DemWinProbability < 0.95 {
			//s = append(s, r.GetText(name)[1:]...)
		}
		if r.DemWinProbability >= 0.0 && r.DemWinProbability < 0.05 {
			counts[0] = append(counts[0], r.Race)
		} else if r.DemWinProbability >= 0.05 && r.DemWinProbability < 0.25 {
			counts[1] = append(counts[1], r.Race)
		} else if r.DemWinProbability >= 0.25 && r.DemWinProbability < 0.4 {
			counts[2] = append(counts[2], r.Race)
		} else if r.DemWinProbability >= 0.4 && r.DemWinProbability <= 0.6 {
			counts[3] = append(counts[3], r.Race)
		} else if r.DemWinProbability > 0.6 && r.DemWinProbability <= 0.75 {
			counts[4] = append(counts[4], r.Race)
		} else if r.DemWinProbability > 0.75 && r.DemWinProbability <= 0.95 {
			counts[5] = append(counts[5], r.Race)
		} else if r.DemWinProbability > 0.95 && r.DemWinProbability <= 1.0 {
			counts[6] = append(counts[6], r.Race)
		}
	}
	// Step 3. Safe/Likely/Lean/Toss-up
	names := []string{"Safe R", "Likely R", "Lean R", "Toss-up", "Lean D", "Likely D", "Safe D"}
	shtml := ""
	for i := 0; i < 7; i++ {
		shtml += fmt.Sprintf("<div class=\"districts-%d\">%s seats: %d", i, names[i], len(counts[i]))
		shtml += fmt.Sprintf(`<style>
.districts-hover-%d {
	display: none;
	position: absolute;
	background-color: #f1f1f1;
	min-width: 160px;
	box-shadow: 0px 8px 16px 0px rgba(0,0,0,0.2);
	z-index: 99999;
	overflow-y: scroll;
	max-height: 500px;
}
.districts-%d {
	display: inline-block;
	padding-left: 10px;
	padding-right: 10px;
	border-left: 1px solid #313030;
	border-right: 1px solid #313030;
}
.districts-%d:hover .districts-hover-%d {
	display: block;
}
</style>`, i, i, i, i)
		shtml += fmt.Sprintf("<div class=\"districts-hover-%d\">", i)
		sort.Strings(counts[i])
		for j, race := range counts[i] {
			shtml += fmt.Sprintf("<a href=\"/%s/%s\" style=\"color: black; text-decoration: none; display: block; position: relative; padding-top:5px; padding-bottom:5px;\">%s</a>", name, race, race)
			if j != len(counts[i])-1 {
				shtml += "<hr>"
			}
		}
		shtml += "</div>"
		shtml += "</div>"
	}
	s = append(s, shtml)
	return s
}

func GenForecast(seats []float64, rp map[string]RaceProbability, past Forecast, now time.Time) Forecast {
	var f Forecast
	f.Date = now
	f.RaceProbabilities = rp
	f.SeatProbabilities = make([]SeatProbability, len(seats))
	var cdf float64
	for i, v := range seats {
		f.SeatProbabilities[i] = SeatProbability{
			DemSeats:    i,
			Probability: v,
		}
		f.DemSeatsExpected += float64(i) * v
		if cdf <= 0.05 && cdf+v > 0.05 {
			f.DemSeatsLow = i
		}
		if cdf <= 0.95 && cdf+v > 0.95 {
			f.DemSeatsHigh = i
		}
		cdf += v
		if 2*i > len(seats) {
			f.DemMajorityProbability += v
		}
	}
	for k, st := range f.RaceProbabilities {
		st.Date = now
		st.Past = append(past.RaceProbabilities[k].Past, st.ShortRaceProbability)
		f.RaceProbabilities[k] = st
	}
	f.Past = append(past.Past, f.ShortForecast)
	return f
}

func LoadForecast(ftype string, F *Forecast) ([]byte, error, time.Time) {
	f, t := LoadFileCache("forecast/"+ftype+"_forecast.json", func(r io.Reader) interface{} {
		data, err := ioutil.ReadAll(r)
		if err != nil {
			panic(err)
		}
		err = json.NewDecoder(bytes.NewReader(data)).Decode(F)
		if err != nil {
			panic(err)
		}
		return RawObject{Raw: data, Object: *F}
	})
	*F = f.(RawObject).Object.(Forecast)
	return f.(RawObject).Raw, nil, t
}

func SaveForecast(name string, forecast Forecast) {
	f, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(forecast); err != nil {
		panic(err)
	}
}
