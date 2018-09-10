var gammaln = function gammaln(x) {
	if(x < 0) return NaN;
	if(x == 0) return Infinity;
	if(!isFinite(x)) return x;

	var lnSqrt2PI = 0.91893853320467274178;
	var gamma_series = [76.18009172947146, -86.50532032941677, 24.01409824083091, -1.231739572450155, 0.1208650973866179e-2, -0.5395239384953e-5];
	var denom;
	var x1;
	var series;

	// Lanczos method
	denom = x+1;
	x1 = x + 5.5;
	series = 1.000000000190015;
	for(var i = 0; i < 6; i++) {
		series += gamma_series[i] / denom;
		denom += 1.0;
	}
	return( lnSqrt2PI + (x+0.5)*Math.log(x1) - x1 + Math.log(series/x) );
};
var data = [];
var x_data = [];

var N = 10000;

for (var i = 0; i <= N; i++) {
	x_data.push((i*100)/N);
}

$.getJSON(window.location.pathname + ".json", function(race) {
	var totalC = 0.0;
	for (var key in race.concentration_params) {
		console.log(key, race.concentration_params, race.concentration_params[key]);
		totalC += race.concentration_params[key];
	}
	var min = N;
	var max = 0;
	for (var key in race.concentration_params) {
		if (race.concentration_params[key]/totalC < 0.05) {
			continue;
		}
		var alpha = race.concentration_params[key];
		var beta = totalC-race.concentration_params[key];
		var gamma = gammaln(totalC)-gammaln(alpha)-gammaln(beta);
		var tmp_data = [];
		for (var i = 0; i <= N; i++) {
			var s = Math.exp(Math.log(i/N)*(alpha-1)+Math.log((N-i)/N)*(beta-1)+gamma);
			if (beta == 0 && i == N) {
				s = 1.0;
			}
			tmp_data.push(s);
			if (i > 0 && tmp_data[tmp_data.length-1] > 1e-5 && tmp_data[tmp_data.length-2] < 1e-5 && i < min) {
				min = i;
			}
			if (i > 0 && tmp_data[tmp_data.length-1] < 1e-5 && tmp_data[tmp_data.length-2] > 1e-5 && i > max) {
				max = i+1;
			}
		}
		data.push({x:x_data, y:tmp_data, mode:'lines', name: key, line: {width: 4, color: {"D": "#1A80C4", "R": "#FF0125"}[key]}});
	}
	if (max == 0) {
		max = N;
	}
	for (var i in data) {
		data[i].text = [];
		var tsy = 0.0;
		for (var j in data[i].y) {
			tsy += data[i].y[j];
		}
		var sy = 0.0;
		for (var j in data[i].y) {
			sy += data[i].y[j] / tsy;
			data[i].text.push((sy*100).toFixed(2).toString()+'%');
			data[i].y[j] *= 100/tsy;
		}
		console.log(i, sy, tsy, data[i]);
	}
	for (var i in data) {
		if (min == max) {
			min--;
		}
		data[i].x = data[i].x.slice(min, max+1);
		data[i].y = data[i].y.slice(min, max+1);
		data[i].text = data[i].text.slice(min, max+1);
	}
	Plotly.newPlot('forecast', data.reverse(), {margin: {r: 10, t: 30, b: 30, l: 10}, legend: {x: 0.92, y: 0.99}, title: document.title, yaxis: {showline: false, showgrid: false, showticklabels: false}, xaxis: {title: 'Percent of vote'}}, {displayModeBar: false});
});