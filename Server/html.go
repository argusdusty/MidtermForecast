package Server

import (
	"net/http"
)

func WriteHtmlHeader(w http.ResponseWriter, name string, refresh, plot, maps bool) {
	w.Write([]byte("<!DOCTYPE html>"))
	w.Write([]byte("<html lang=\"en\">"))
	w.Write([]byte("<head>"))
	w.Write([]byte("<title>" + name + "</title>"))
	if refresh {
		w.Write([]byte("<meta http-equiv=\"refresh\" content=\"15\"/>"))
	}
	w.Write([]byte("<link rel=\"icon\" href=\"/favicon.ico\"/>"))
	w.Write([]byte("<meta name=\"author\" content=\"argusdusty\"/>"))
	w.Write([]byte("<meta name=\"theme-color\" content=\"#008080\">"))
	w.Write([]byte("<meta name=\"description\" content=\"" + name + "\"/>"))
	w.Write([]byte("<meta name=\"twitter:card\" content=\"summary\"/>"))
	w.Write([]byte("<meta name=\"twitter:site\" content=\"@argusdusty\"/>"))
	w.Write([]byte("<meta name=\"twitter:title\" content=\"" + name + "\"/>"))
	w.Write([]byte("<meta name=\"twitter:description\" content=\"" + name + "\"/>"))
	w.Write([]byte("<meta name=\"twitter:image\" content=\"https://avatars2.githubusercontent.com/u/627841\"/>"))
	w.Write([]byte("<meta name=\"og:site_name\" content=\"argusdusty\"/>"))
	w.Write([]byte("<meta name=\"og:title\" content=\"" + name + "\"/>"))
	w.Write([]byte("<meta name=\"og:description\" content=\"" + name + "\"/>"))
	w.Write([]byte("<meta name=\"og:url\" content=\"http://midterms.argusdusty.com\"/>"))
	w.Write([]byte("<meta name=\"og:image:type\" content=\"image/png\"/>"))
	w.Write([]byte("<meta name=\"og:image\" content=\"https://avatars2.githubusercontent.com/u/627841\"/>"))
	if plot {
		w.Write([]byte(`
<script src="https://cdn.plot.ly/plotly-1.40.1.min.js" integrity="sha512-HMt7wrktXhUZN8FdNsvegUHHG4hp4MxDlMAw+WFljh4BYncLBOStNhqHflACfq0rVV9+kh3+4ljgfCEQsEWF5Q==" crossorigin="anonymous"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/mathjs/4.0.0/math.min.js" integrity="sha512-43A30cNU92nqP4eEw4kukii1xoPQXUm19/diDPMtjmL3R746fUfvNUXvdTNXr/nLj4yNGsrvs+fwwuWAwFVlIA==" crossorigin="anonymous"></script>
<script src="https://code.jquery.com/jquery-3.3.1.min.js" integrity="sha512-+NqPlbbtM1QqiK8ZAo4Yrj2c4lNQoGv8P79DPtKzj++l5jnN39rHA/xsqn8zE9l0uSoxaCdrOgFs6yjyfbBxSg==" crossorigin="anonymous"></script>
<script src="https://unpkg.com/chroma-js@1.3.6/chroma.js" integrity="sha512-8EvwGP4UyuLfixHVUfxPmivN2n5KGxsGa5KpASRSDDRyM9xsUW05s63nILJHQA9U/E5WyAC1znrNkXelgbz1RQ==" crossorigin="anonymous"></script>`))
	}
	if maps {
		w.Write([]byte(`
<link rel="stylesheet" href="https://unpkg.com/leaflet@1.3.1/dist/leaflet.css" integrity="sha256-iYUgmrapfDGvBrePJPrMWQZDcObdAcStKBpjP3Az+3s=" crossorigin="anonymous">
<script src="https://unpkg.com/leaflet@1.3.1/dist/leaflet.js" integrity="sha512-/Nsx9X4HebavoBvEBuyp3I7od5tA0UzAxs+j83KgC8PU0kgB4XiK4Lfe4y4cgBtaRJQEIFCW+oC506aPT2L1zw==" crossorigin="anonymous"></script>
<script src="https://code.jquery.com/jquery-3.3.1.min.js" integrity="sha512-+NqPlbbtM1QqiK8ZAo4Yrj2c4lNQoGv8P79DPtKzj++l5jnN39rHA/xsqn8zE9l0uSoxaCdrOgFs6yjyfbBxSg==" crossorigin="anonymous"></script>
<script src="https://unpkg.com/topojson-client@3.0.0/dist/topojson-client.js" integrity="sha512-+ACuObPkytEib/hU5nkAVPOeIfnYzKJsBr6Chwp5WiZ/k9n3a2rlb0gLGj9n0NISTSKOghznrQacjJtpyqdW0w==" crossorigin="anonymous"></script>`))
		if !plot {
			w.Write([]byte(`
<script src="https://unpkg.com/chroma-js@1.3.6/chroma.js" integrity="sha512-8EvwGP4UyuLfixHVUfxPmivN2n5KGxsGa5KpASRSDDRyM9xsUW05s63nILJHQA9U/E5WyAC1znrNkXelgbz1RQ==" crossorigin="anonymous"></script>`))
		}
	}
	w.Write([]byte("</head>"))
	w.Write([]byte("<body>"))
}

func WriteHtmlLines(w http.ResponseWriter, text []string) {
	for _, line := range text {
		if len(line) == 0 {
			w.Write([]byte("<br>"))
		} else {
			w.Write([]byte("<p>" + line + "</p>"))
		}
	}
}

func WriteHtmlScript(w http.ResponseWriter, script string) {
	w.Write([]byte("<script type=\"text/javascript\">"))
	w.Write([]byte(script))
	w.Write([]byte("</script>"))
}

func WriteHtmlFooter(w http.ResponseWriter) {
	w.Write([]byte("</body>"))
	w.Write([]byte("</html>"))
}

func WriteForecastScript(w http.ResponseWriter) {
	w.Write([]byte("<div id=\"forecast\"></div>"))
	w.Write([]byte("<script src=\"/forecast.js\"></script>"))
}

func WriteSeatsScript(w http.ResponseWriter) {
	w.Write([]byte("<div id=\"seats\"></div>"))
	w.Write([]byte("<script src=\"/seats.js\"></script>"))
}

func WritePastScript(w http.ResponseWriter) {
	w.Write([]byte("<div id=\"past\"></div>"))
	w.Write([]byte("<script src=\"/past.js\"></script>"))
}

func WriteMapScript(w http.ResponseWriter) {
	w.Write([]byte(`<style>
#map { width: 100%; max-width: 600px; height: 500px; }
.info { padding: 6px 8px; font: 14px/16px Arial, Helvetica, sans-serif; background: white; background: rgba(255,255,255,0.8); box-shadow: 0 0 15px rgba(0,0,0,0.2); border-radius: 5px; }
.info h4 { margin: 0 0 5px; color: #777; }
.legend { text-align: left; line-height: 18px; color: #555; }
.legend i { width: 18px; height: 18px; float: left; margin-right: 8px; }
</style>`))
	w.Write([]byte("<div id=\"map\"></div>"))
	w.Write([]byte("<script src=\"/map.js\"></script>"))
}
