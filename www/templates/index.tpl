<!DOCTYPE html>
<html>
<head></head>
<body style="margin: 50px auto; width: 700px; text-align: center;">
	<div id="charts"></div>

	<script src="https://code.jquery.com/jquery-2.1.1.min.js"></script>
	<script src="http://code.highcharts.com/highcharts.js"></script>

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

			colorIndex = 0;

			$.each(pools, function(poolname, hosts) {
				$("#charts").append('<div id="'+ poolname+'"></div><br/>');

				poolsData = [];
				weightsData = [];

				totalScore = 0;
				$.each(hosts, function(key, host) {
					totalScore += host.Score;
				});

				$.each(hosts, function(key, host) {
					poolsData.push(
						{
							name: host.Hostname,
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

				$('#'+poolname).highcharts({
					chart: {
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
			});
		}

		setTimeout(function() {
			location.reload();
		}, 10000);
	</script>
</body>
</html>
