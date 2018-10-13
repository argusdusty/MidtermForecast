package APIs

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	// Pre-computed in historical_optimize.go
	COOK_WEIGHT     = 302.0708670497893
	COOK_PVI_WEIGHT = 0.3588032283134363

	CookConfidence = [][2]float64{{0.67, 0.33}, {0.56, 0.44}, {0.535, 0.465}, {0.5, 0.5}, {0.5, 0.5}, {0.465, 0.535}, {0.44, 0.56}, {0.33, 0.67}} // Based on 538's analysis of expert forecast margins
)

func LoadCookRatings(file string) (ratings []map[string]string) {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	if err := dec.Decode(&ratings); err != nil {
		panic(err)
	}
	return
}

func LoadCookSenateRatingsHtml(id uint64) (ratings map[string][2]float64, incumbents map[string]string, pvis map[string]float64) {
	var data []map[string]string
	if id == 0 { // 2018
		data = []map[string]string{
			{
				"CA": "D",
				"CT": "D",
				"DE": "D",
				"HI": "D",
				"MA": "D",
				"MD": "D",
				"ME": "D",
				"MN": "D",
				"NM": "D",
				"NY": "D",
				"RI": "D",
				"VA": "D",
				"VT": "D",
				"WA": "D",
			},
			{
				"MI": "D",
				"OH": "D",
				"PA": "D",
				"WI": "D",
			},
			{
				"MN-2": "D",
				"NJ":   "D",
				"WV":   "D",
			},
			{
				"FL": "D",
				"IN": "D",
				"MO": "D",
				"MT": "D",
				"ND": "D",
			},
			{
				"AZ": "O",
				"NV": "R",
				"TN": "O",
				"TX": "R",
			},
			{},
			{
				"MS-2": "R",
			},
			{
				"MS": "R",
				"NE": "R",
				"UT": "O",
				"WY": "R",
			},
		}
	} else if id == 139360 { // 2016
		data = []map[string]string{
			{
				"CA": "D",
				"CT": "D",
				"HI": "D",
				"MD": "O",
				"NY": "D",
				"OR": "D",
				"VT": "D",
				"WA": "D",
			},
			{
				"CO": "D",
			},
			{
				"IL": "R",
			},
			{
				"NV": "O",
			},
			{
				"IN": "O",
				"MO": "R",
				"NC": "R",
				"NH": "R",
				"PA": "R",
				"WI": "R",
			},
			{
				"AZ": "R",
				"FL": "R",
				"OH": "R",
			},
			{
				"AK": "R",
				"GA": "R",
				"IA": "R",
			},
			{
				"AL": "R",
				"AR": "R",
				"ID": "R",
				"KS": "R",
				"KY": "R",
				"LA": "R",
				"ND": "R",
				"OK": "R",
				"SC": "R",
				"SD": "R",
				"UT": "R",
			},
		}
	} else if id == 139256 { // 2014
		data = []map[string]string{
			{
				"DE": "D",
				"HI": "D",
				"IL": "O",
				"MA": "D",
				"NJ": "D",
				"NM": "D",
				"RI": "D",
			},
			{
				"MN": "D",
				"OR": "D",
				"VA": "D",
			},
			{
				"MI": "O",
			},
			{
				"AK": "D",
				"AR": "D",
				"CO": "D",
				"IA": "O",
				"LA": "D",
				"NC": "D",
				"NH": "D",
			},
			{
				"GA": "O",
				"KS": "R",
			},
			{
				"SD": "O",
				"KY": "R",
			},
			{
				"WV": "O",
				"MS": "R",
			},
			{
				"MT":   "O",
				"AL":   "R",
				"ID":   "R",
				"ME":   "R",
				"NE":   "O",
				"OK":   "R",
				"OK-2": "R",
				"SC":   "R",
				"SC-2": "R",
				"TN":   "R",
				"TX":   "R",
				"WY":   "R",
			},
		}
	} else if id == 139117 { // 2012
		data = []map[string]string{
			{
				"CA": "D",
				"DE": "D",
				"MD": "D",
				"MN": "D",
				"NY": "D",
				"RI": "D",
				"VT": "D",
				"WA": "D",
			},
			{
				"MI": "D",
				"MO": "D",
				"NJ": "D",
				"WV": "D",
			},
			{
				"FL": "D",
				"HI": "O",
				"NM": "O",
				"OH": "D",
				"PA": "D",
			},
			{
				"CT": "D",
				"MT": "D",
				"ND": "O",
				"VA": "O",
				"WI": "O",
			},
			{
				"AZ": "O",
				"IN": "O",
				"MA": "R",
				"ME": "O",
				"NV": "R",
			},
			{
				"NE": "O",
			},
			{},
			{
				"MS": "R",
				"TN": "R",
				"TX": "O",
				"UT": "R",
				"WY": "R",
			},
		}
	} else if id == 139080 { // 2008
		data = []map[string]string{
			{
				"AR": "D",
				"DE": "D",
				"IA": "D",
				"IL": "D",
				"MA": "D",
				"MI": "D",
				"MT": "D",
				"NJ": "D",
				"RI": "D",
				"WV": "D",
			},
			{
				"SD": "D",
				"VA": "R",
			},
			{
				"LA": "D",
				"AK": "R",
				"CO": "O",
				"NM": "O",
			},
			{},
			{
				"GA":   "R",
				"MN":   "R",
				"MS-2": "R",
				"NC":   "R",
				"NH":   "R",
				"OR":   "R",
			},
			{
				"KY": "R",
			},
			{
				"ME": "R",
				"NE": "O",
				"OK": "R",
			},
			{
				"AL":   "R",
				"ID":   "O",
				"KS":   "R",
				"MS":   "R",
				"SC":   "R",
				"TN":   "R",
				"TX":   "R",
				"WY":   "R",
				"WY-2": "R",
			},
		}
	}
	ratings = map[string][2]float64{}
	incumbents = map[string]string{}
	pvis = map[string]float64{}
	for idx, m := range data {
		for st, i := range m {
			if _, ok := ratings[st]; ok {
				fmt.Println(id, st)
				panic(nil)
			}
			ratings[st] = [2]float64{CookConfidence[idx][0], CookConfidence[idx][1]}
			incumbents[st] = i
			pvis[st] = CookConfidence[idx][0] - CookConfidence[idx][1]
		}
	}
	if id == 0 {
		pvis = map[string]float64{
			"AZ":   -0.10,
			"CA":   1.0, // Uncontested
			"CT":   0.12,
			"DE":   0.12,
			"FL":   -0.04,
			"HI":   0.36,
			"IN":   -0.18,
			"ME":   0.06,
			"MD":   0.24,
			"MA":   0.24,
			"MI":   0.02,
			"MN":   0.02,
			"MN-2": 0.02,
			"MS":   -0.18,
			"MS-2": -0.18,
			"MO":   -0.18,
			"MT":   -0.22,
			"NE":   -0.28,
			"NV":   0.02,
			"NJ":   0.14,
			"NM":   0.06,
			"NY":   0.24,
			"ND":   -0.34,
			"OH":   -0.06,
			"PA":   0.0,
			"RI":   0.2,
			"TN":   -0.28,
			"TX":   -0.16,
			"UT":   -0.4,
			"VT":   0.3,
			"VA":   0.02,
			"WA":   0.14,
			"WV":   -0.38,
			"WI":   0.0,
			"WY":   -0.5,
		}
	}
	return
}

func LoadCookHouseRatingsHtml(id uint64) (ratings map[string][2]float64, incumbents map[string]string, pvis map[string]float64) {
	var url, cache string
	var maxAge time.Duration
	if id == 0 {
		url = "https://www.cookpolitical.com/ratings/house-race-ratings"
		cache = "cache/Cook_House.html"
		maxAge = time.Hour
	} else {
		url = fmt.Sprintf("https://www.cookpolitical.com/ratings/house-race-ratings/%d", id)
		cache = fmt.Sprintf("cache/Cook_%d.html", id)
		maxAge = -1
	}
	data := LoadCache(url, cache, maxAge, func(r io.Reader) interface{} {
		if rdata, err := ioutil.ReadAll(r); err == nil {
			return string(rdata)
		} else {
			panic(err)
		}
	}).(string)
	ratings = map[string][2]float64{}
	incumbents = map[string]string{}
	pvis = map[string]float64{}
	for i := 0; i < 8; i++ {
		a := strings.Index(data, "solid-seats-modal-in-title") + len("solid-seats-modal-in-title") + 2 // Start of segment
		data = data[a:]
		a = strings.Index(data, "</p>")
		mode := data[:a] // e.g. "Toss-Up Republican"
		idx := map[string]int{"Solid Democratic": 0, "Likely Democratic": 1, "Lean Democratic": 2, "Toss-Up Democratic": 3, "Toss-Up Republican": 4, "Lean Republican": 5, "Likely Republican": 6, "Solid Republican": 7}[mode]
		//idx := i
		b := strings.Index(data, "solid-seats-modal-in-title") // End of segment
		if b == -1 {
			b = len(data)
		}
		for {
			a = strings.Index(data[:b], "-li-color")
			if a == -1 {
				break
			}
			party := data[strings.Index(data, "popup-table-data-cell ")+len("popup-table-data-cell ") : a]
			seat := html.UnescapeString(strings.TrimSpace(data[a+len("-li-color")+2 : strings.Index(data, "</a>")]))
			if seat[3:] == "AL" {
				seat = seat[:3] + "00"
			}
			data = data[a:]
			b -= a
			a = strings.Index(data, "popup-table-data-cell") + len("popup-table-data-cell") + 2
			candidate := html.UnescapeString(data[a : strings.Index(data[a:], "</div>")+a])
			data = data[a:]
			b -= a
			a = strings.Index(data, "popup-table-data-cell") + len("popup-table-data-cell") + 2
			pvi_str := html.UnescapeString(data[a : strings.Index(data[a:], "</div>")+a])
			data = data[a:]
			b -= a
			ratings[seat] = [2]float64{CookConfidence[idx][0], CookConfidence[idx][1]}
			if candidate == "Open" || candidate == "Vacant" {
				incumbents[seat] = ""
			} else {
				incumbents[seat] = map[string]string{"dem": "D", "rep": "R"}[party]
			}
			var pvi float64
			if pvi_str != "EVEN" {
				if pvi_str == "" || id != 0 {
					pvi = CookConfidence[idx][0] - CookConfidence[idx][1]
				} else {
					pvi_int, err := strconv.Atoi(pvi_str[2:])
					if err != nil {
						panic(err)
					}
					pvi = 2 * float64(pvi_int) / 100.0 // Yes, it has to be doubled
					if pvi_str[0] == 'R' {
						pvi = -pvi
					}
				}
			}
			pvis[seat] = pvi
		}
	}
	return
}

func LoadCookGovRatingsHtml(id uint64) (ratings map[string][2]float64, incumbents map[string]string, pvis map[string]float64) {
	var data []map[string]string
	if id == 0 { // 2018
		data = []map[string]string{
			{
				"CA": "O",
				"HI": "D",
				"NY": "D",
			},
			{
				"MN": "O",
				"PA": "D",
				"IL": "R",
			},
			{
				"CO": "O",
				"OR": "D",
				"RI": "D",
				"NM": "O",
				"MI": "O",
			},
			{
				"CT": "O",
			},
			{
				"FL": "O",
				"GA": "O",
				"IA": "R",
				"KS": "O",
				"ME": "O",
				"NV": "O",
				"OH": "O",
				"SD": "O",
				"WI": "R",
			},
			{
				"AK": "D",
				"OK": "O",
			},
			{
				"AZ": "R",
				"NH": "R",
				"SC": "R",
				"TN": "O",
				"MD": "R",
			},
			{
				"AL": "R",
				"AR": "R",
				"ID": "O",
				"MA": "R",
				"NE": "R",
				"TX": "R",
				"WY": "O",
				"VT": "R",
			},
		}
	} else if id == 139343 { // 2016
		data = []map[string]string{
			{
				"DE": "O",
				"WA": "D",
			},
			{
				"OR": "D",
			},
			{
				"MT": "D",
			},
			{
				"MO": "O",
				"NH": "O",
				"VT": "O",
				"WV": "O",
			},
			{
				"IN": "O",
				"NC": "R",
			},
			{},
			{},
			{
				"ND": "O",
				"UT": "R",
			},
		}
	} else if id == 139257 { // 2014
		data = []map[string]string{
			{
				"CA": "D",
				"NY": "D",
				"VT": "D",
			},
			{
				"MN": "D",
				"OR": "D",
				"PA": "R",
			},
			{
				"HI": "O",
				"NH": "O",
			},
			{
				"CO": "D",
				"CT": "D",
				"IL": "D",
				"MD": "O",
				"MA": "O",
				"RI": "O",
			},
			{
				"AK": "R",
				"FL": "R",
				"GA": "R",
				"KS": "R",
				"ME": "R",
				"MI": "R",
				"WI": "R",
			},
			{
				"AZ": "O",
				"AR": "O",
			},
			{
				"IA": "R",
				"NM": "R",
				"SC": "R",
				"TX": "O",
			},
			{
				"AL": "R",
				"ID": "R",
				"NE": "O",
				"NV": "R",
				"OH": "R",
				"OK": "R",
				"SD": "R",
				"TN": "R",
				"WY": "R",
			},
		}
	} else if id == 139101 { // 2012
		data = []map[string]string{
			{
				"DE": "D",
				"VT": "D",
			},
			{},
			{
				"MO": "D",
				"WV": "D",
			},
			{
				"MT": "O",
				"NH": "O",
				"WA": "O",
			},
			{},
			{
				"NC": "O",
			},
			{
				"IN": "O",
			},
			{
				"ND": "R",
				"UT": "R",
			},
		}
	} else if id == 139082 { // 2008
		data = []map[string]string{
			{
				"AR": "D",
				"DE": "D",
				"IA": "D",
				"IL": "D",
			},
			{},
			{
				"NM": "O",
			},
			{
				"WA": "D",
			},
			{},
			{
				"IN": "R",
			},
			{
				"VT": "R",
			},
			{
				"ND": "R",
				"UT": "R",
			},
		}
	}
	ratings = map[string][2]float64{}
	incumbents = map[string]string{}
	pvis = map[string]float64{}
	for idx, m := range data {
		for st, i := range m {
			if _, ok := ratings[st]; ok {
				fmt.Println(id, st)
				panic(nil)
			}
			ratings[st] = [2]float64{CookConfidence[idx][0], CookConfidence[idx][1]}
			incumbents[st] = i
			pvis[st] = CookConfidence[idx][0] - CookConfidence[idx][1]
		}
	}
	if id == 0 {
		pvis = map[string]float64{
			"AK": -0.18,
			"AL": -0.28,
			"AR": -0.3,
			"AZ": -0.10,
			"CA": 0.24,
			"CO": 0.02,
			"CT": 0.12,
			"FL": -0.04,
			"GA": -0.10,
			"HI": 0.36,
			"IA": -0.06,
			"ID": -0.38,
			"IL": 0.14,
			"KS": -0.26,
			"MA": 0.24,
			"MD": 0.24,
			"ME": 0.06,
			"MI": 0.02,
			"MN": 0.02,
			"NE": -0.28,
			"NH": 0.00,
			"NM": 0.06,
			"NV": 0.02,
			"NY": 0.24,
			"OH": -0.06,
			"OK": -0.40,
			"OR": 0.10,
			"PA": 0.00,
			"RI": 0.20,
			"SC": -0.16,
			"SD": -0.28,
			"TN": -0.28,
			"TX": -0.16,
			"VT": 0.30,
			"WI": 0.00,
			"WY": -0.50,
		}
	}
	return
}
