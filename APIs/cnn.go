package APIs

import (
	"encoding/json"
	"fmt"
)

func LoadCNNHouseResults(year string) map[string][2]float64 {
	results := make(map[string][2]float64)
	for page := 0; page < 14; page++ {
		url := fmt.Sprintf("http://data.cnn.com/ELECTION/%s/full/H%02d.json", year, page+1)
		cache := fmt.Sprintf("cache/CNN_%s_House_%d.json", year, page+1)
		r := LoadCache(url, cache, -1)
		dec := json.NewDecoder(r)
		var data map[string]interface{}
		dec.Decode(&data)
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
	r := LoadCache(url, cache, -1)
	dec := json.NewDecoder(r)
	var data map[string]interface{}
	dec.Decode(&data)
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
	r := LoadCache(url, cache, -1)
	dec := json.NewDecoder(r)
	var data map[string]interface{}
	dec.Decode(&data)
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
