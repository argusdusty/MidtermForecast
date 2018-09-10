package APIs

import (
	"compress/zlib"
	"io/ioutil"
	"strconv"
	"strings"
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
	r := LoadCache(url, "cache/NYT_"+year+"_Senate.tsv", -1)
	var s string
	if year == "2008" {
		zr, err := zlib.NewReader(r)
		if err != nil {
			panic(err)
		}
		b, err := ioutil.ReadAll(zr)
		if err != nil {
			panic(err)
		}
		s = string(b)
	} else if year == "2010" {
		b, err := ioutil.ReadAll(r)
		if err != nil {
			panic(err)
		}
		s = string(b)
	}
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
	r := LoadCache(url, "cache/NYT_"+year+"_House.tsv", -1)
	var s string
	if year == "2008" {
		zr, err := zlib.NewReader(r)
		if err != nil {
			panic(err)
		}
		b, err := ioutil.ReadAll(zr)
		if err != nil {
			panic(err)
		}
		s = string(b)
	} else if year == "2010" {
		b, err := ioutil.ReadAll(r)
		if err != nil {
			panic(err)
		}
		s = string(b)
	}
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
