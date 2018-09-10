package Server

import (
	. "MidtermForecast/Predict"
	"net/http"
	"strings"
	"time"
)

func loadExperts(vars map[string]string) (interface{}, time.Time, error) {
	var E RaceMapExperts
	err, modtime := LoadExperts(vars["type"], &E)
	return E[vars["race"]], modtime, err
}

func writeExperts(w http.ResponseWriter, v interface{}, vars map[string]string) {
	E := v.(MapExperts)
	name := strings.Title(vars["type"]) + " " + vars["race"] + " Expert Forecasts"
	WriteHtmlHeader(w, name, false, false, false)
	WriteHtmlLines(w, E.GetText(name))
	WriteHtmlFooter(w)
}

func ExpertsHandler(w http.ResponseWriter, r *http.Request) {
	ValueHandler(w, r, "Experts", loadExperts, writeExperts)
}
