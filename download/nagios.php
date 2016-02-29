#!/usr/bin/php
<?php
/*
 * Nagios-plugin to analyze JSON-output from SLA/download
 * and report if something is troublesome.
 *
 * What is reported?
 * - Average download speed
 * - Errors reported by sla/download
 *
 * @author mpdroog <mark@itshosted.nl>
 */
define("JSON_STAT", "/tmp/sla.download.json.tmp");
define("MINSPEED", 100); // 100KB/s

if (@php_sapi_name() !== 'cli') {
	echo "UNKNOWN - CLI tool only!\n";
	exit(1);
}

$stat = @json_decode(file_get_contents(JSON_STAT), true);
if (! is_array($stat)) {
	echo sprintf("UNKNOWN - Could not read %s\n", JSON_STAT);
	exit(1);
}

if (count($stat["Error"]) > 0) {
	echo sprintf("CRITICAL - %s\n", implode(", ", $stat["Error"]));
	exit(2);
}
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
	if ($artsec < MINSPEED) {
		echo sprintf("CRITICAL - Average download speed %d KB/s\n", $artsec);
		exit(2);
	}
}

echo "OK - SLA/DOWNLOAD\n";
exit(0);
