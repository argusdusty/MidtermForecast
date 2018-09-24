package Server

import (
	. "MidtermForecast/Predict"
	. "MidtermForecast/Utils"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func loadPolls(vars map[string]string) (interface{}, []byte, time.Time, error) {
	var P RaceMapPolls
	raw, err, modtime := LoadPolls(vars["type"], &P)
	if vars["race"] != "" {
		var b []byte
		b, err = json.Marshal(P[vars["race"]])
		return P[vars["race"]], b, modtime, err
	}
	return P, raw, modtime, err
}

func writePolls(w http.ResponseWriter, v interface{}, vars map[string]string) {
	switch v.(type) {
	case RaceMapPolls:
		P := v.(RaceMapPolls)
		name := strings.Title(vars["type"]) + " Polls"
		WriteHtmlHeader(w, name, false, false, false)
		WriteHtmlLines(w, P.GetText(name))
		WriteHtmlFooter(w)
	case []Poll:
		P := v.([]Poll)
		name := strings.Title(vars["type"]) + " " + vars["race"] + " Polls"
		WriteHtmlHeader(w, name, false, false, false)
		WriteHtmlLines(w, Polls(P).GetText(name))
		WriteHtmlFooter(w)
	}
}

func PollsHandler(w http.ResponseWriter, r *http.Request) {
	ValueHandler(w, r, "Polls", loadPolls, writePolls)
}
