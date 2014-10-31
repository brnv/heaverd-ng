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

	_getTotalScore: function(hosts) {
		totalScore = 0;
		$.each(hosts, function(key, host) {
			if (host.Score <= heaverd._scoreThreshold) {
				host.Score = heaverd._scoreThreshold;
			}
			totalScore += host.Score;
		});
		return totalScore;
	},

	_getPools: function (hosts) {
		pools = {};
		$.each(hosts, function(hostname, info) {
			$.each(info.Pools, function(key, poolname) {
				if (!pools[poolname]) {
					pools[poolname] = [];
				}
				pools[poolname].push(info);
			});
		});
		return pools;
	},

	_render: function(hosts) {
		$.each(heaverd._getPools(hosts), function(poolname, hosts) {
			poolsData = [];
			weightsData = [];

			totalScore = heaverd._getTotalScore(hosts);

			$.each(hosts, function(key, host) {
				color = heaverd._colors.ok;
				if (host.Score <= heaverd._scoreThreshold) {
					host.Score = heaverd._scoreThreshold;
					color = heaverd._colors.warning;
				}

				poolsData.push(
						{name: host.Hostname + " (" +
							Object.keys(host.Containers).length+")",
							y: host.Score/totalScore *100,
							color:color,
						});

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
							color = Highcharts.Color(heaverd._colors.ok).
								brighten(brightness).get();
							name = value[1];
							if (value[0] <= heaverd._scoreThreshold) {
								value[0] = 0.5; //to see small sector in drawn chart
								color = heaverd._colors.warning;
								name = '<span class="warning">' + value[1] + '</span>';
							}
							//this assignments is for seeing small sector in drawn chart
							if (host.CpuWeight == 0) {
								host.CpuWeight = 0.5; 
							}
							if (host.RamWeight == 0) {
								host.RamWeight = 0.5;
							}
							if (host.DiskWeight == 0) {
								host.DiskWeight = 0.5;
							}
							weightsData.push({
								name: name,
								y: host.Score/totalScore*100/(host.CpuWeight+
										host.RamWeight+
										host.DiskWeight)*value[0],
								color: color,
							});
							j += 1;
						});
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
							distance: -50,
						}
					}, {
						animation: false,
						data: poolsData,
						data: weightsData,
						size: '100%',
						innerSize: '85%',
						dataLabels: {
							useHTML: true,
						}
					}]
				});
			} else {
				heaverd._charts[poolname].series[0].setData(poolsData, true);
				heaverd._charts[poolname].series[1].setData(weightsData, true);
				heaverd._charts[poolname].redraw();
			}
		});
	},

	_colors: {
		ok: "#457B34",
		warning: "red",
	},

	_charts: {},
	_selector: "#charts",
	_scoreThreshold: 0.05,
	_resourceThreshold: 0.4,
}
