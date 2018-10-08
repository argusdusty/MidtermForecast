package main

import (
	. "MidtermForecast/APIs"
	. "MidtermForecast/Predict"
	. "MidtermForecast/Utils"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

var (
	MODEL_WEIGHT_538 = 0.9
	SUM_BETA         = 16.0
)

var (
	election_date                = time.Date(2018, 11, 7, 0, 0, 0, 0, time.UTC)
	polls_538_senate             = map[string][]Poll{}
	polls_538_house              = map[string][]Poll{}
	polls_538_gov                = map[string][]Poll{}
	congressional_ballot float64 = 0.08
	house_forecast_538           = map[string]float64{} // district -> pvi
	senate_forecast_538          = map[string]float64{} // district -> pvi
	past_senate_pvi              = map[string]float64{}
	past_house_pvi               = map[string]float64{}
	past_gov_pvi                 = map[string]float64{
		"AL": 0 - 0.05887177043707405,
		"AK": 0 - 0.05887177043707405,
		"AZ": 0 - 0.05887177043707405,
		"AR": 0 - 0.05887177043707405,
		"CA": 0 - 0.05887177043707405,
		"CO": 0 - 0.05887177043707405,
		"CT": 0 - 0.05887177043707405,
		"FL": 0 - 0.05887177043707405,
		"GA": 0 - 0.05887177043707405,
		"HI": 0 - 0.05887177043707405,
		"ID": 0 - 0.05887177043707405,
		"IA": 0 - 0.05887177043707405,
		"KS": 0 - 0.05887177043707405,
		"ME": 0 - 0.05887177043707405,
		"MD": 0 - 0.05887177043707405,
		"MA": 0 - 0.05887177043707405,
		"MN": 0 - 0.05887177043707405,
		"NV": 0 - 0.05887177043707405,
		"NE": 0 - 0.05887177043707405,
		"NH": 0 - 0.05887177043707405,
		"NM": 0 - 0.05887177043707405,
		"NY": 0 - 0.05887177043707405,
		"OH": 0 - 0.05887177043707405,
		"OK": 0 - 0.05887177043707405,
		"OR": 0 - 0.05887177043707405,
		"PA": 0 - 0.05887177043707405,
		"RI": 0 - 0.05887177043707405,
		"SC": 0 - 0.05887177043707405,
		"SD": 0 - 0.05887177043707405,
		"TN": 0 - 0.05887177043707405,
		"TX": 0 - 0.05887177043707405,
		"VT": 0 - 0.05887177043707405,
		"WI": 0 - 0.05887177043707405,
		"WY": 0 - 0.05887177043707405,
	}
	fundraising_senate_ratios = map[string]float64{}
	fundraising_house_ratios  = map[string]float64{}

	house_polls  = map[string][]Poll{}
	senate_polls = map[string][]Poll{}
	gov_polls    = map[string][]Poll{}

	house_experts  RaceMapExperts
	senate_experts RaceMapExperts
	gov_experts    RaceMapExperts

	house_fundamentals  RaceFundamentals
	senate_fundamentals RaceFundamentals
	gov_fundamentals    RaceFundamentals

	house_nyt_polls  map[string][]Poll
	senate_nyt_polls map[string][]Poll
)

type District struct {
	State    string  `json:"state"`
	MapType  string  `json:"maptype"`
	District string  `json:"district"`
	PVI      float64 `json:"pvi"`
}

func setup() {
	congressional_ballot = Load538GenericBallot()
	cb := congressional_ballot
	house_forecast_538, _, congressional_ballot = Load538HouseForecast()
	senate_forecast_538, _ = Load538SenateForecast()
	if congressional_ballot == 0.0 {
		congressional_ballot = cb
	} else {
		congressional_ballot = cb*0.5 + congressional_ballot*0.5
	}
	polls_538_senate, polls_538_house, polls_538_gov = Load538Polls()

	pvi_2012 := 0.012023691374835306
	for k, v := range LoadCNNSenateResults("2012") {
		if k == "MA-2" || k == "NJ-2" {
			continue
		}
		past_senate_pvi[k] = (v[0]-v[1])/(v[0]+v[1]) - pvi_2012
	}
	for k, v := range map[string]float64{"AZ": -1, "CA": 1, "DE": 1, "FL": 1, "MD": 1, "MA": -1, "MI": 1, "MN": 1, "MS": -1, "MO": 1, "MT": 1, "NV": -1, "NJ": 1, "NY": 1, "OH": 1, "PA": 1, "TN": -1, "UT": -1, "VT": 1, "WA": 1, "WV": 1, "WY": -1} {
		past_senate_pvi[k] -= v * INCUMBENT_ADVANTAGE_PVI
	}
	past_senate_pvi["MN-2"] = (1053205.0-850227.0)/(1053205.0+850227.0) - (35624357-40081282)/(35624357+40081282)
	past_senate_pvi["MS-2"] = (239439.0-378481.0)/(239439.0+378481.0) - (35624357-40081282)/(35624357+40081282) + INCUMBENT_ADVANTAGE_PVI

	pvi_2016 := -0.011182528000377494
	_, incumbents_house_2016, _ := LoadCookHouseRatingsHtml(139361)
	for k, v := range LoadCNNHouseResults("2016") {
		past_house_pvi[k] = (v[0]-v[1])/(v[0]+v[1]) - pvi_2016
		if incumbents_house_2016[k] == "R" {
			past_house_pvi[k] += INCUMBENT_ADVANTAGE_PVI
		} else if incumbents_house_2016[k] == "D" {
			past_house_pvi[k] -= INCUMBENT_ADVANTAGE_PVI
		}
	}

	pvi_2014 := -0.05887177043707405
	_, incumbents_gov_2016, _ := LoadCookGovRatingsHtml(139257)
	for k, v := range LoadCNNGovResults("2014") {
		past_gov_pvi[k] = (v[0]-v[1])/(v[0]+v[1]) - pvi_2014
		if incumbents_gov_2016[k] == "R" {
			past_gov_pvi[k] += INCUMBENT_ADVANTAGE_PVI
		} else if incumbents_gov_2016[k] == "D" {
			past_gov_pvi[k] -= INCUMBENT_ADVANTAGE_PVI
		}
	}

	fundraising_senate_ratios = LoadFECRaces("S", 2018)
	fundraising_house_ratios = LoadFECRaces("H", 2018)

	house_nyt_polls, senate_nyt_polls = LoadNYTLivePolls()
}

func merge_senate_classes(now time.Time) (map[string][2]float64, map[string]map[string][2]float64) {
	races := map[string]struct{}{}
	cook_ratings, incumbents, cook_pvis := LoadCookSenateRatingsHtml(0)
	for k := range cook_ratings {
		races[k] = struct{}{}
	}

	expert_ratings := []map[string][2]float64{cook_ratings, GetBetasFromPVIs(senate_forecast_538)}
	expert_weights := []float64{COOK_WEIGHT / 4, 3 * COOK_WEIGHT / 4}

	polling_data_538 := polls_538_senate
	f, err := os.Open("extra_senate_polls.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var extra_polls map[string][]Poll
	dec := json.NewDecoder(f)
	dec.Decode(&extra_polls)
	polling_data := []map[string][]Poll{polling_data_538, extra_polls, senate_nyt_polls}

	pvi_estimates := []map[string]float64{cook_pvis, past_senate_pvi}
	pvi_weights := []float64{COOK_PVI_WEIGHT, PAST_PVI_WEIGHT}

	unchallenged_races := [2]map[string]struct{}{{"CA": {}}, nil}

	senate_polls = CombinePolls(polling_data)
	polling_ratings := GetPollingRatings(senate_polls, election_date)
	var fundamentals_ratings map[string][2]float64
	fundamentals_ratings, senate_fundamentals = GetFundamentalsRatings(incumbents, fundraising_senate_ratios, pvi_estimates, pvi_weights, []string{"Cook", "Historical"}, congressional_ballot)
	var experts_ratings map[string][2]float64
	experts_ratings, senate_experts = CombineExpertRatings(expert_ratings, expert_weights, []string{"Cook", "538"})
	ratings := map[string]map[string][2]float64{"polling": polling_ratings, "fundamentals": fundamentals_ratings, "experts": experts_ratings}

	return MergeRaces(races, unchallenged_races, ratings, election_date, now)
}

func merge_house_classes(now time.Time) (map[string][2]float64, map[string]map[string][2]float64) {
	races := map[string]struct{}{}
	f, err := os.Open("districts.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var districts []District
	dec := json.NewDecoder(f)
	if err := dec.Decode(&districts); err != nil {
		panic(err)
	}
	for _, d := range districts {
		races[d.State+"-"+d.District] = struct{}{}
	}

	cook_ratings, incumbents, cook_pvis := LoadCookHouseRatingsHtml(0)

	expert_ratings := []map[string][2]float64{cook_ratings, GetBetasFromPVIs(house_forecast_538)}
	expert_weights := []float64{COOK_WEIGHT / 4, 3 * COOK_WEIGHT / 4}

	polling_data_538 := polls_538_house
	f, err = os.Open("extra_house_polls.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var extra_polls map[string][]Poll
	dec = json.NewDecoder(f)
	dec.Decode(&extra_polls)
	polling_data := []map[string][]Poll{polling_data_538, extra_polls, house_nyt_polls}

	pvi_estimates_dist := make(map[string]float64, len(districts))
	for _, d := range districts {
		pvi_estimates_dist[d.State+"-"+d.District] = d.PVI / 100.0
	}
	pvi_estimates := []map[string]float64{cook_pvis, pvi_estimates_dist, past_house_pvi}
	pvi_weights := []float64{COOK_PVI_WEIGHT / 2, COOK_PVI_WEIGHT / 2, PAST_PVI_WEIGHT}

	unchallenged_races := [2]map[string]struct{}{
		{"AL-07": {}, "AZ-07": {}, "CA-13": {}, "CA-20": {}, "CA-27": {}, "CA-34": {}, "CA-40": {}, "CA-44": {}, "CA-05": {}, "CA-06": {}, "FL-10": {}, "FL-14": {}, "FL-20": {}, "FL-21": {}, "FL-24": {}, "GA-05": {}, "LA-02": {}, "MA-01": {}, "MA-04": {}, "MA-07": {}, "MA-08": {}, "MI-13": {}, "MS-02": {}, "NY-16": {}, "NY-17": {}, "NY-05": {}, "NY-06": {}, "NY-07": {}, "NY-08": {}, "OR-03": {}, "PA-18": {}, "TX-20": {}, "TX-28": {}, "TX-30": {}, "TX-09": {}, "VA-03": {}, "WA-02": {}, "WI-02": {}},
		{"CA-08": {}, "GA-08": {}, "MI-01": {}, "NC-03": {}},
	}

	house_polls = CombinePolls(polling_data)
	polling_ratings := GetPollingRatings(house_polls, election_date)
	var fundamentals_ratings map[string][2]float64
	fundamentals_ratings, house_fundamentals = GetFundamentalsRatings(incumbents, fundraising_house_ratios, pvi_estimates, pvi_weights, []string{"Cook", "538", "Historical"}, congressional_ballot)
	var experts_ratings map[string][2]float64
	experts_ratings, house_experts = CombineExpertRatings(expert_ratings, expert_weights, []string{"Cook", "538"})
	ratings := map[string]map[string][2]float64{"polling": polling_ratings, "fundamentals": fundamentals_ratings, "experts": experts_ratings}

	return MergeRaces(races, unchallenged_races, ratings, election_date, now)
}

func merge_gov_classes(now time.Time) (map[string][2]float64, map[string]map[string][2]float64) {
	races := map[string]struct{}{}
	cook_ratings, incumbents, cook_pvis := LoadCookGovRatingsHtml(0)
	for k := range cook_ratings {
		races[k] = struct{}{}
	}

	expert_ratings := []map[string][2]float64{cook_ratings}
	expert_weights := []float64{COOK_WEIGHT}

	polling_data_538 := polls_538_gov
	f, err := os.Open("extra_gov_polls.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var extra_polls map[string][]Poll
	dec := json.NewDecoder(f)
	dec.Decode(&extra_polls)
	polling_data := []map[string][]Poll{polling_data_538, extra_polls}

	pvi_estimates := []map[string]float64{cook_pvis, past_gov_pvi}
	pvi_weights := []float64{COOK_PVI_WEIGHT, PAST_PVI_WEIGHT}

	unchallenged_races := [2]map[string]struct{}{nil, nil}

	gov_polls = CombinePolls(polling_data)
	polling_ratings := GetPollingRatings(gov_polls, election_date)
	var fundamentals_ratings map[string][2]float64
	fundamentals_ratings, gov_fundamentals = GetFundamentalsRatings(incumbents, nil, pvi_estimates, pvi_weights, []string{"Cook", "Historical"}, congressional_ballot)
	var experts_ratings map[string][2]float64
	experts_ratings, gov_experts = CombineExpertRatings(expert_ratings, expert_weights, []string{"Cook"})
	ratings := map[string]map[string][2]float64{"polling": polling_ratings, "fundamentals": fundamentals_ratings, "experts": experts_ratings}

	return MergeRaces(races, unchallenged_races, ratings, election_date, now)
}

func forecast() {
	setup()
	now := time.Now()
	days := election_date.Sub(now).Hours() / 24

	senate_races, senate_sources := merge_senate_classes(now)
	house_races, house_sources := merge_house_classes(now)
	gov_races, gov_sources := merge_gov_classes(now)

	fmt.Println("Senate:", len(senate_races))
	senate_odds, senate_races_map := Prob(senate_races, days, 1.0)
	senate_odds = append(append(make([]float64, 23), senate_odds...), make([]float64, 42)...) // Shift by the predetermined seats

	time.Sleep(5 * time.Second)

	fmt.Println("House:", len(house_races))
	house_odds, house_races_map := Prob(house_races, days, 1.0)

	time.Sleep(5 * time.Second)

	fmt.Println("Gov:", len(gov_races))
	gov_odds, gov_races_map := Prob(gov_races, days, 1.0)
	gov_odds = append(append(make([]float64, 7), gov_odds...), make([]float64, 7)...) // Shift by the predetermined seats

	time.Sleep(5 * time.Second)

	senate_raceprobs := make(map[string]RaceProbability)
	for k, v := range senate_races_map {
		(&v).Fix(senate_sources[k])
		senate_raceprobs[k] = v
	}

	house_raceprobs := make(map[string]RaceProbability)
	for k, v := range house_races_map {
		(&v).Fix(house_sources[k])
		house_raceprobs[k] = v
	}

	gov_raceprobs := make(map[string]RaceProbability)
	for k, v := range gov_races_map {
		(&v).Fix(gov_sources[k])
		gov_raceprobs[k] = v
	}

	err := os.MkdirAll("forecast", 0755)
	if err != nil {
		panic(err)
	}

	SavePolls("forecast/senate_polls.json", senate_polls)
	SavePolls("forecast/house_polls.json", house_polls)
	SavePolls("forecast/gov_polls.json", gov_polls)

	SaveExperts("forecast/senate_experts.json", senate_experts)
	SaveExperts("forecast/house_experts.json", house_experts)
	SaveExperts("forecast/gov_experts.json", gov_experts)

	SaveFundamentals("forecast/senate_fundamentals.json", senate_fundamentals)
	SaveFundamentals("forecast/house_fundamentals.json", house_fundamentals)
	SaveFundamentals("forecast/gov_fundamentals.json", gov_fundamentals)

	var senate_past Forecast
	func() {
		defer func() { recover() }()
		LoadForecast("senate", &senate_past)
	}()
	SaveForecast("forecast/senate_forecast.json", GenForecast(senate_odds, senate_raceprobs, senate_past, now))

	var house_past Forecast
	func() {
		defer func() { recover() }()
		LoadForecast("house", &house_past)
	}()
	SaveForecast("forecast/house_forecast.json", GenForecast(house_odds, house_raceprobs, house_past, now))

	var gov_past Forecast
	func() {
		defer func() { recover() }()
		LoadForecast("gov", &gov_past)
	}()
	SaveForecast("forecast/gov_forecast.json", GenForecast(gov_odds, gov_raceprobs, gov_past, now))
}

func main() {
	//OptimizeHistorical() // Don't call this because it's already been done
	forecast()
	if len(os.Args) > 1 && os.Args[1] == "once" {
		return
	}
	for range time.Tick(1 * time.Hour) {
		forecast()
	}
}
