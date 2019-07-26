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

		var day_stats = $('<div class="col-md-6"><h4 id="StatsTop">Daily totals</h4><div id="DayStats' + ifname + '"></div></div>');
		var month_stats = $('<div class="col-md-6"><h4>Monthly totals</h4><div id="MonthStats' + ifname + '"></div></div>');

		var stats_row = $('<div class="row">');
		stats_row.append(day_stats);
		stats_row.append(month_stats);
		tabcontent.append(stats_row);

		loadMonthlyStats(ifname);

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
				return humanFileSize(max) + '/s';
			},
			yIntermediateFormatter: function (intermediate, precision) {
				return humanFileSize(intermediate) + '/s';
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
		if (Math.abs(bytes) < thresh) {
			return bytes + ' kB';
		}
		var units = si
			? ['MB', 'GB', 'TB']
			: ['MiB', 'GiB', 'TiB'];
		var u = -1;
		do {
			bytes /= thresh;
			++u;
		} while (Math.abs(bytes) >= thresh && u < units.length - 1);
		return bytes.toFixed(1) + ' ' + units[u];
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
					'<span class="rx">' + humanFileSize(nw.Rx) + '</span> / ' +
					'<span class="tx">' + humanFileSize(nw.Tx) + '</span>'
				);

				$('#' + nw.If + ' tr.live-stat').each(function(){
					var rx = $(this).find('td[data-rx]').first();
					var tx = $(this).find('td[data-tx]').first();
					var total = $(this).find('td[data-total]').first();

					var new_rx = rx.data('rx') + nw.Rx;
					rx.html(humanFileSize(new_rx));
					rx.data('rx', new_rx);

					var new_tx = tx.data('tx') + nw.Tx;
					tx.html(humanFileSize(new_tx));
					tx.data('tx', new_tx);

					var new_total = humanFileSize(new_rx + new_tx);
					total.html(new_total)

				});
			});
		}

		ws.onclose = function (e) {
			$('#Led').removeClass('connected');
			setTimeout(function () {
				connect(); // reconnect
			}, 3000);
		};

		ws.onerror = function (err) {
			ws.close();
		}
	}

	connect();

	function loadMonthlyStats(nwif, highlight = false) {
		var MyDate = new Date();
		var cur_date = MyDate.getFullYear() + '-' + ('0' + (MyDate.getMonth()+1)).slice(-2);

		$.getJSON('/stats/' + nwif, function (data) {
			var fresh_start = $('#MonthStats' + nwif).is(':empty');
			if (!fresh_start) {
				$('#MonthStats' + nwif).empty();
			}
			var table = $('<table>', { class: 'table table-bordered table-dark table-hover table-hover table-sm' });
			var thds = $('<tr><th>Month</th><th class="text-right">Downloaded</th><th class="text-right">Uploaded</th><th class="text-right">Total</th></tr>')
			table.append(thds);
			var tbody = $('<tbody>');
			table.append(tbody);

			$.each(data, function (idx, vals) {
				var tr = $('<tr>', {class: 'clickable ' + vals.Date});
				if (vals.Date == cur_date) {
					tr.addClass('live-stat');
				}
				tr.data('month', vals.Date);
				tr.on('clickload', function() {
					$(this).parent().find('tr').removeClass('table-active');
					$(this).addClass('table-active');
					loadDaily(nwif, $(this).data('month'));
				});
				tr.on('click', function() {
					var page_pos = $(window).scrollTop();
					var stats_pos = $("#StatsTop").offset().top - 20;
					if (page_pos > stats_pos) {
						$('html, body').animate({
							scrollTop: stats_pos
						}, 500);
					}
					$(this).trigger('clickload');
				});
				tbody.append(tr);
				var td = $('<td>' + vals.Date + '</td>');
				tr.append(td);
				var td = $('<td class="text-right" data-rx="' + vals.RX + '">' + humanFileSize(vals.RX) + '</td>');
				tr.append(td);
				var td = $('<td class="text-right" data-tx="' + vals.TX + '">' + humanFileSize(vals.TX) + '</td>');
				tr.append(td);
				var td = $('<td class="text-right" data-total="' + (vals.RX + vals.TX) + '">' + humanFileSize(vals.RX + vals.TX) + '</td>');
				tr.append(td);
			});

			$('#MonthStats' + nwif).append(table);

			if (fresh_start) {
				hideMore('#MonthStats' + nwif, 500, 'view all months');
			}

			var clicked_row = false;

			if (highlight) {
				clicked_row = tbody.find('tr.' + highlight).first();
			}

			if (!clicked_row) {
				clicked_row = tbody.find('tr').first();
			}

			clicked_row.trigger('clickload');

			if (fresh_start) {
				// auto-refresh
				setInterval(function() {
					var cur_selected_month = $('#MonthStats' + nwif).find('tr.table-active').first().data('month');
					loadMonthlyStats(nwif, cur_selected_month);
				}, 60000);
			}
		});
	}

	function loadDaily(nwif, month) {
		var MyDate = new Date();
		var cur_date = MyDate.getFullYear() + '-' + ('0' + (MyDate.getMonth()+1)).slice(-2) + '-' + ('0' + MyDate.getDate()).slice(-2);

		$.getJSON('/stats/' + nwif + '/' + month, function (data) {
			$('#DayStats' + nwif).empty();
			$('#DayStats' + nwif).data('month', month);
			var table = $('<table>', { class: 'table table-bordered table-dark table-sm' });
			var thds = $('<tr><th>Day</th><th class="text-right">Downloaded</th><th class="text-right">Uploaded</th><th class="text-right">Total</th></tr>')
			table.append(thds);
			$.each(data, function (idx, vals) {
				var tr = $('<tr>');
				if (vals.Date == cur_date) {
					tr.addClass('live-stat');
				}
				table.append(tr);
				var td = $('<td>' + vals.Date + '</td>');
				tr.append(td);
				var td = $('<td class="text-right" data-rx="' + vals.RX + '">' + humanFileSize(vals.RX) + '</td>');
				tr.append(td);
				var td = $('<td class="text-right" data-tx="' + vals.TX + '">' + humanFileSize(vals.TX) + '</td>');
				tr.append(td);
				var td = $('<td class="text-right" data-total="' + (vals.RX + vals.TX) + '">' + humanFileSize(vals.RX + vals.TX) + '</td>');
				tr.append(td);
			});

			$('#DayStats' + nwif).append(table);
		});
	}


	function hideMore(el, max_height, message = "View all") {
		var e = $(el);
		if (!e.length || e.height() < max_height) {
			return;
		}
		e.addClass('truncated');
		e.height(max_height);
		var view_more = $('<div class="view-more"></div>');
		var l = $('<a href="#">' + message + ' &DownArrowBar;</a>');
		l.on('click', function (ev) {
			var tr = $(this).parent().parent().find('.truncated');
			tr.height('auto');
			tr.removeClass('truncated');
			$(this).parent().remove();
			return false;
		});
		view_more.append(l);
		e.after(view_more);
	}

});
