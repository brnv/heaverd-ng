var heaverd = {

	Charts: function(selector) {
		if (selector == undefined) {
			selector = heaverd._selector;
		}
		var draw = function() {
			heaverd._fetchData(function(hosts) {
				heaverd._render(hosts, selector);
			});
		}
		draw();
		setInterval(function() {
			draw();
		}, 3000);
	},

	_render: function(hosts, selector) {
		$.each(heaverd._getPools(hosts), function(poolname, hosts) {
			var poolsData = heaverd._getPoolsData(hosts);
			var weightsData = heaverd._getWeightsData(hosts);
			if ($('div#chart-'+poolname).length == 0) {
				$(selector).append(
						heaverd._chartHtmlTemplate.replace("%ID%", poolname));
				heaverd._charts[poolname] =	heaverd._getNewHighchart(
						poolname, poolsData, weightsData);
			} else {
				heaverd._charts[poolname].series[0].setData(poolsData, true);
				heaverd._charts[poolname].series[1].setData(weightsData, true);
				heaverd._charts[poolname].redraw();
			}
		});
	},

	_getNewHighchart: function(poolname, poolsData, weightsData) {
		return new Highcharts.Chart({
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
	},

	_getWeightsData: function(hosts) {
		var weightsData = [];
		$.each(hosts, function(key, host) {
			var hostScore = host.Score;
			if (host.Score <= heaverd._scoreThreshold) {
				hostScore = heaverd._scoreThreshold;
			}
			var weights = {
				"cpu": [host.CpuWeight, "CPU " +
					(host.CpuCapacity-host.CpuUsage).toFixed(2)+"%"],
				"ram": [host.RamWeight, "RAM " +
					(host.RamFree/1024/1024).toFixed(2)+" GiB"],
				"disk": [host.DiskWeight, "HDD " +
					(host.DiskFree/1024/1024).toFixed(2)+" GiB"],
				"ioawait": [host.DiskIOWeight, "IO await " +
					(host.IostatAwait).toFixed(2)+" ms"]
			};
			var j = 0;
			$.each(weights, function(param, value) {
				var wl = Object.keys(weights).length
					brightness = 1/wl-(j/wl)/wl;
				var color = Highcharts.Color(heaverd._colors.ok).
					brighten(brightness).get();
				var name = value[1];
				var chartSectorSize  = value[0];
				if (value[0] <= heaverd._scoreThreshold) {
					chartSectorSize = heaverd._warningSectorSize;
					color = heaverd._colors.warning;
					name = '<span class="warning">' + value[1] + '</span>';
				}
				var cpuWeight = (host.CpuWeight > 0) ?
					host.CpuWeight : heaverd._warningSectorSize;
				var ramWeight = (host.RamWeight > 0) ?
					host.RamWeight : heaverd._warningSectorSize;
				var diskWeight= (host.DiskWeight > 0) ?
					host.DiskWeight : heaverd._warningSectorSize;
				var diskIOWeight= (host.DiskIOWeight > 0) ?
					host.DiskIOWeight : heaverd._warningSectorSize;
				var chartValue = hostScore / heaverd._getTotalScore(hosts) *
					100 / (cpuWeight + ramWeight + diskWeight + diskIOWeight) *
					chartSectorSize;
				weightsData.push({
					name: name,
					y: chartValue,
					color: color,
				});
				j += 1;
			}
			);
		});
		return weightsData;
	},

	_getPoolsData: function(hosts) {
		var poolsData = [];
		$.each(hosts, function(key, host) {
			var color = heaverd._colors.ok;
			var hostScore = host.Score;
			if (host.Score <= heaverd._scoreThreshold) {
				hostScore = heaverd._scoreThreshold;
				color = heaverd._colors.warning;
			}
			var chartValue = hostScore/heaverd._getTotalScore(hosts)* 100;
			poolsData.push(
					{
						name: host.Hostname + " (" +
							Object.keys(host.Containers).length+")",
						y: chartValue,
						color:color,
					});
		});
		return poolsData;
	},

	_getPools: function (hosts) {
		var pools = {};
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

	_getTotalScore: function(hosts) {
		var totalScore = 0;
		$.each(hosts, function(key, host) {
			totalScore += (host.Score <= heaverd._scoreThreshold) ?
				heaverd._scoreThreshold : host.Score;
		});
		return totalScore;
	},

	_fetchData: function(callback) {
		$.ajax({
			url: "/v2/h",
			success: function(data) {
				callback(data);
			}});
	},

	_charts: {},

	_colors: {
		ok: "#457B34",
		warning: "#D00000",
	},

	_selector: "#charts",

	_scoreThreshold: 0.05,

	_warningSectorSize: 0.3,

	_chartHtmlTemplate: '<div id="chart-%ID%"></div><br/>',
}
