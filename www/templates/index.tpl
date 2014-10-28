<!DOCTYPE html>
<html>
<head></head>
<body style="margin: 50px auto; width: 700px; text-align: center;">

	<div id="charts"></div>

	<script src="/static/js/jquery-2.1.1.min.js"></script>
	<script src="/static/js/highcharts.js"></script>
	<script type="text/javascript" charset="utf-8">
		var fetchData = function() {
			$.ajax({
				url: "/v2/h",
				success: function(hosts) {
					render(hosts);
				}
			});
			setTimeout(function() {
				fetchData()
			}, 5000);
		}

		fetchData();

		colors = [
			"#416E32",
			"#A73853",
			"#457B34",
			"#AA3938",
			"#498736",
			"#AC5638",
			"#AF7538",
			"#4D9437",
			"#B29538",
		];

		charts = {};

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

			colorIndex = 0;

			$.each(pools, function(poolname, hosts) {
				poolsData = [];
				weightsData = [];

				totalScore = 0;
				$.each(hosts, function(key, host) {
					if (host.Score == 0) {
						return;
					}
					totalScore += host.Score;
				});

				$.each(hosts, function(key, host) {
					if (host.Score == 0) {
						return;
					}
					poolsData.push(
						{
							name: host.Hostname + " ("+Object.keys(host.Containers).length+")",
							y: host.Score/totalScore*100,
							color:colors[colorIndex],
						}
					);

					weights = {
						"cpu": [host.CpuWeight, "CPU "+(host.CpuCapacity-host.CpuUsage).toFixed(2)+"%"],
						"ram": [host.RamWeight, "RAM "+(host.RamFree/1024/1024).toFixed(2)+" GiB"],
						"disk": [host.DiskWeight, "HDD "+(host.DiskFree/1024/1024).toFixed(2)+" GiB"]
					};

					j = 0;
					$.each(weights,
						function(param, value) {
							wl = Object.keys(weights).length
							brightness = 1/wl-(j/wl)/wl;
							weightsData.push({
									name: value[1],
									y: host.Score/totalScore*100/(host.CpuWeight+host.RamWeight+host.DiskWeight)*value[0],
									color: Highcharts.Color(colors[colorIndex]).brighten(brightness).get()
							});
							j += 1;
						}
					);
					colorIndex += 1;
					if (colorIndex == colors.length) {
						colorIndex = 0;
					}
				});

				if ($('div#chart-'+poolname).length == 0) {
					$("#charts").append('<div id="chart-'+ poolname+'"></div><br/>');

					charts[poolname] = new Highcharts.Chart({
						chart: {
							renderTo: 'chart-'+poolname,
							type: 'pie'
						},
						tooltip: {
							enabled: false
						},
						title: {
							text: poolname+' pool'
						},
						series: [
							{
								animation: false,
								data: poolsData,
								size: '85%',
								dataLabels: {
									color: 'white',
									distance: -50
								}
							}, {
								animation: false,
								data: poolsData,
								data: weightsData,
								size: '100%',
								innerSize: '85%'
							}
						]
					});
				} else {
					charts[poolname].series[0].setData(poolsData, true);
					charts[poolname].series[1].setData(weightsData, true);
					charts[poolname].redraw();
				}
			});
		}
	</script>
</body>
</html>
