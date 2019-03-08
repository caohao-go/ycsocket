<?php
date_default_timezone_set('Asia/Shanghai');
header('Content-Type: text/html; charset=UTF-8');

//错误显示级别
ini_set('display_errors', 'On');
error_reporting(E_ERROR);

define("APPPATH", realpath(dirname(__FILE__) . '/'));
define("BASEPATH", APPPATH . '/system');
define("APP_ROOT", APPPATH . '/application');

include(BASEPATH . "/Application.php");

//根据分区设置端口
$zone_id = isset($_SERVER['argv'][1]) ? $_SERVER['argv'][1] : 1;
if($zone_id <= 1) {
    define('DABAOJIAN_ZONE_ID', 1);
    $port = 9509;
} else if($zone_id == 2) {
    define('DABAOJIAN_ZONE_ID', 2);
    $port = 9510;
}

echo "\n初始化成功\n\n";

//创建Websocket Reactor 
$ws = new swoole_websocket_server("0.0.0.0", $port,  SWOOLE_PROCESS, SWOOLE_SOCK_TCP | SWOOLE_SSL);

//设置配置
$ws->set(
    array(
        'daemonize' => false,      // 是否是守护进程
        'worker_num' => 1,  //工作者进程数
        'max_request' => 10000,    // 最大连接数量
        'dispatch_mode' => 2,
        'debug_mode'=> 1,
		//'ssl_key_file' => '/etc/letsencrypt/live/api.gaoqu.site/privkey.pem',
		//'ssl_cert_file' => '/etc/letsencrypt/live/api.gaoqu.site/fullchain.pem',
		'ssl_key_file' => '/etc/nginx/ssl/game.fx4j.com/privkey.pem',
		'ssl_cert_file' => '/etc/nginx/ssl/game.fx4j.com/fullchain.pem',
        // 心跳检测的设置，自动踢掉掉线的fd
        'heartbeat_check_interval' => 5,
        'heartbeat_idle_time' => 600,
    )
);

//监听WebSocket连接打开事件
$ws->on('open', function ($ws, $request) {
});

//监听WebSocket消息事件，其他：swoole提供了bind方法，支持uid和fd绑定
$ws->on('message', $handle_function);

//监听WebSocket连接关闭事件
$ws->on('close', function ($ws, $fd) {
    $ws->close($fd);   // 销毁fd链接信息
});

$ws->start();
