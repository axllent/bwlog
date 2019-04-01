$(function () {

	var charts = [];

	function setupchart(ifname) {
		var tab = $('<li class="nav-item">' +
			'<a class="nav-link" id="' + ifname + '-tab" ' +
			'data-toggle="tab" href="#' + ifname + '" role="tab" ' +
			'aria-controls="' + ifname + '">' + ifname + '</a>'
		);
		$('#nwtabs').append(tab);

		var tabcontent = $('<div class="tab-pane" id="' + ifname + '" '+
			'role="tabpanel" aria-labelledby="' + ifname + '-tab">'
		);

		$('#nwinfo').append(tabcontent);

		var canv = $('<canvas>', { 'id': 'chart_' + ifname, 'class': 'smoothie-chart' });

		tabcontent.append(canv);

		// open first tab
		$('#nwtabs li:first-child a').tab('show');

		charts[ifname] = new SmoothieChart({
			millisPerPixel:100,
			maxValueScale:1.0,
			tooltip:true,
			responsive: true,
			grid:{
				fillStyle: 'rgba(0,0,0,0.88)',
				strokeStyle: 'rgba(0,0,0,0.88)',
				verticalSections: 10,
				verticalSections: 5
			},
			yMinFormatter: function(min, precision) { // callback function that formats the min y value label
				return parseFloat(min).toFixed(0).replace(/\d(?=(\d{3})+)/g, '$&,') + ' kB/s';
			},
			yMaxFormatter: function(max, precision) { // callback function that formats the max y value label
				return parseFloat(max).toFixed(0).replace(/\d(?=(\d{3})+)/g, '$&,') + ' kB/s';
			},
			yIntermediateFormatter: function(intermediate, precision) { // callback function that formats the intermediate y value labels
				return parseFloat(intermediate).toFixed(0).replace(/\d(?=(\d{3})+)/g, '$&,') + ' kB/s';
			}
		});

		var canvas = document.getElementById('chart_' + ifname)
		charts[ifname + '_rx'] = new TimeSeries();
		charts[ifname + '_tx'] = new TimeSeries();

		charts[ifname].addTimeSeries(charts[ifname + '_rx'], {lineWidth:2,strokeStyle:'#00ff00'});
		charts[ifname].addTimeSeries(charts[ifname + '_tx'], {lineWidth:2,strokeStyle:'#ff0018'});
		charts[ifname].streamTo(canvas, 500);
	}

	function connect() {
		var ws = new WebSocket("ws://" + document.location.host + "/stream");

		ws.onmessage = function (event) {
			var obj = JSON.parse(event.data);
			$("#output").empty();
			table = $('<table>', { 'class': 'table' });
			table.append('<tr><th>Interface</td><th>Download</th><th>Upload</th></tr>');
			$(obj).each(function (k, nw) {
				if (charts[nw.If] == undefined ) {
					setupchart(nw.If);
				}
				charts[nw.If + '_rx'].append(new Date().getTime(), nw.Rx);
				charts[nw.If + '_tx'].append(new Date().getTime(), nw.Tx);
				table.append('<tr><td>' + nw.If + '</td><td>' + nw.Rx + '</td><td>' + nw.Tx + '</td></tr>');
			})
			$("#output").append(table)
		}

		ws.onclose = function (e) {
			// console.log('Socket is closed. Reconnect will be attempted in 3 seconds.', e.reason);
			setTimeout(function () {
				connect();
			}, 3000);
		};

		ws.onerror = function (err) {
			// console.error('Socket encountered error: ', err.message, 'Closing socket');
			ws.close();
		}
	}

	connect();
});
