package Server

import (
	. "MidtermForecast/Predict"
	"net/http"
	"strings"
	"time"
)

func loadForecast(vars map[string]string) (interface{}, time.Time, error) {
	var F Forecast
	err, modtime := LoadForecast(vars["type"], &F)
	if vars["race"] != "" {
		for _, r := range F.RaceProbabilities {
			if r.Race == vars["race"] {
				return r, modtime, err
			}
		}
	}
	return F, modtime, err
}

func writeForecast(w http.ResponseWriter, v interface{}, vars map[string]string) {
	switch v.(type) {
	case Forecast:
		F := v.(Forecast)
		name := strings.Title(vars["type"]) + " Forecast"
		WriteHtmlHeader(w, name, false, true, true)
		WriteHtmlLines(w, F.GetText(vars["type"]))
		if _, ok := F.RaceProbabilities["KS"]; !ok {
			// Not Gov
			WriteSeatsScript(w)
		}
		WriteMapScript(w)
		WritePastScript(w)
		WriteHtmlFooter(w)
	case RaceProbability:
		R := v.(RaceProbability)
		name := strings.Title(vars["type"]) + " " + vars["race"] + " Forecast"
		WriteHtmlHeader(w, name, false, true, false)
		WriteHtmlLines(w, R.GetText(vars["type"]))
		WriteForecastScript(w)
		WritePastScript(w)
		WriteHtmlFooter(w)
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 page not found"))
	}
}

func ForecastHandler(w http.ResponseWriter, r *http.Request) {
	ValueHandler(w, r, "Forecast", loadForecast, writeForecast)
}
