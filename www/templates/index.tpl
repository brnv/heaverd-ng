<!DOCTYPE html>
<html>
<head>
    <title></title>
    <meta charset="utf-8" />
<script type="text/javascript">
	window.onload = function () {
		var score = new CanvasJS.Chart("scoreContainer",
		{
			title:{
				text: "Hosts score",
			},
			legend: {
				verticalAlign: "bottom",
				horizontalAlign: "center"
			},
			data: [
				{
					type: "pie",
					toolTipContent: "{name}: {y}",
					indexLabel: "{name} #percent%", 
					dataPoints: {{.Score}}
				}
			],
			animationEnabled: false,
		});
		score.render();
	}
</script>
<script src="/js/canvasjs.min.js" type="text/javascript" charset="utf-8"></script>
</head>
<body>
    <content>
		<div id="scoreContainer" style="height: 300px; width: 100%;"></div>
    </content>
</body>
</html>
