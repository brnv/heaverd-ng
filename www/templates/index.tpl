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

			jQuery.each(hosts, function(hostname, info) {
				option = {
					count:Object.keys(info.Containers).length,
					name: /(\w+)\./g.exec(hostname)[1],
					y: info.Score
				};
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
					indexLabel: "({count}) {name} #percent%",
					dataPoints: options
				}],
				animationEnabled: false,
			})
			score.render();
		}
	</script>
</body>
</html>
