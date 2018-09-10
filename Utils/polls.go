package Utils

import (
	"time"
)

var (
	PollsterBiases = map[string][2]float64{} // weight, bias for each pollster, reverse-engineered off of 538's ratings data
)

type Poll struct {
	Pollster      string             `json:"pollster"`
	URL           string             `json:"url"`
	StartDate     time.Time          `json:"start"`
	EndDate       time.Time          `json:"end"`
	Subpopulation string             `json:"subpop"`
	Number        float64            `json:"num"`
	Candidates    map[string]float64 `json:"candidates"`
}

func (A Poll) Compare(B Poll) bool {
	// Compare dates
	if A.EndDate.After(B.EndDate) {
		return true
	} else if A.EndDate.Before(B.EndDate) {
		return false
	}
	if A.StartDate.After(B.StartDate) {
		return true
	} else if A.StartDate.Before(B.StartDate) {
		return false
	}
	// Compare sample size
	if A.Number > B.Number {
		return true
	} else if A.Number < B.Number {
		return false
	}
	if A.Pollster != B.Pollster {
		// Compare historical pollster quality
		if PollsterBiases[A.Pollster][0] > PollsterBiases[B.Pollster][0] {
			return true
		} else if PollsterBiases[A.Pollster][0] < PollsterBiases[B.Pollster][0] {
			return false
		}
		// Compare subpopulation quality
		if A.Subpopulation != B.Subpopulation {
			switch A.Subpopulation {
			case "LV":
				return true
			case "RV":
				if B.Subpopulation == "LV" {
					return false
				}
				return true
			case "A", "V":
				if B.Subpopulation == "LV" || B.Subpopulation == "RV" {
					return false
				}
				if !(B.Subpopulation == "A" || B.Subpopulation == "V") {
					return true
				}
			}
		}
		// Compare pollster names
		if A.Pollster < B.Pollster {
			return true
		} else if A.Pollster < B.Pollster {
			return false
		}
	}
	// Compare what percentage they give to the Dems/GOP? idk at this point
	if A.Candidates["D"] > B.Candidates["D"] {
		return true
	} else if A.Candidates["D"] < B.Candidates["D"] {
		return false
	}
	if A.Candidates["R"] > B.Candidates["R"] {
		return true
	} else if A.Candidates["R"] < B.Candidates["R"] {
		return false
	}
	// Okay, the polls are identical. You've had your fun
	return true
}
