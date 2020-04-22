<?php
//status 1-正常 2-即将启动 3-维护
$config['zone_info'] = [
    ['zone_id' => 1, 'zone_name' => '世外桃园', 'hot' => '爆满', 'status' => 1, 'socket' => 's1', 'port' => 9507, 'time' => '2020-03-28 10:00:00'],
    ['zone_id' => 2, 'zone_name' => '开天辟地', 'hot' => '流畅', 'status' => 1, 'socket' => 's2', 'port' => 9508, 'time' => '2020-04-07 10:00:00'],
];

//推荐区
$config['recommend_zone'] = [2, 1];

return $config;
