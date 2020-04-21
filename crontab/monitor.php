<?php
date_default_timezone_set('Asia/Shanghai');
header('Content-Type: text/html; charset=UTF-8');
ini_set('display_errors', 'On');
error_reporting(E_ERROR); /*E_ALL,E_ERROR*/

define("APPPATH", realpath(dirname(__FILE__) . '/../'));

exec("ps -ef | grep 'monitor.php' | grep -v 'grep'", $out);
if(count($out) > 1) {
	die("monitor is running\n");
}

while(1) {
	$zoneinfos_array = require(APPPATH . "/application/config/zoneinfo.php");
	$zoneinfos = $zoneinfos_array['zone_info'];
	unset($zoneinfos_array);
	
	foreach($zoneinfos as $zoneinfo) {
		$zone_id = $zoneinfo['zone_id'];
		
		if($zoneinfo['status'] != 1) {
			continue;
		}
		
		$out = null;
		$data = exec("ps -ef | grep 'co_server.php {$zone_id}' | grep -v 'grep'", $out);
		if(count($out) == 0) {
			exec("/usr/bin/php /data/app/super_server/co_server.php {$zone_id} > /dev/null &", $ret);
		}
	}
	
	sleep(1);
}
