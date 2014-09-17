<!DOCTYPE html>
<html>
<head></head>
<body style="margin: 50px auto; width: 500px; text-align: center;">
	<div id="charts">
	</div>

	<script src="https://code.jquery.com/jquery-2.1.1.min.js"></script>
	<script src="http://www.chartjs.org/assets/Chart.min.js"></script>

	<script type="text/javascript" charset="utf-8">
		$.ajax({
			url: "http://container.s:8081/v2/h",
			success: function(data){
				hosts = $.parseJSON(data);
				render(hosts);
			}
		});

		function render(hosts) {
			pools = {};

			$.each(hosts, function(hostname, info) {
				$.each(info.Pools, function(key, poolname) {
					if (!pools[poolname]) {
						pools[poolname] = [];
					}
					pools[poolname].push(info);
				});
			});

			var colorIndex = 0;

			var colors= [
				"#3D5527",
				"#4A6A2B",
				"#8A2F65",
				"#8F2F3D",
				"#934E2F"
			];

			$.each(pools, function(poolname, hosts) {
				$("#charts").append('<canvas id="'+
					poolname+'" width="500" height="300"></canvas>'+
					'<p>'+poolname+' pool</p><p id="'+poolname+'"></p></br>')

				data = []
				options = {
					animation:false,
					showTooltips:false,
					legendTemplate : "<ul class=\"<%=name.toLowerCase()%>-legend\"><% for (var i=0; i<segments.length; i++){%><li><span style=\"background-color:<%=segments[i].lineColor%>\"></span><%if(segments[i].label){%><%=segments[i].label%><%}%></li><%}%></ul>"
				}

				$.each(hosts, function(key, host) {
					data.push({
						value: host.Score,
						color: colors[colorIndex],
						label: "("+Object.keys(host.Containers).length+") "+host.Hostname,
					});

					var chart = new Chart($("#"+poolname).
						get(0).getContext("2d")).Pie(data, options);
						console.log(chart)
					$('p#'+poolname).html(chart.generateLegend())

					colorIndex = colorIndex + 1;
					if (colorIndex == colors.length) {
						colorIndex = 0;
					}
				})

			});
		}
	</script>
</body>
</html>
