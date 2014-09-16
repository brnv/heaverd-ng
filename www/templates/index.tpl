<!DOCTYPE html>
<html>
<head></head>
<body>
	<div id="scoreContainer" style="height: 100%;"></div>

	<script src="https://code.jquery.com/jquery-2.1.1.min.js"></script>
	<script src="/js/canvasjs.min.js" type="text/javascript" charset="utf-8"></script>

	<script type="text/javascript" charset="utf-8">
		jQuery.ajax({
			url: "http://yaci.yard.s:8081/v2/h",
			success: function(data){
				hosts = jQuery.parseJSON(data);
				render(hosts);
			}
		});

		function render(hosts) {
			options = [];
			weightThreshold = 0.4;

			jQuery.each(hosts, function(hostname, info) {
				option = {
					count:Object.keys(info.Containers).length,
					name: /(\w+)\./g.exec(hostname)[1],
					y: info.Score,
					messages: [],
				};
				if (info.CpuWeight < weightThreshold) {
					option.messages.push("cpu")
				}
				if (info.DiskWeight < weightThreshold) {
					option.messages.push("disk")
				}
				if (info.RamWeight < weightThreshold) {
					option.messages.push("ram")
				}
				if (option.messages.length > 0) {
					option.messages[0] = "(low: " + option.messages[0]
					option.messages[option.messages.length-1] =
						option.messages[option.messages.length-1] + ")"
				}

				options.push(option);
			})

			var score = new CanvasJS.Chart("scoreContainer",
			{
				title:{
					text: "Hosts score",
				},
				legend: {
					verticalAlign: "bottom",
					horizontalAlign: "center"
				},
				data: [{
					type: "pie",
					toolTipContent: "{name}: {y}",
					indexLabel: "({count}) {name} {messages} #percent%",
					dataPoints: options
				}],
				animationEnabled: false,
			})
			score.render();
		}
	</script>
</body>
</html>
