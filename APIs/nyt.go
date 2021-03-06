package APIs

import (
	. "MidtermForecast/Utils"
	"compress/zlib"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// 2008, 2010
func LoadNYTSenateResults(year string) map[string][2]float64 {
	var url string
	if year == "2008" {
		url = "https://static01.nyt.com/packages/xml/1min/election_results/senate.txt"
	} else if year == "2010" {
		url = "https://int.nyt.com/data/elections/2010/results/senate.tsv"
	} else {
		return nil
	}
	s := LoadCache(url, "cache/NYT_"+year+"_Senate.tsv", -1, func(r io.Reader) interface{} {
		if year == "2008" {
			zr, err := zlib.NewReader(r)
			if err != nil {
				panic(err)
			}
			b, err := ioutil.ReadAll(zr)
			if err != nil {
				panic(err)
			}
			return string(b)
		} else if year == "2010" {
			b, err := ioutil.ReadAll(r)
			if err != nil {
				panic(err)
			}
			return string(b)
		} else {
			return ""
		}
	}).(string)
	results := make(map[string][2]float64)
	parties := make(map[string][]string)
	var lines []string
	if year == "2008" {
		lines = strings.Split(s, "\n")[47:119]
	} else if year == "2010" {
		lines = strings.Split(s, "\n")[9:173]
	}
	for _, line := range lines {
		vals := strings.Split(line, "\t")
		st := vals[0]
		if len(st) == 3 {
			st = st[:2] + "-" + st[2:]
		}
		party := vals[2]
		votes, err := strconv.ParseInt(vals[3], 10, 64)
		if err != nil {
			panic(err)
		}
		r := results[st]
		if party == "Dem" {
			r[0] += float64(votes)
		} else if party == "GOP" || (party == "NPA" && vals[2] == "Murkowski") {
			r[1] += float64(votes)
		}
		parties[st] = append(parties[st], party)
		results[st] = r
	}
	for st, r := range results {
		if r[0] == 0 && r[1] == 0 {
			if parties[st][0] == "Dem" {
				r[0] = 1
			} else if parties[st][0] == "GOP" {
				r[1] = 1
			}
			results[st] = r
		}
	}
	return results
}

func LoadNYTHouseResults(year string) map[string][2]float64 {
	var url string
	if year == "2008" {
		url = "https://static01.nyt.com/packages/xml/1min/election_results/house.txt"
	} else if year == "2010" {
		url = "https://int.nyt.com/data/elections/2010/results/house.tsv"
	} else {
		return nil
	}
	s := LoadCache(url, "cache/NYT_"+year+"_House.tsv", -1, func(r io.Reader) interface{} {
		if year == "2008" {
			zr, err := zlib.NewReader(r)
			if err != nil {
				panic(err)
			}
			b, err := ioutil.ReadAll(zr)
			if err != nil {
				panic(err)
			}
			return string(b)
		} else if year == "2010" {
			b, err := ioutil.ReadAll(r)
			if err != nil {
				panic(err)
			}
			return string(b)
		} else {
			return ""
		}
	}).(string)
	results := make(map[string][2]float64)
	parties := make(map[string][]string)
	var lines []string
	if year == "2008" {
		lines = strings.Split(s, "\n")[445:1306]
	} else if year == "2010" {
		lines = strings.Split(s, "\n")[447:1728]
	}
	for _, line := range lines {
		vals := strings.Split(line, "\t")
		st := vals[0]
		dist, err := strconv.ParseInt(vals[1], 10, 64)
		if err != nil {
			panic(err)
		}
		st = get_st_dist(st, dist)
		party := vals[3]
		votes, err := strconv.ParseInt(vals[4], 10, 64)
		if err != nil {
			panic(err)
		}
		r := results[st]
		if party == "Dem" {
			r[0] += float64(votes)
		} else if party == "GOP" {
			r[1] += float64(votes)
		}
		parties[st] = append(parties[st], party)
		results[st] = r
	}
	for st, r := range results {
		if r[0] == 0 && r[1] == 0 {
			if parties[st][0] == "Dem" {
				r[0] = 1
			} else if parties[st][0] == "GOP" {
				r[1] = 1
			}
			results[st] = r
		}
	}
	return results
}

func parseTime(value string) time.Time {
	// e.g. Sept. 6, Oct. 10
	vals := strings.Split(value, " ")
	month := map[string]time.Month{
		"Aug.":  time.August,
		"Sept.": time.September,
		"Oct.":  time.October,
		"Nov.":  time.November,
		"Dec.":  time.December,
	}[vals[0]]
	day, err := strconv.Atoi(vals[1])
	if err != nil {
		panic(err)
	}
	return time.Date(2018, month, day, 0, 0, 0, 0, time.UTC)
}

// Load ongoing polls from https://www.nytimes.com/interactive/2018/upshot/elections-polls.html
// Partial data is better than no data
func LoadNYTLivePolls() (housePolls, senatePolls map[string][]Poll) {
	resp, err := http.Get("https://int.nyt.com/newsgraphics/2018/live-polls-2018/all-races.json")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var data []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		panic(err)
	}
	housePolls = make(map[string][]Poll)
	senatePolls = make(map[string][]Poll)
	for _, p := range data {
		var poll Poll
		var seat string
		if s, ok := p["id"]; ok {
			seat = s.(string)
		} else if ds, ok := p["district_id"]; ok {
			seat = ds.(string)
		} else {
			continue
		}
		if len(seat) == 4 {
			if seat[2:] == "AL" {
				// Just in case
				seat = seat[:2] + "00"
			}
			seat = seat[:2] + "-" + seat[2:]
		}
		var senate bool = false
		if strings.Contains(p["name"].(string), "Senate") {
			seat = seat[:2]
			senate = true
		}
		poll.Pollster = "Siena College"
		poll.URL = "https://www.nytimes.com/interactive/2018/upshot/" + p["page_id"].(string) + ".html"
		if p["startDate"] == nil || p["endDate"] == nil {
			continue
		}
		poll.StartDate = parseTime(p["startDate"].(string))
		poll.EndDate = parseTime(p["endDate"].(string))
		poll.Subpopulation = "LV"
		poll.Number = p["n"].(float64)
		poll.Candidates = map[string]float64{"D": p["nDem"].(float64) / poll.Number, "R": p["nRep"].(float64) / poll.Number}
		if senate {
			senatePolls[seat] = append(senatePolls[seat], poll)
		} else {
			housePolls[seat] = append(housePolls[seat], poll)
		}
	}
	return
}
