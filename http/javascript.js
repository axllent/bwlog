$(function () {

	var charts = [];

	function setupchart(ifname) {
		var tab = $('<li class="nav-item">' +
			'<a class="nav-link" id="' + ifname + '-tab" ' +
			'data-toggle="tab" href="#' + ifname + '" role="tab" ' +
			'aria-controls="' + ifname + '">' + ifname + '</a>'
		);
		// $('#nwtabs').append(tab);
		tab.insertBefore('#nwtabs .header')

		var tabcontent = $('<div class="tab-pane fade" id="' + ifname + '" ' +
			'role="tabpanel" aria-labelledby="' + ifname + '-tab">' +
			'<p class="float-right mt-2 mb-0"><small id="CurStats' + ifname + '">' +
			'<h3 class="mt-2 mb-0">' + ifname + '</h3>'
		);

		$('#nwinfo').append(tabcontent);

		var canv = $('<canvas>', { 'id': 'chart_' + ifname, 'class': 'smoothie-chart' });

		tabcontent.append(canv);

		// open first tab
		$('#nwtabs a').first().tab('show');

		charts[ifname] = new SmoothieChart({
			millisPerPixel: 100,
			maxValueScale: 1.0,
			tooltip: true,
			responsive: true,
			grid: {
				// fillStyle: 'rgba(0,0,0,0.88)',
				// strokeStyle: 'rgba(0,0,0,0.88)',
				fillStyle: '#333333',
				strokeStyle: 'rgba(255,255,255,0.1)',
				verticalSections: 10,
				verticalSections: 5
			},
			yMinFormatter: function (min, precision) { // callback function that formats the min y value label
				return parseFloat(min).toFixed(0).replace(/\d(?=(\d{3})+)/g, '$&,') + ' kB/s';
			},
			yMaxFormatter: function (max, precision) { // callback function that formats the max y value label
				return parseFloat(max).toFixed(0).replace(/\d(?=(\d{3})+)/g, '$&,') + ' kB/s';
			},
			yIntermediateFormatter: function (intermediate, precision) { // callback function that formats the intermediate y value labels
				return parseFloat(intermediate).toFixed(0).replace(/\d(?=(\d{3})+)/g, '$&,') + ' kB/s';
			}
		});

		var canvas = document.getElementById('chart_' + ifname)
		charts[ifname + '_rx'] = new TimeSeries();
		charts[ifname + '_tx'] = new TimeSeries();

		charts[ifname].addTimeSeries(charts[ifname + '_rx'], { lineWidth: 2, strokeStyle: '#00ff00' });
		charts[ifname].addTimeSeries(charts[ifname + '_tx'], { lineWidth: 2, strokeStyle: '#ff0018' });
		charts[ifname].streamTo(canvas, 500);
	}

	function connect() {
		var ws = new WebSocket("ws://" + document.location.host + "/stream");

		ws.onmessage = function (event) {
			$('#Led').addClass('connected');
			var obj = JSON.parse(event.data);
			$(obj).each(function (k, nw) {
				if (charts[nw.If] == undefined) {
					setupchart(nw.If);
				}
				charts[nw.If + '_rx'].append(new Date().getTime(), nw.Rx);
				charts[nw.If + '_tx'].append(new Date().getTime(), nw.Tx);

				$('#CurStats' + nw.If).html(
					'<span class="rx">' + nw.Rx + ' kB/s</span> / ' +
					'<span class="tx">' + nw.Tx + ' kB/s</span>'
				);
			})
			$("#output").append(table)
		}

		ws.onclose = function (e) {
			$('#Led').removeClass('connected');
			setTimeout(function () {
				// reconnect
				connect();
			}, 3000);
		};

		ws.onerror = function (err) {
			ws.close();
		}
	}

	connect();
});
