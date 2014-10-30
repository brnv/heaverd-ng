var heaverd = {

	Charts: function(selector) {
		if (selector != undefined) {
			heaverd._selector = selector;
		}
		heaverd._fetchData();
	},

	_fetchData: function() {
		$.ajax({
			url: "/v2/h",
			success: function(hosts) {
				heaverd._render(hosts);
			}});
		setTimeout(function() {
			heaverd._fetchData()
		}, 3000);
	},

	_render: function(hosts) {
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
				poolsData.push({name: host.Hostname + " (" +
					Object.keys(host.Containers).length+")",
					y: host.Score/totalScore*100,
					color:heaverd._colors[colorIndex],});

				weights = {
					"cpu": [host.CpuWeight, "CPU " +
						(host.CpuCapacity-host.CpuUsage).toFixed(2)+"%"],
					"ram": [host.RamWeight, "RAM " +
						(host.RamFree/1024/1024).toFixed(2)+" GiB"],
					"disk": [host.DiskWeight, "HDD " +
						(host.DiskFree/1024/1024).toFixed(2)+" GiB"]
				};

				j = 0;
				$.each(weights,
						function(param, value) {
							wl = Object.keys(weights).length
								brightness = 1/wl-(j/wl)/wl;
							weightsData.push({
								name: value[1],
								y: host.Score/totalScore*100/(host.CpuWeight+
										host.RamWeight+
										host.DiskWeight)*value[0],
								color: Highcharts.Color(
										heaverd._colors[colorIndex]).
									brighten(brightness).get()
							});
							j += 1;
						});
				colorIndex += 1;
				if (colorIndex == heaverd._colors.length) {
					colorIndex = 0;
				}
			});

			if ($('div#chart-'+poolname).length == 0) {
				$(heaverd._selector).append('<div id="chart-'+ poolname+'"></div><br/>');
				heaverd._charts[poolname] = new Highcharts.Chart({
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
					series: [{
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
					}]
				});
			} else {
				heaverd._charts[poolname].series[0].setData(poolsData, true);
				heaverd._charts[poolname].series[1].setData(weightsData, true);
				heaverd._charts[poolname].redraw();
			}
		});
	},

	_colors: [
		"#457B34"
	],

	_charts: {},

	_selector: "#charts",
}
