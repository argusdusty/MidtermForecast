package main

import (
	. "MidtermForecast/Server"
	"crypto/tls"
	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/acme/autocert"
	"net/http"
)

func RunAutocertServer() {
	certManager := &autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache("certs"),
	}

	server := &http.Server{
		Addr: ":https",
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}

	go http.ListenAndServe(":http", certManager.HTTPHandler(nil))
	panic(server.ListenAndServeTLS("", ""))
}

func FaviconHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Cache-Control", "max-age=3000")
	http.ServeFile(w, r, "favicon.ico")
}

func ForecastJsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Cache-Control", "max-age=300")
	http.ServeFile(w, r, "forecast.js")
}

func SeatsJsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Cache-Control", "max-age=300")
	http.ServeFile(w, r, "seats.js")
}

func MapJsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Cache-Control", "max-age=300")
	http.ServeFile(w, r, "map.js")
}

func PastJsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Cache-Control", "max-age=300")
	http.ServeFile(w, r, "past.js")
}

func USCdsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Cache-Control", "max-age=3000")
	http.ServeFile(w, r, "us-cds.json")
}

func USStatesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Cache-Control", "max-age=3000")
	http.ServeFile(w, r, "us-states.json")
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Cache-Control", "max-age=3000")
	WriteHtmlHeader(w, "Forecasts", false, false, false)
	WriteHtmlLines(w, []string{
		"<a href=\"/senate\">Senate Forecast</a>",
		"<a href=\"/house\">House Forecast</a>",
		"<a href=\"/gov\">Governor Forecast</a>",
	})
	WriteHtmlFooter(w)
}

func main() {
	router := mux.NewRouter().StrictSlash(true).Host("{subdomain}.{domain}.{tld}").Subrouter()
	router.HandleFunc("/favicon.ico", FaviconHandler)
	router.HandleFunc("/forecast.js", ForecastJsHandler)
	router.HandleFunc("/seats.js", SeatsJsHandler)
	router.HandleFunc("/map.js", MapJsHandler)
	router.HandleFunc("/past.js", PastJsHandler)
	router.HandleFunc("/us-cds.json", USCdsHandler)
	router.HandleFunc("/us-states.json", USStatesHandler)
	router.HandleFunc("/{type:(?:senate|house|gov)}/{race}/polling.{format}", PollsHandler)
	router.HandleFunc("/{type:(?:senate|house|gov)}/{race}/polling", PollsHandler)
	router.HandleFunc("/{type:(?:senate|house|gov)}/{race}/experts.{format}", ExpertsHandler)
	router.HandleFunc("/{type:(?:senate|house|gov)}/{race}/experts", ExpertsHandler)
	router.HandleFunc("/{type:(?:senate|house|gov)}/{race}/fundamentals.{format}", FundamentalsHandler)
	router.HandleFunc("/{type:(?:senate|house|gov)}/{race}/fundamentals", FundamentalsHandler)
	router.HandleFunc("/{type:(?:senate|house|gov)}/polling.{format}", PollsHandler)
	router.HandleFunc("/{type:(?:senate|house|gov)}/polling", PollsHandler)
	router.HandleFunc("/{type:(?:senate|house|gov)}/{race}.{format}", ForecastHandler)
	router.HandleFunc("/{type:(?:senate|house|gov)}/{race}", ForecastHandler)
	router.HandleFunc("/{type:(?:senate|house|gov)}.{format}", ForecastHandler)
	router.HandleFunc("/{type:(?:senate|house|gov)}", ForecastHandler)
	router.HandleFunc("/", IndexHandler)
	gzipRouter := gziphandler.GzipHandler(router)
	http.Handle("/", gzipRouter)
	RunAutocertServer()
}
