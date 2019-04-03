$(function () {

	var charts = [];

	function setupchart(ifname) {
		var tab = $('<li class="nav-item">' +
			'<a class="nav-link" id="' + ifname + '-tab" ' +
			'data-toggle="tab" href="#' + ifname + '" role="tab" ' +
			'aria-controls="' + ifname + '">' + ifname + '</a>'
		);
		tab.insertBefore('#nwtabs .header');

		var tabcontent = $('<div class="tab-pane fade" id="' + ifname + '" ' +
			'role="tabpanel" aria-labelledby="' + ifname + '-tab">' +
			'<p class="float-right my-1"><small id="CurStats' + ifname + '">' +
			'<h3 class="my-2">' + ifname + '</h3>'
		);

		$('#nwinfo').append(tabcontent);

		var canv = $('<canvas>', { 'id': 'chart_' + ifname, 'class': 'smoothie-chart' });

		tabcontent.append(canv);

		var day_stats = $('<div class="col-md-6"><h4>Daily totals</h4><div id="DayStats' + ifname + '"></div></div>');
		var month_stats = $('<div class="col-md-6"><h4>Monthly totals</h4><div id="MonthStats' + ifname + '"></div></div>');

		var stats_row = $('<div class="row">');
		stats_row.append(day_stats);
		stats_row.append(month_stats);
		tabcontent.append(stats_row);

		refreshMonthly(ifname);


		// open first tab
		$('#nwtabs a').first().tab('show');

		charts[ifname] = new SmoothieChart({
			millisPerPixel: 100,
			maxValueScale: 1.0,
			tooltip: true,
			responsive: true,
			grid: {
				fillStyle: '#333333',
				strokeStyle: 'rgba(255,255,255,0.1)',
				verticalSections: 10,
				verticalSections: 5
			},
			yMinFormatter: function (min, precision) {
				return 0;
			},
			yMaxFormatter: function (max, precision) {
				return humanFileSize(max, 1024) + '/s';
			},
			yIntermediateFormatter: function (intermediate, precision) {
				return humanFileSize(intermediate, 1024) + '/s';
			}
		});

		var canvas = document.getElementById('chart_' + ifname)
		charts[ifname + '_rx'] = new TimeSeries();
		charts[ifname + '_tx'] = new TimeSeries();

		charts[ifname].addTimeSeries(charts[ifname + '_rx'], { lineWidth: 2, strokeStyle: '#00ff00'/*, fillStyle:'rgba(0,255,0,0.15)'*/ });
		charts[ifname].addTimeSeries(charts[ifname + '_tx'], { lineWidth: 2, strokeStyle: '#00edff'/*, fillStyle:'rgba(0,237,255,0.15)'*/ });
		charts[ifname].streamTo(canvas, 500);
	}

	function humanFileSize(bytes, si) {
		var thresh = si ? 1000 : 1024;
		if(Math.abs(bytes) < thresh) {
			return bytes + ' kB/s';
		}
		var units = si
			? ['MB','GB','TB']
			: ['MiB','GiB','TiB'];
		var u = -1;
		do {
			bytes /= thresh;
			++u;
		} while(Math.abs(bytes) >= thresh && u < units.length - 1);
		return bytes.toFixed(1)+' '+units[u];
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
					'<span class="rx">' + humanFileSize(nw.Rx, 1024) + '</span> / ' +
					'<span class="tx">' + humanFileSize(nw.Tx, 1024) + '</span>'
				);
			});
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

	function refreshMonthly(nwif) {
		$.getJSON('/stats/' + nwif, function( data ) {
			$('#MonthStats' + nwif).empty();
			var table = $('<table>', {class: 'table'});
			var ths = $('<tr><th>Month</th><th class="text-right">Downloaded</th><th class="text-right">Uploaded</th></tr>')
			table.append(ths);
			$.each( data, function(idx, vals) {
				console.log(vals);
				var tr = $('<tr>');
				table.append(tr);
				var td = $('<td>' + vals.Date + '</td>');
				tr.append(td);
				var td = $('<td class="text-right">' + humanFileSize(vals.RX, 1024) + '</td>');
				tr.append(td);
				var td = $('<td class="text-right">' + humanFileSize(vals.TX, 1024) + '</td>');
				tr.append(td);
			});

			$('#MonthStats' + nwif).append(table);
		});
	}
});
