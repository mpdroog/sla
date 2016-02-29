#!/usr/bin/php
<?php
/*
 * Munin-plugin to collect SLA/download-statistics.
 *
 * @arg argv1 Optional. Meta-data, allowed=[default,config]
 * @author jvginkel <jethro@itshosted.nl>
 */
define("JSON_STAT", "/tmp/sla.download.json.tmp");

if (@php_sapi_name() !== 'cli') {
	echo "UNKNOWN - CLI tool only!\n";
	exit(1);
}

$type = isset($argv[1]) ? strtolower($argv[1]) : "default";
if (! in_array($type, ["default", "config"])) {
	echo sprintf("UNKNOWN - Argument unsupported: %s\n", $type);
	exit(1);
}

if ($type === "config") {
	echo "graph_title SLA Download speed results
graph_scale no
graph_category sla
Conn.label Connect time in ms
Conn.min 0
Auth.min 0
Auth.label Auth time in ms
Arts.label Avg arts download time in ms
Arts.min 0
graph_info sla download timings in ms
";
	exit(0);
}

$stat = @json_decode(file_get_contents(JSON_STAT), true);
if (! is_array($stat)) {
	echo sprintf("UNKNOWN - Could not read %s\n", JSON_STAT);
	exit(1);
}
echo sprintf("Conn.value %s\n", $stat['Conn']);
echo sprintf("Auth.value %s\n", $stat['Auth']);

// Calc average replytime per article
{
	$arts = count($stat['Arts']);
	$sum = 0;
	foreach ($stat['Arts'] as $art) {
		$sum += $art;
	}

	$artsec = 0;
	if ($arts > 0) {
		$artsec = round($sum/$arts, 2);
	}
	echo sprintf("Arts.value %s\n", $artsec);
}
