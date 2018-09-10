var info = L.control();

info.onAdd = function (map) {
	this._div = L.DomUtil.create('div', 'info');
	this.reset();
	return this._div;
};

info.update = function (raceProbs) {
	this._div.innerHTML = '<h4>' + raceProbs.race + '</h4>';
	this._div.innerHTML += 'Dem win probability: ' + (raceProbs.dem_win_prob*100).toFixed(1).toString() + '%</br>';
	this._div.innerHTML += 'Dem vote margin expected: ' + ((raceProbs.dem_vote_exp-0.5)*2*100).toFixed(2).toString() + '%</br>';
};

info.reset = function () {
	this._div.innerHTML = '<h4>Area</h4>Hover over a region';
};

function getRaceProbs(feature) {
	if ('iso_3166_2' in feature.properties) {
		id = feature.properties.iso_3166_2;
	} else {
		id = feature.id;
	}
	if (id in forecast.race_probabilities) {
		return forecast.race_probabilities[id]
	}
	return {};
}

function getColor(raceProbs) {
	return chroma.scale(['red', 'blue'])(raceProbs.dem_win_prob);
}

function style(feature) {
	var color = getColor(getRaceProbs(feature));
	return {
		weight: 1,
		opacity: 1,
		color: 'grey',
		fillOpacity: color.alpha(),
		fillColor: color.hex()
	};
}

function onClick(e) {
	var layer = e.target;
	raceProbs = getRaceProbs(layer.feature);

	 	
	window.location.href = window.location.pathname + "/" + raceProbs.race;
}

function highlightFeature(e) {
	var layer = e.target;

	layer.setStyle({
		weight: 3,
		color: '#666666',
	});

	if (!L.Browser.ie && !L.Browser.opera && !L.Browser.edge) {
		layer.bringToFront();
	}

	info.update(getRaceProbs(layer.feature));
}

function resetHighlight(e) {
	geojson.resetStyle(e.target);
	info.reset();
}

function onEachFeature(feature, layer) {
	raceProbs = getRaceProbs(feature);
	if (!('race' in raceProbs)) {
		return;
	}
	layer.on({
		mouseover: highlightFeature,
		mouseout: resetHighlight,
		click: onClick,
	});
}

$.getJSON(window.location.pathname + ".json", function(data) {
	forecast = data;
	var topopath = "/us-states.json";
	if (window.location.pathname.includes("house")) {
		topopath = "/us-cds.json"
	}
	$.getJSON(topopath, function(topodata) {
		map = L.map('map');
		map.setView([0, 0], 0);

		for (key in topodata.objects) {
			if (key != "settings" && key != "roads" && key != "outline" && key != "cities") {
				geodata = topojson.feature(topodata, topodata.objects[key]);
				geojson = L.geoJson(geodata, {style: style, onEachFeature: onEachFeature, coordsToLatLng: function(coords) {
					if (topopath == "/us-states.json") {
						return [coords[1], coords[0]];
					}
					return [coords[1]/100000.0, coords[0]/100000.0];
				}}).addTo(map);
				map.fitBounds(geojson.getBounds());
			}
		}

		info.addTo(map);
		//map.setMinZoom(map.getZoom()-1);
		//map.setMaxZoom(map.getZoom()+4);
	});
});