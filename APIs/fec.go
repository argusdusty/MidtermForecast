package APIs

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

var (
	FEC_API_KEY = "DEMO_KEY" // Allows 40 API calls per time period, we only need 27, but you can set this if you like
)

func LoadFECRaces(office string, year int) map[string]float64 { // Ratio of D fundraising / R fundraising for each race
	fundraising_totals := map[string][2]float64{}
	url := fmt.Sprintf("https://www.fec.gov/files/bulk-downloads/%d/weball%02d.zip", year, year%100)
	cache := fmt.Sprintf("cache/fec_weball%d.zip", year)
	cacheTime := 4 * time.Hour
	if year != 2018 {
		cacheTime = -1
	}
	return LoadCache(url, cache, cacheTime, func(r io.Reader) interface{} {
		zr, err := zip.OpenReader(cache)
		if err != nil {
			panic(err)
		}
		for _, zf := range zr.File {
			zfr, err := zf.Open()
			if err != nil {
				panic(err)
			}
			scanner := bufio.NewScanner(zfr)
			for scanner.Scan() {
				line := scanner.Text()
				vals := strings.Split(line, "|") // https://www.fec.gov/campaign-finance-data/all-candidates-file-description/
				candidate_id := vals[0]          // CAND_ID
				party := vals[4]                 // CAND_PTY_AFFILIATION
				raised_raw := vals[5]            // TTL_RECEIPTS
				seat := vals[18]                 // CAND_OFFICE_ST
				if office == "S" && (candidate_id == "S8MN00578" || candidate_id == "S8MN00586" || candidate_id == "S8MS00287" || candidate_id == "S4MS00120" || candidate_id == "S8MS00261") {
					seat += "-2"
				} else if office == "H" {
					seat += "-" + vals[19] // CAND_OFFICE_DISTRICT
				}
				raised, err := strconv.ParseFloat(raised_raw, 64)
				if err != nil {
					panic(err)
				}
				if raised < 0 {
					raised = 0
				}
				if party == "DEM" || party == "DFL" || (party == "IND" && (candidate_id == "S4VT00033" || candidate_id == "S2ME00109")) {
					fundraising_totals[seat] = [2]float64{fundraising_totals[seat][0] + raised, fundraising_totals[seat][1]}
				} else if party == "REP" || party == "GOP" {
					fundraising_totals[seat] = [2]float64{fundraising_totals[seat][0], fundraising_totals[seat][1] + raised}
				}
			}
			if err := scanner.Err(); err != nil {
				panic(err)
			}
		}
		fundraising_ratios := make(map[string]float64, len(fundraising_totals))
		for k, v := range fundraising_totals {
			fundraising_ratios[k] = (v[0] + 10000) / (v[1] + 10000)
		}
		return fundraising_ratios
	}).(map[string]float64)
}

// Apparently this is missing about 30% of candidates for some fucking reason. Good job, FEC. Instead, use the bulk zip file above.
func LoadFECRacesOld(office string, year int) map[string]float64 { // Ratio of D fundraising / R fundraising for each race
	fundraising_totals := map[string][2]float64{}
	var pages int = 1
	for page := 1; page <= pages; page++ {
		url := fmt.Sprintf("https://api.open.fec.gov/v1/candidates/totals/?api_key=%s&office=%s&election_year=%d&page=%d&per_page=100", FEC_API_KEY, office, year, page)
		cache := fmt.Sprintf("cache/fec_%s_%d_%d.json", office, year, page)
		data := LoadCache(url, cache, 36*time.Hour, func(r io.Reader) interface{} {
			dec := json.NewDecoder(r)
			var data map[string]interface{}
			if err := dec.Decode(&data); err != nil {
				panic(err)
			}
			return data
		}).(map[string]interface{})
		pages = int(data["pagination"].(map[string]interface{})["pages"].(float64))
		for _, candidate_data := range data["results"].([]interface{}) {
			candidate := candidate_data.(map[string]interface{})
			if candidate["cycle"].(float64) != 2018 {
				continue
			}
			raised := candidate["receipts"].(float64)
			party := candidate["party"].(string)
			seat := candidate["state"].(string)
			candidate_id := candidate["candidate_id"].(string)
			if office == "S" && (candidate_id == "S8MN00578" || candidate_id == "S8MN00586" || candidate_id == "S8MS00287" || candidate_id == "S4MS00120" || candidate_id == "S8MS00261") {
				seat += "-2"
			} else if office == "H" {
				seat += "-" + candidate["district"].(string)
			}
			if party == "DEM" || party == "DFL" || (party == "IND" && (candidate_id == "S4VT00033" || candidate_id == "S2ME00109")) {
				fundraising_totals[seat] = [2]float64{fundraising_totals[seat][0] + raised, fundraising_totals[seat][1]}
			} else if party == "REP" || party == "GOP" {
				fundraising_totals[seat] = [2]float64{fundraising_totals[seat][0], fundraising_totals[seat][1] + raised}
			}
		}
	}
	fundraising_ratios := make(map[string]float64, len(fundraising_totals))
	for k, v := range fundraising_totals {
		fundraising_ratios[k] = (v[0] + 10000) / (v[1] + 10000)
	}
	return fundraising_ratios
}
