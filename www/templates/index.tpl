<!DOCTYPE html>
<html>
<head>
<style>
body {
	margin: 50px auto;
	width: 700px;
	text-align: center;
}
span.warning {
	font-weight: bold;
	color: white;
	padding: 3px;
	background-color: red;
}
</style>
</head>
<body >
<div id="charts"></div>
<script type="text/javascript" charset="utf-8">
window.onload = function() {
	heaverd.Charts("#charts");
}
</script>
<script src="/static/js/jquery-2.1.1.min.js"></script>
<script src="/static/js/highcharts.js"></script>
<script src="/static/js/heaverd-ng-score.js"></script>
</body>
</html>
