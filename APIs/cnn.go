package APIs

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

func LoadCNNHouseResults(year string) map[string][2]float64 {
	results := make(map[string][2]float64)
	for page := 0; page < 14; page++ {
		url := fmt.Sprintf("http://data.cnn.com/ELECTION/%s/full/H%02d.json", year, page+1)
		cache := fmt.Sprintf("cache/CNN_%s_House_%d.json", year, page+1)
		data := LoadCache(url, cache, -1, func(r io.Reader) interface{} {
			dec := json.NewDecoder(r)
			var data map[string]interface{}
			if err := dec.Decode(&data); err != nil {
				panic(err)
			}
			return data
		}).(map[string]interface{})
		for _, raw_race := range data["races"].([]interface{}) {
			race := raw_race.(map[string]interface{})
			raceid := race["raceid"].(string)
			seat := raceid[:2] + "-" + raceid[3:]
			if seat[:2] == "US" {
				continue
			}
			if seat == "AK-01" || seat == "DE-01" || seat == "MT-01" || seat == "ND-01" || seat == "SD-01" || seat == "VT-01" || seat == "WY-01" {
				seat = seat[:3] + "00"
			}
			var d_votes, r_votes float64
			for _, raw_cand := range race["candidates"].([]interface{}) {
				cand := raw_cand.(map[string]interface{})
				party := cand["party"].(string)
				votes := cand["votes"].(float64)
				if party == "R" {
					r_votes += votes
				} else if party == "D" {
					d_votes += votes
				}
			}
			if r_votes == 0 && d_votes == 0 {
				if len(race["candidates"].([]interface{})) != 1 {
					fmt.Println(race)
					panic(nil)
				}
				if race["candidates"].([]interface{})[0].(map[string]interface{})["party"].(string) == "R" {
					r_votes = 1
				} else {
					d_votes = 1
				}
			}
			results[seat] = [2]float64{d_votes, r_votes}
		}
	}
	return results
}

func LoadCNNSenateResults(year string) map[string][2]float64 {
	results := make(map[string][2]float64)
	url := fmt.Sprintf("http://data.cnn.com/ELECTION/%s/full/S.full.json", year)
	cache := fmt.Sprintf("cache/CNN_%s_Senate.json", year)
	data := LoadCache(url, cache, -1, func(r io.Reader) interface{} {
		dec := json.NewDecoder(r)
		var data map[string]interface{}
		if err := dec.Decode(&data); err != nil {
			panic(err)
		}
		return data
	}).(map[string]interface{})
	for _, raw_race := range data["races"].([]interface{}) {
		race := raw_race.(map[string]interface{})
		raceid := race["raceid"].(string)
		seat := raceid[:2]
		if raceid[3:] == "02" {
			seat += "-2"
		}
		if seat == "US" {
			continue
		}
		var d_votes, r_votes float64
		for _, raw_cand := range race["candidates"].([]interface{}) {
			cand := raw_cand.(map[string]interface{})
			party := cand["party"].(string)
			votes := cand["votes"].(float64)
			if party == "R" {
				r_votes += votes
			} else if party == "D" || (party == "I" && (cand["lname"] == "Sanders" || cand["fname"] == "Angus" || cand["lname"] == "Orman")) {
				d_votes += votes
			}
		}
		if r_votes == 0 && d_votes == 0 {
			if len(race["candidates"].([]interface{})) != 1 {
				fmt.Println(race)
				panic(nil)
			}
			if race["candidates"].([]interface{})[0].(map[string]interface{})["party"].(string) == "R" {
				r_votes = 1
			} else {
				d_votes = 1
			}
		}
		results[seat] = [2]float64{d_votes, r_votes}
	}
	return results
}

func LoadCNNGovResults(year string) map[string][2]float64 {
	results := make(map[string][2]float64)
	url := fmt.Sprintf("http://data.cnn.com/ELECTION/%s/full/G.full.json", year)
	cache := fmt.Sprintf("cache/CNN_%s_Gov.json", year)
	data := LoadCache(url, cache, -1, func(r io.Reader) interface{} {
		dec := json.NewDecoder(r)
		var data map[string]interface{}
		if err := dec.Decode(&data); err != nil {
			panic(err)
		}
		return data
	}).(map[string]interface{})
	for _, raw_race := range data["races"].([]interface{}) {
		race := raw_race.(map[string]interface{})
		raceid := race["raceid"].(string)
		seat := raceid[:2]
		if seat == "US" {
			continue
		}
		var d_votes, r_votes float64
		for _, raw_cand := range race["candidates"].([]interface{}) {
			cand := raw_cand.(map[string]interface{})
			party := cand["party"].(string)
			votes := cand["votes"].(float64)
			if party == "R" {
				r_votes += votes
			} else if party == "D" || (party == "O" && (cand["lname"] == "Walker")) {
				d_votes += votes
			}
		}
		if r_votes == 0 && d_votes == 0 {
			if len(race["candidates"].([]interface{})) != 1 {
				fmt.Println(race)
				panic(nil)
			}
			if race["candidates"].([]interface{})[0].(map[string]interface{})["party"].(string) == "R" {
				r_votes = 1
			} else {
				d_votes = 1
			}
		}
		results[seat] = [2]float64{d_votes, r_votes}
	}
	return results
}

type CNNForecast struct {
	Topline        CNNForecastTopline       `json:"topline"`
	Competitive    []CNNForecastRaceDetails `json:"competitive"`
	Swing          []CNNForecastRaceDetails `json:"swing"`
	Districts      []CNNForecastDistrict    `json:"districts"`
	Seats          []CNNForecastSeat        `json:"seats"`
	LastUpdatedEDT string                   `json:"lastUpdatedEDT"`
}

type CNNForecastTopline struct {
	Timestamp   string                      `json:"timestamp"`
	Party       string                      `json:"party"`
	MarginHigh  float64                     `json:"marginHigh"`
	MarginWin   float64                     `json:"marginWin"`
	MarginLow   float64                     `json:"marginLow"`
	DaysOut     int                         `json:"daysOut"`
	Race        string                      `json:"race"`
	Predictions []CNNForecastPrediction     `json:"predictions"`
	History     []CNNForecastToplineHistory `json:"history"`
}

type CNNForecastHistory struct {
	Timestamp  string  `json:"timestamp"`
	Party      string  `json:"party"`
	MarginHigh float64 `json:"marginHigh"`
	MarginWin  float64 `json:"marginWin"`
	MarginLow  float64 `json:"marginLow"`
}

type CNNForecastToplineHistory struct {
	CNNForecastHistory
	DaysOut int `json:"daysOut"`
}

type CNNForecastPrediction struct {
	Bin int     `json:"bin"`
	Pct float64 `json:"pct"`
}

type CNNForecastRaceDetails struct {
	ID                              string                  `json:"id"`
	State                           string                  `json:"state"`
	District                        string                  `json:"district"`
	Prediction                      float64                 `json:"prediction"`
	LowMargin                       float64                 `json:"lowMargin"`
	HiMargin                        float64                 `json:"hiMargin"`
	Predictions                     []CNNForecastPrediction `json:"predictions"`
	LastHouseResultDemocraticMargin float64                 `json:"last_house_result_democratic_margin"`
	UnopposedPrior                  bool                    `json:"unopposedPrior"`
	ExcludedCurrent                 bool                    `json:"excludedCurrent"`
}

type CNNForecastDistrict struct {
	ID         string  `json:"id"`
	State      string  `json:"state"`
	District   string  `json:"district"`
	Prediction float64 `json:"prediction"`
	LowMargin  float64 `json:"lowMargin"`
	HiMargin   float64 `json:"hiMargin"`
	Rcand      string  `json:"rcand,omitempty"`
	Rlast      string  `json:"rlast,omitempty"`
	Dcand      string  `json:"dcand,omitempty"`
	Dlast      string  `json:"dlast,omitempty"`
}

type CNNForecastSeat struct {
	Seat        string                  `json:"seat"`
	LowMargin   float64                 `json:"lowMargin"`
	Prediction  float64                 `json:"prediction"`
	HiMargin    float64                 `json:"hiMargin"`
	History     []CNNForecastHistory    `json:"history"`
	Predictions []CNNForecastPrediction `json:"predictions"`
	Dcand       string                  `json:"dcand,omitempty"`
	Dlast       string                  `json:"dlast,omitempty"`
	Rcand       string                  `json:"rcand,omitempty"`
	Rlast       string                  `json:"rlast,omitempty"`
	Dcand2      string                  `json:"dcand2"`
	Icand       string                  `json:"icand,omitempty"`
	Ilast       string                  `json:"ilast,omitempty"`
}

func LoadCNNSenateForecast() map[string]float64 {
	data := LoadCache("https://data.cnn.com/interactive/2018/stage/senate/overview.json", "cache/CNN_senate_forecast.json", 20*time.Minute, func(r io.Reader) interface{} {
		dec := json.NewDecoder(r)
		var data CNNForecast
		if err := dec.Decode(&data); err != nil {
			panic(err)
		}
		return data
	}).(CNNForecast)
	r := make(map[string]float64, len(data.Seats))
	for _, district := range data.Seats {
		st := USStateAbbrev[district.Seat]
		if strings.HasSuffix(district.Seat, " (Special)") {
			st = USStateAbbrev[district.Seat[:len(district.Seat)-10]] + "-2"
		}
		r[st] = district.Prediction / 100.0
	}
	return r
}

func LoadCNNHouseForecast() map[string]float64 {
	data := LoadCache("https://data.cnn.com/interactive/2018/stage/house/overview.json", "cache/CNN_house_forecast.json", 20*time.Minute, func(r io.Reader) interface{} {
		dec := json.NewDecoder(r)
		var data CNNForecast
		if err := dec.Decode(&data); err != nil {
			panic(err)
		}
		return data
	}).(CNNForecast)
	r := make(map[string]float64, len(data.Districts))
	for _, district := range data.Districts {
		st := district.State
		dist, err := strconv.Atoi(district.District)
		if err != nil {
			panic(err)
		}
		st = get_st_dist(st, int64(dist))
		r[st] = district.Prediction / 100.0
	}
	return r
}
