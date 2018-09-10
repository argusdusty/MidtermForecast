$.getJSON(window.location.pathname + ".json", function(race) {
	var x_data = [];
	var y_data = [];
	var text_data = [];
	var color_data = [];
	var min = race.seat_probabilities.length;
	var max = 0;
	for (var i = 0; i < race.seat_probabilities.length; i++) {
		x_data.push(race.seat_probabilities[i].dem_seats);
		y_data.push(race.seat_probabilities[i].probability);
		text_data.push((race.seat_probabilities[i].probability*100).toFixed(2).toString()+'%');
		if (race.seat_probabilities[i].dem_seats < race.seat_probabilities.length/2) {
			color_data.push("rgb(255,0,0)");
		} else {
			color_data.push("rgb(0,0,255)");
		}
		if (i > 0 && y_data[y_data.length-1] > 1e-5 && y_data[y_data.length-2] < 1e-5 && i < min) {
			min = i;
		}
		if (i > 0 && y_data[y_data.length-1] < 1e-5 && y_data[y_data.length-2] > 1e-5 && i > max) {
			max = i+1;
		}
	}
	if (max == 0) {
		max = race.seat_probabilities.length;
	}
	if (min == max) {
		min--;
	}
	x_data = x_data.slice(min, max+1);
	y_data = y_data.slice(min, max+1);
	text_data = text_data.slice(min, max+1);
	color_data = color_data.slice(min, max+1);
	Plotly.newPlot('seats', [{x:x_data, y:y_data, type:'bar', text: text_data, marker:{color: color_data}}], {margin: {r: 10, t: 30, b: 30, l: 10}, legend: {x: 0.92, y: 0.99}, title: document.title, yaxis: {showline: false, showgrid: false, showticklabels: false}, xaxis: {title: 'Number of seats won by Dems'}}, {displayModeBar: false});
});