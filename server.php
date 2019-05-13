<?php

use EasySwoole\Actor\Actor;

date_default_timezone_set('Asia/Shanghai');
header('Content-Type: text/html; charset=UTF-8');
ini_set('display_errors', 'On');
error_reporting(E_ERROR);

define("SERVER_NAME", 'TT');

define("APPPATH", realpath(dirname(__FILE__) . '/'));
define("BASEPATH", APPPATH . '/system');
define("APPROOT", APPPATH . '/application');

include(BASEPATH . "/Application.php");

//创建WebSocket
$ws = new swoole_websocket_server("0.0.0.0", 9508,  SWOOLE_PROCESS, SWOOLE_SOCK_TCP | SWOOLE_SSL);

$ws->set(array(
        'daemonize' => false,	//是否是守护进程
        'worker_num' => 4,		//工作者进程数
        'max_request' => 10000,	//最大连接数量
        'dispatch_mode' => 2,
        'debug_mode'=> 1,
        'ssl_key_file' => '/etc/nginx/ssl/game.fx4j.com/privkey.pem',
        'ssl_cert_file' => '/etc/nginx/ssl/game.fx4j.com/fullchain.pem',
        'heartbeat_check_interval' => 5,
        'heartbeat_idle_time' => 600,
));

Actor::getInstance()->attachToServer($ws);

//监听WebSocket连接打开事件
$ws->on('WorkerStart', function ($ws, $request) {
	Userfd::getInstance($ws);
});

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
        if(!empty($input['userid'])) { //绑定 userid 到 fd
            $ws->bind($frame->fd, $input['userid']);
            Userfd::getInstance()->set($input['userid'], $frame->fd);
        }
        
        $application = new Application($frame->fd);
        
        $result = $application->run($input, $ws->getClientInfo($frame->fd), $ws);
        unset($application);

        if (empty($result)) {
        	//无返回
        } else if ($result['send_user'] === 'all') {
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
});

//监听WebSocket连接关闭事件
$ws->on('close', function ($ws, $fd) {
    $client_info = $ws->getClientInfo($fd);
    $uid = $client_info['uid'];
    Userfd::getInstance()->del($uid);
    RoomLogic::getInstance()->proxyGame($uid);
});

echo "\n初始化成功\n\n";

$ws->start();
