function getProb(e) {
	if ('dem_maj_prob' in e) {
		return e.dem_maj_prob;
	} else {
		return e.dem_win_prob;
	}
}

$.getJSON(window.location.pathname + ".json", function(race) {
	var x_data = [];
	var y_data = [];
	var text_data = [];
	var color_data = [];
	for (var i = 0; i < race.past.length; i++) {
		x_data.push(race.past[i].date);
		y_data.push(getProb(race.past[i])*100);
		text_data.push((getProb(race.past[i])*100).toFixed(2).toString()+'%');
		color_data.push(chroma.scale(['red', 'blue'])(getProb(race.past[i])));
	}
	console.log(x_data);
	console.log(y_data);
	Plotly.newPlot('past', [{x:x_data, y:y_data, type:'lines', text: text_data, marker:{color: color_data}}], {margin: {r: 30, t: 30, b: 30, l: 60}, legend: {x: 0.92, y: 0.99}, title: document.title + " History", yaxis: {range: [0, 100], title: 'Past Dem Majority Odds'}, xaxis: {title: 'Time'}}, {displayModeBar: false});
});