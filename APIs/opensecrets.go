package APIs

import (
	"encoding/csv"
	"fmt"
	"time"
)

// Not used. Using FEC instead
func LoadOpenSecretsRace(mode, st string) map[string]float64 {
	var id string
	if mode == "S" {
		// 1 vs 2 based on schedule/class, not whether or not it's a special
		id = map[string]string{
			"AZ":   "AZS2",
			"CA":   "CAS2",
			"CT":   "CTS1",
			"DE":   "DES1",
			"FL":   "FLS1",
			"HI":   "HIS2",
			"IN":   "INS1",
			"ME":   "MES1",
			"MD":   "MDS1",
			"MA":   "MAS1",
			"MI":   "MIS2",
			"MN":   "MNS2",
			"MN-2": "MNS1",
			"MS":   "MSS1",
			"MS-2": "MSS2",
			"MO":   "MOS1",
			"MT":   "MTS1",
			"NE":   "NES1",
			"NV":   "NVS1",
			"NJ":   "NJS1",
			"NM":   "NMS1",
			"NY":   "NYS1",
			"ND":   "NDS1",
			"OH":   "OHS1",
			"PA":   "PAS1",
			"RI":   "RIS1",
			"TN":   "TNS1",
			"TX":   "TXS1",
			"UT":   "UTS1",
			"VT":   "VTS1",
			"VA":   "VAS1",
			"WA":   "WAS1",
			"WV":   "WVS1",
			"WI":   "WIS1",
			"WY":   "WYS1",
		}[st]
	} else if mode == "H" {
		if st[3:] == "00" {
			id = st[:2] + "01"
		} else {
			id = st[:2] + st[3:]
		}
	}
	r := LoadCache(fmt.Sprintf("https://www.opensecrets.org/races/summary.csv?cycle=2018&id=%s", id), fmt.Sprintf("opensecrets_%s.csv", id), 36*time.Hour)
	lines, err := csv.NewReader(r).ReadAll()
	if err != nil {
		panic(err)
	}
	// TODO
	lines[0] = nil // Junk line for compiler
	return nil
}
