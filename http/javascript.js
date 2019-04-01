$(function () {
	function connect() {
		var ws = new WebSocket("ws://" + document.location.host + "/stream");

		ws.onmessage = function (event) {
			var obj = JSON.parse(event.data);
			$("#output").empty();
			table = $('<table>', { 'class': 'table' });
			table.append('<tr><th>Interface</td><th>Download</th><th>Upload</th></tr>');
			$(obj).each(function (k, nw) {
				table.append('<tr><td>' + nw.If + '</td><td>' + nw.Rx + '</td><td>' + nw.Tx + '</td></tr>');
			})
			$("#output").append(table)
		}

		ws.onclose = function (e) {
			// console.log('Socket is closed. Reconnect will be attempted in 1 second.', e.reason);
			setTimeout(function () {
				connect();
			}, 5000);
		};

		ws.onerror = function (err) {
			// console.error('Socket encountered error: ', err.message, 'Closing socket');
			ws.close();
		}
	}

	connect();
});
