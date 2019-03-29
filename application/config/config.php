<?php  if ( ! defined('BASEPATH')) exit('No direct script access allowed');

/********************************程序应用配置***************************************/

$config['zone_info'] = [
    ['zone_id' => 1, 'zone_name' => '一区', 'hequ' => 1, 'hot' => '爆满', 'status' => 1, 'socket' => 'wss0', 'port' => 9504, 'time' => '2018-11-08 10:00:00'],
    ['zone_id' => 2, 'zone_name' => '二区', 'hequ' => 1, 'hot' => '爆满', 'status' => 1, 'socket' => 'wss0', 'port' => 9504, 'time' => '2018-12-08 10:00:00'],
    ['zone_id' => 3, 'zone_name' => '三区', 'hequ' => 1, 'hot' => '爆满', 'status' => 1, 'socket' => 'wss0', 'port' => 9504, 'time' => '2018-12-08 10:00:00'],
    ['zone_id' => 4, 'zone_name' => '四区', 'hequ' => 0, 'hot' => '新服', 'status' => 1, 'socket' => 'wss4', 'port' => 9507, 'time' => '2019-01-16 10点'],
];

$config['source_zone'] = [
    1 => 1,
    2 => 1,
    3 => 1,
    4 => 0,
];

return $config;