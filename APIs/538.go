package APIs

import (
	. "MidtermForecast/Utils"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

var (
	USStateAbbrev = map[string]string{
		"Alabama":        "AL",
		"Alaska":         "AK",
		"Arizona":        "AZ",
		"Arkansas":       "AR",
		"California":     "CA",
		"Colorado":       "CO",
		"Connecticut":    "CT",
		"Delaware":       "DE",
		"Florida":        "FL",
		"Georgia":        "GA",
		"Hawaii":         "HI",
		"Idaho":          "ID",
		"Illinois":       "IL",
		"Indiana":        "IN",
		"Iowa":           "IA",
		"Kansas":         "KS",
		"Kentucky":       "KY",
		"Louisiana":      "LA",
		"Maine":          "ME",
		"Maryland":       "MD",
		"Massachusetts":  "MA",
		"Michigan":       "MI",
		"Minnesota":      "MN",
		"Mississippi":    "MS",
		"Missouri":       "MO",
		"Montana":        "MT",
		"Nebraska":       "NE",
		"Nevada":         "NV",
		"New Hampshire":  "NH",
		"New Jersey":     "NJ",
		"New Mexico":     "NM",
		"New York":       "NY",
		"North Carolina": "NC",
		"North Dakota":   "ND",
		"Ohio":           "OH",
		"Oklahoma":       "OK",
		"Oregon":         "OR",
		"Pennsylvania":   "PA",
		"Rhode Island":   "RI",
		"South Carolina": "SC",
		"South Dakota":   "SD",
		"Tennessee":      "TN",
		"Texas":          "TX",
		"Utah":           "UT",
		"Vermont":        "VT",
		"Virginia":       "VA",
		"Washington":     "WA",
		"West Virginia":  "WV",
		"Wisconsin":      "WI",
		"Wyoming":        "WY",
	}
	pollsterAliases = map[string]string{
		"Siena College/New York Times": "Siena College",
		"CNN/SSRS":                     "SSRS",
		"Fox News":                     "Fox News/Anderson Robbins Research/Shaw & Co. Research",
	}
)

func get_st_dist(st string, dist int64) string {
	st_dist := st
	if st_dist == "AK" || st_dist == "DE" || st_dist == "MT" || st_dist == "ND" || st_dist == "SD" || st_dist == "VT" || st_dist == "WY" {
		st_dist += "-00"
	} else {
		st_dist += fmt.Sprintf("-%02d", dist)
	}
	return st_dist
}

func Load538Polls() (polls_538_senate, polls_538_house, polls_538_gov map[string][]Poll) {
	polls_538_senate = map[string][]Poll{}
	polls_538_house = map[string][]Poll{}
	polls_538_gov = map[string][]Poll{}
	data := LoadCache("https://projects.fivethirtyeight.com/polls/polls.json", "cache/538_polls.json", 20*time.Minute, func(r io.Reader) interface{} {
		dec := json.NewDecoder(r)
		var data []map[string]interface{}
		if err := dec.Decode(&data); err != nil {
			panic(err)
		}
		return data
	}).([]map[string]interface{})
	for _, p := range data {
		if p["type"].(string) != "senate" && p["type"].(string) != "house" && p["type"].(string) != "governor" {
			continue
		}
		var poll Poll
		poll.Pollster = p["pollster"].(string)
		if p, ok := pollsterAliases[poll.Pollster]; ok {
			poll.Pollster = p
		}
		start, err := time.Parse("2006-01-02", p["startDate"].(string))
		if err != nil {
			panic(err)
		}
		poll.URL = p["url"].(string)
		poll.StartDate = start
		end, err := time.Parse("2006-01-02", p["endDate"].(string))
		if err != nil {
			panic(err)
		}
		poll.EndDate = end
		poll.Subpopulation = "A"
		if p["population"] != nil {
			poll.Subpopulation = strings.ToUpper(p["population"].(string))
		}
		if p["sampleSize"] == nil {
			continue
		}
		num, err := strconv.ParseFloat(p["sampleSize"].(string), 64)
		if err != nil {
			panic(err)
		}
		poll.Number = num
		candidates := map[string]float64{}
		switch p["type"].(string) {
		case "senate":
			st := USStateAbbrev[p["state"].(string)]
			second := false // Second senate election in the state
			for _, x := range p["answers"].([]interface{}) {
				v := x.(map[string]interface{})
				pct, err := strconv.ParseFloat(v["pct"].(string), 64)
				if err != nil {
					panic(err)
				}
				if v["party"].(string) == "Dem" {
					if v["choice"].(string) == "Klobuchar" {
						second = true
					} else if v["choice"].(string) == "Espy" {
						second = true
					}
					if pct/100 > candidates["D"] {
						candidates["D"] = pct / 100
					}
				} else if v["party"].(string) == "Rep" {
					if v["choice"] == "King" && st == "ME" {
						// 538 fail
						if pct/100 > candidates["D"] {
							candidates["D"] = pct / 100
						}
					} else {
						if pct/100 > candidates["R"] {
							candidates["R"] = pct / 100
						}
					}
				} else if v["party"].(string) == "Ind" {
					if (v["choice"] == "King" && st == "ME") || (v["choice"] == "Sanders" && st == "VT") {
						if pct/100 > candidates["D"] {
							candidates["D"] = pct / 100
						}
					}
				}
			}
			poll.Candidates = candidates
			if second {
				st += "-2"
			}
			if _, ok := polls_538_senate[st]; !ok {
				polls_538_senate[st] = make([]Poll, 0)
			}
			polls_538_senate[st] = append(polls_538_senate[st], poll)
		case "house":
			district, err := strconv.ParseInt(p["district"].(string), 10, 64)
			if err != nil {
				panic(err)
			}
			st := get_st_dist(USStateAbbrev[p["state"].(string)], district)
			for _, x := range p["answers"].([]interface{}) {
				v := x.(map[string]interface{})
				pct, err := strconv.ParseFloat(v["pct"].(string), 64)
				if err != nil {
					panic(err)
				}
				if v["party"].(string) == "Dem" {
					if pct/100 > candidates["D"] {
						candidates["D"] = pct / 100
					}
				} else if v["party"].(string) == "Rep" {
					if pct/100 > candidates["R"] {
						candidates["R"] = pct / 100
					}
				}
			}
			poll.Candidates = candidates
			if _, ok := polls_538_house[st]; !ok {
				polls_538_house[st] = make([]Poll, 0)
			}
			polls_538_house[st] = append(polls_538_house[st], poll)
		case "governor":
			st := USStateAbbrev[p["state"].(string)]
			for _, x := range p["answers"].([]interface{}) {
				v := x.(map[string]interface{})
				pct, err := strconv.ParseFloat(v["pct"].(string), 64)
				if err != nil {
					panic(err)
				}
				if v["party"].(string) == "Dem" {
					if pct/100 > candidates["D"] {
						candidates["D"] = pct / 100
					}
				} else if v["party"].(string) == "Rep" {
					if pct/100 > candidates["R"] {
						candidates["R"] = pct / 100
					}
				}
			}
			poll.Candidates = candidates
			if _, ok := polls_538_gov[st]; !ok {
				polls_538_gov[st] = make([]Poll, 0)
			}
			polls_538_gov[st] = append(polls_538_gov[st], poll)
		}
	}
	return
}

type GenericBallot538 struct {
	Date     string             `json:"date"`
	Subgroup string             `json:"subgroup"`
	Revised  GenericEstimate538 `json:"revised"`
	Original GenericEstimate538 `json"original"`
}

type GenericEstimate538 struct {
	DemEstimate float64 `json:"dem_estimate"`
	DemHi       float64 `json:"dem_hi"`
	DemLo       float64 `json:"dem_lo"`
	RepEstimate float64 `json:"rep_estimate"`
	RepHi       float64 `json:"rep_hi"`
	RepLo       float64 `json:"rep_lo"`
}

func Load538GenericBallot() float64 {
	v := LoadCache("https://projects.fivethirtyeight.com/congress-generic-ballot-polls/generic.json", "cache/538_generic.json", 20*time.Minute, func(r io.Reader) interface{} {
		dec := json.NewDecoder(r)
		var v []GenericBallot538
		if err := dec.Decode(&v); err != nil {
			panic(err)
		}
		return v
	}).([]GenericBallot538)
	dem := v[len(v)-1].Revised.DemEstimate
	rep := v[len(v)-1].Revised.RepEstimate
	return (dem - rep) / 100.0
}

type Forecast538 struct {
	//CongressPartySplits []CongressPartySplit538 `json:"congress_party_splits"`
	//SeatChances         []SeatChance538         `json:"seat_chances"`
	//NationalTrends      []NationalTrend538      `json:"nationalTrends"`
	DistrictForecasts []DistrictForecast538 `json:"districtForecasts"` // House
	SeatForecasts     []SeatForecast538     `json:"seatForecasts"`     // Senate
}

type DistrictForecast538 struct {
	State    string                 `json:"state"`
	District string                 `json:"district"`
	Forecast []CandidateForecast538 `json:"forecast"`
	//Incumbents []Incumbent `json:"incumbent"
}

type SeatForecast538 struct {
	State    string                 `json:"state"`
	Class    int                    `json:"class"`
	Forecast []CandidateForecast538 `json:"forecast"`
	//Incumbents []Incumbent `json:"incumbent"
}

type CandidateForecast538 struct {
	Candidate string            `json:"candidate"`
	Party     string            `json:"party"`
	Date      string            `json:"date"`
	Models    ModelForecasts538 `json:"models"`
}

type ModelForecasts538 struct {
	Lite    ModelForecast538 `json:"lite"`
	Classic ModelForecast538 `json:"classic"`
	Deluxe  ModelForecast538 `json:"deluxe"`
}

type ModelForecast538 struct {
	WinProb     float64 `json:"winprob"`
	VoteShare   float64 `json:"voteshare"`
	VoteShareHi float64 `json:"voteshare_hi"`
	VoteShareLo float64 `json:"voteshare_lo"`
	Margin      float64 `json:"margin"` // only for US
}

func Load538HouseForecast() (forecast_538 map[string]float64, parties_538 map[string]map[string]string, congressional_ballot float64) {
	forecast_538 = map[string]float64{}
	parties_538 = map[string]map[string]string{}
	data := LoadCache("https://projects.fivethirtyeight.com/2018-midterm-election-forecast/house/home.json", "cache/538_house.json", 20*time.Minute, func(r io.Reader) interface{} {
		dec := json.NewDecoder(r)
		var data Forecast538
		if err := dec.Decode(&data); err != nil {
			panic(err)
		}
		return data
	}).(Forecast538)
	for _, d := range data.DistrictForecasts {
		st := d.State
		if st == "US" {
			for _, c := range d.Forecast {
				if c.Party == "D" {
					congressional_ballot = (c.Models.Lite.Margin*0.15 + c.Models.Classic.Margin*0.35 + c.Models.Deluxe.Margin*0.5) / 100.0
				}
			}
			continue
		}
		district, err := strconv.ParseInt(d.District, 10, 64)
		if err != nil {
			panic(err)
		}
		st = get_st_dist(d.State, district)
		parties_538[st] = map[string]string{}
		voteshares := map[string]float64{}
		for _, c := range d.Forecast {
			parties_538[st][c.Candidate] = c.Party
			//winProb := c.Models.Lite.WinProb*0.15 + c.Models.Classic.WinProb*0.35 + c.Models.Deluxe.WinProb*0.5
			voteShare := c.Models.Lite.VoteShare*0.15 + c.Models.Classic.VoteShare*0.35 + c.Models.Deluxe.VoteShare*0.5
			voteshares[c.Party] += voteShare
		}
		pvi := (voteshares["D"] - voteshares["R"]) / (voteshares["D"] + voteshares["R"])
		forecast_538[st] = pvi
	}
	return
}

func Load538SenateForecast() (forecast_538 map[string]float64, parties_538 map[string]map[string]string) {
	forecast_538 = map[string]float64{}
	parties_538 = map[string]map[string]string{}
	data := LoadCache("https://projects.fivethirtyeight.com/2018-midterm-election-forecast/senate/home.json", "cache/538_senate.json", 20*time.Minute, func(r io.Reader) interface{} {
		dec := json.NewDecoder(r)
		var data Forecast538
		if err := dec.Decode(&data); err != nil {
			panic(err)
		}
		return data
	}).(Forecast538)
	for _, d := range data.SeatForecasts {
		st := d.State
		if st == "US" {
			continue
		}
		if d.Class == 2 {
			st += "-2"
		}
		parties_538[st] = map[string]string{}
		voteshares := map[string]float64{}
		for _, c := range d.Forecast {
			if c.Party == "I" && c.Candidate == "Bernard Sanders" || c.Candidate == "Angus S. King Jr." {
				c.Party = "D"
			}
			parties_538[st][c.Candidate] = c.Party
			//winProb := c.Models.Lite.WinProb*0.15 + c.Models.Classic.WinProb*0.35 + c.Models.Deluxe.WinProb*0.5
			voteShare := c.Models.Lite.VoteShare*0.15 + c.Models.Classic.VoteShare*0.35 + c.Models.Deluxe.VoteShare*0.5
			voteshares[c.Party] += voteShare
		}
		pvi := (voteshares["D"] - voteshares["R"]) / (voteshares["D"] + voteshares["R"])
		forecast_538[st] = pvi
	}
	return
}

func Load538GovForecast() (forecast_538 map[string]float64, parties_538 map[string]map[string]string) {
	forecast_538 = map[string]float64{}
	parties_538 = map[string]map[string]string{}
	data := LoadCache("https://projects.fivethirtyeight.com/2018-midterm-election-forecast/governor/home.json", "cache/538_governor.json", 20*time.Minute, func(r io.Reader) interface{} {
		dec := json.NewDecoder(r)
		var data Forecast538
		if err := dec.Decode(&data); err != nil {
			panic(err)
		}
		return data
	}).(Forecast538)
	for _, d := range data.SeatForecasts {
		st := d.State
		if st == "US" {
			continue
		}
		parties_538[st] = map[string]string{}
		voteshares := map[string]float64{}
		for _, c := range d.Forecast {
			if c.Party == "U" && c.Candidate == "Bill Walker" {
				c.Party = "D"
			}
			parties_538[st][c.Candidate] = c.Party
			//winProb := c.Models.Lite.WinProb*0.15 + c.Models.Classic.WinProb*0.35 + c.Models.Deluxe.WinProb*0.5
			voteShare := c.Models.Lite.VoteShare*0.15 + c.Models.Classic.VoteShare*0.35 + c.Models.Deluxe.VoteShare*0.5
			voteshares[c.Party] += voteShare
		}
		pvi := (voteshares["D"] - voteshares["R"]) / (voteshares["D"] + voteshares["R"])
		forecast_538[st] = pvi
	}
	return
}

func Load2016SenatePolls() map[string][]Poll {
	data := LoadCache("https://projects.fivethirtyeight.com/2016-election-forecast/senate/updates.json", "cache/538_Senate_polls.json", -1, func(r io.Reader) interface{} {
		dec := json.NewDecoder(r)
		var data []map[string]interface{}
		if err := dec.Decode(&data); err != nil {
			panic(err)
		}
		return data
	}).([]map[string]interface{})
	var polls = map[string][]Poll{}
	for _, p := range data {
		var poll Poll
		poll.Pollster = p["pollster"].(string)
		if p, ok := pollsterAliases[poll.Pollster]; ok {
			poll.Pollster = p
		}
		start, err := time.Parse("2006-01-02", p["startDate"].(string))
		if err != nil {
			panic(err)
		}
		poll.StartDate = start
		end, err := time.Parse("2006-01-02", p["endDate"].(string))
		if err != nil {
			panic(err)
		}
		poll.EndDate = end
		poll.Subpopulation = strings.ToUpper(p["population"].(string))
		if p["sampleSize"] == nil {
			continue
		}
		num, err := strconv.ParseFloat(p["sampleSize"].(string), 64)
		if err != nil {
			panic(err)
		}
		poll.Number = num
		st := p["state"].(string)
		if st == "USA" || st == "US" {
			continue
		}
		candidates := map[string]float64{}
		for _, x := range p["votingAnswers"].([]interface{}) {
			v := x.(map[string]interface{})
			candidates[v["party"].(string)] += v["pct"].(float64) / 100
		}
		poll.Candidates = candidates
		if _, ok := polls[st]; !ok {
			polls[st] = make([]Poll, 0)
		}
		polls[st] = append(polls[st], poll)
	}
	return polls
}

func Load538HouseElasticities() map[string]float64 {
	r := LoadCache("https://raw.githubusercontent.com/fivethirtyeight/data/master/political-elasticity-scores/elasticity-by-district.csv", "cache/elasticity-by-district.csv", 48*time.Hour, func(r io.Reader) interface{} {
		if b, err := ioutil.ReadAll(r); err == nil {
			return b
		} else {
			panic(err)
		}
	}).([]byte)
	reader := csv.NewReader(bytes.NewReader(r))
	reader.Read() // Skip header line
	elasticities := make(map[string]float64)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		district := record[0]
		st := district[:2]
		dist, err := strconv.Atoi(district[3:])
		if err != nil {
			panic(err)
		}
		st = get_st_dist(st, int64(dist))
		elasticity, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			panic(err)
		}
		elasticities[st] = elasticity
	}
	return elasticities
}

func Load538SenateElasticities() map[string]float64 {
	r := LoadCache("https://raw.githubusercontent.com/fivethirtyeight/data/master/political-elasticity-scores/elasticity-by-state.csv", "cache/elasticity-by-state.csv", 48*time.Hour, func(r io.Reader) interface{} {
		if b, err := ioutil.ReadAll(r); err == nil {
			return b
		} else {
			panic(err)
		}
	}).([]byte)
	reader := csv.NewReader(bytes.NewReader(r))
	reader.Read() // Skip header line
	elasticities := make(map[string]float64)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		st := record[0]
		elasticity, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			panic(err)
		}
		elasticities[st] = elasticity
		elasticities[st+"-2"] = elasticity
	}
	return elasticities
}
