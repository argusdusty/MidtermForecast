package Server

import (
	. "MidtermForecast/Predict"
	"net/http"
	"strings"
	"time"
)

func loadFundamentals(vars map[string]string) (interface{}, time.Time, error) {
	var F RaceFundamentals
	err, modtime := LoadFundamentals(vars["type"], &F)
	return F[vars["race"]], modtime, err
}

func writeFundamentals(w http.ResponseWriter, v interface{}, vars map[string]string) {
	F := v.(Fundamentals)
	name := strings.Title(vars["type"]) + " " + vars["race"] + " Fundamental Steps"
	WriteHtmlHeader(w, name, false, false, false)
	WriteHtmlLines(w, F.GetText(name))
	WriteHtmlFooter(w)
}

func FundamentalsHandler(w http.ResponseWriter, r *http.Request) {
	ValueHandler(w, r, "Fundamentals", loadFundamentals, writeFundamentals)
}
