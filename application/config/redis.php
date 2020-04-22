<?php if (!defined('BASEPATH')) exit('No direct script access allowed');

/********************************程序应用配置***************************************/
$util_redis_conf['userinfo']['host'] = '127.0.0.1';
$util_redis_conf['userinfo']['port'] = 6381;

if (GAME_ZONE_ID == 1) {
    $util_redis_conf['default']['host'] = '127.0.0.1';
    $util_redis_conf['default']['port'] = 6379;

    $util_redis_conf['pika']['host'] = '127.0.0.1';
    $util_redis_conf['pika']['port'] = 9221;
} else if (GAME_ZONE_ID == 2) {
    $util_redis_conf['default']['host'] = '127.0.0.1';
    $util_redis_conf['default']['port'] = 6380;

    $util_redis_conf['pika']['host'] = '127.0.0.1';
    $util_redis_conf['pika']['port'] = 9222;
}

