<?php
date_default_timezone_set('Asia/Shanghai');
header('Content-Type: text/html; charset=UTF-8');
ini_set('display_errors', 'On');
error_reporting(E_ERROR);

define("APPPATH", realpath(dirname(__FILE__) . '/'));
define("BASEPATH", APPPATH . '/system');
define("APP_ROOT", APPPATH . '/application');

//根据游戏不同区设置端口
$zone_id = isset($_SERVER['argv'][1]) ? $_SERVER['argv'][1] : 1;
if ($zone_id <= 1) {
    define('DABAOJIAN_ZONE_ID', 1);
    $port = 9509;
} else if ($zone_id == 2) {
    define('DABAOJIAN_ZONE_ID', 2);
    $port = 9510;
}

include(BASEPATH . "/Application.php");

//创建WebSocket
$ws = new swoole_websocket_server("0.0.0.0", $port,  SWOOLE_PROCESS, SWOOLE_SOCK_TCP | SWOOLE_SSL);

$ws->set(
    array(
        'daemonize' => false,	// 是否是守护进程
        'worker_num' => 1,		//工作者进程数
        'max_request' => 10000,	// 最大连接数量
        'dispatch_mode' => 2,
        'debug_mode'=> 1,
        'ssl_key_file' => '/etc/nginx/ssl/game.fx4j.com/privkey.pem',
        'ssl_cert_file' => '/etc/nginx/ssl/game.fx4j.com/fullchain.pem',
        'heartbeat_check_interval' => 5,
        'heartbeat_idle_time' => 600,
    )
);

//监听WebSocket连接打开事件
$ws->on('open', function ($ws, $request) {
});

//监听WebSocket消息事件，其他：swoole提供了bind方法，支持uid和fd绑定
$ws->on('message', function($ws, $frame) {
    if ($frame->data == 'heartbeat') { //客户端心跳
        $ws->push($frame->fd, time());
        return;
    }

    $input = json_decode($frame->data, true);
    if (empty($input) || empty($input['c']) || empty($input['m'])) { //输入格式错误
        $ws->push($frame->fd, json_encode(array("tagcode" => "1", "description" => "input error", "data" => $frame->data)));
    } else {
        $application = new Application($frame->fd);
        $result = $application->run($input, $ws->getClientInfo($frame->fd));
        unset($application);

        if ($result['send_user'] === 'all') {
            $start_fd = 0;
            while (true) {
                //分批从 connection_list 函数获取现在连接中的fd，每轮100个
                $conn_list = $ws->connection_list($start_fd, 100);
                //var_dump($conn_list);

                if ($conn_list === false || count($conn_list) === 0) {
                    return;
                }

                $start_fd = end($conn_list);

                foreach($conn_list as $fd) {
                    $ws->push($fd, $result['msg']);
                }
            }
        } else if (is_array($result['send_user']) && !empty($result['send_user'])) {
            foreach($result['send_user'] as $fd) {
                $ws->push($fd, $result['msg']);
            }
        }
    }
}
       );

//监听WebSocket连接关闭事件
$ws->on('close', function ($ws, $fd) {
    $ws->close($fd);   // 销毁fd链接信息
}
       );

echo "\n初始化成功\n\n";

$ws->start();
