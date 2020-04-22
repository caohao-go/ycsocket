<?php
date_default_timezone_set('Asia/Shanghai');
header('Content-Type: text/html; charset=UTF-8');
ini_set('display_errors', 'On');
error_reporting(E_ERROR); /*E_ALL,E_ERROR*/

define("ENABLE_COROUTINE", true);

define('LOG_PATH', '/data/app/logs/super_server'); //日志目录
define('PHP_LOG_THRESHOLD', 1); //错误记录级别 ERROR=1, DEBUG=2, WARNING=3, NOTICE=4, INFO=5, ALL=6

define("APPPATH", realpath(dirname(__FILE__) . '/'));
define("BASEPATH", APPPATH . '/system');
define("APP_ROOT", APPPATH . '/application');

Swoole\Runtime::enableCoroutine(SWOOLE_HOOK_ALL | SWOOLE_HOOK_CURL);

//区号
$input_zone_id = intval($argv[1]);
if ($input_zone_id < 1) {
    die("请输入区号\n");
}

//根据游戏不同区设置端口
define('GAME_ZONE_ID', $input_zone_id);  //游戏区
$zoneinfo = require(APP_ROOT . "/config/zoneinfo.php");
$port = $zoneinfo['zone_info'][$input_zone_id - 1]['port'];

include(BASEPATH . "/Application.php");

$sch = new Co\Scheduler;
$sch->set(['max_coroutine' => 100000, 'enable_preemptive_scheduler' => true]);
$sch->add(function () use ($port) {
    echo "监听端口:" . $port . "\n";

    /// 初始化代码可以写在这里 ///
    Zoneinfo::update();
    Zoneinfo::game_version();
    Task::init_task();
    /////////// end //////////

    //创建WebSocket
    $ws = new Co\Http\Server("0.0.0.0", $port, true);
    
    $ws->set([
	    'ssl_cert_file' => '/data/ssl/fullchain1.pem',
	    'ssl_key_file' => '/data/ssl/privkey1.pem',
	]);

    $ws->handle('/', function ($request, $ws) {
        try {
            if (strtolower($request->header['connection']) != 'upgrade') {
                return;
            }

            $ws->upgrade();

            while (true) {
                $frame = $ws->recv();

                if (empty($frame) || empty($frame->data)) {
                    break;
                } else {
                    if (substr($frame->data, 0, 5) == 'close') { //关闭连接，比如 close_123593 ，关闭 123593 玩家的连接
                        $uid_array = explode("_", $frame->data);
                        $uid = $uid_array[1];
                        
                        if (!empty($uid)) {
                            Connector::close($uid);
                        }
                        $ws->close();
                        return;
                    }

                    if (substr($frame->data, 0, 9) == 'heartbeat') { //客户端心跳，比如 heartbeat_123593 ，玩家 123593 发送的心跳
                        $uid_array = explode("_", $frame->data);
                        $uid = $uid_array[1];
                        
                        if (empty($uid)) {
                            $ws->push("hearbeat no uid");
                            return;
                        }

                        //2分钟未收到心跳，则主动关闭连接
                        Connector::set_connect_expire($uid);
                        $ws->push(time());
                        continue;
                    }

                    $input = json_decode($frame->data, true);
                    if (empty($input) || empty($input['c']) || empty($input['m'])) { //输入格式错误
                        $ws->push(json_encode(array("code" => "1", "msg" => "input error", "data" => $frame->data)));
                    } else {
                        $uid = intval($input['userid']);
                        if($uid > 0) {
							Connector::set_fd($uid, $ws);
                        }

                        $application = new Application();

                        $clientInfo = array();
                        $clientInfo['remote_ip'] = empty($request->header['x-real-ip']) ? $request->header['x-forwarded-for'] : $request->header['x-forwarded-for'];
                        $result = $application->run($input, $clientInfo);
                        unset($application);
                        
                        if ($input['c'] == 'user' && $input['m'] == 'login' && $uid == 0) {
                        	$login_ret = json_decode($result['msg'], true);
                        	$uid = intval($login_ret['userid']);
                        	Connector::set_fd($uid, $ws);
                        }

                        if ($result['send_user'] === 'me') {
                            $ws->push($result['msg']);
                        } else if ($result['send_user'] === 'all') {
                            Connector::send_all($result['msg']);
                        } else if (is_array($result['send_user']) && !empty($result['send_user'])) {
                            Connector::send_fds($result['send_user'], $result['msg']);
                        }
                    }
                }
            }
        } catch (Exception $e) {
            Logger::error("Catch An Exception File=[" . $e->getFile() . "|" . $e->getLine() . "] Code=[" . $e->getCode() . "], Message=[" . $e->getMessage() . "]");
        }
    });

    //连接超时
    Swoole\Timer::tick(60000, function () {
        Connector::connect_expire();
    });

    //更新区信息
    Swoole\Timer::tick(5000, function () {
        Zoneinfo::update();
    });
    
    //更新版本信息
    Swoole\Timer::tick(10000, function () {
        Zoneinfo::game_version();
    });

    //公会战结束
    Swoole\Timer::tick(30000, function () {
        $w = date('w');
        $h = date('H');

        if (($w == 1 || $w == 3 || $w == 5) && intval($h) == 21 && RedisProxy::get_guild_fight_status() != 0) {
        	RedisProxy::set_guild_fight_status(0);
        }
    });

    echo "\n初始化成功\n\n";
    $ws->start();
});

//守护进程
Swoole\Process::daemon();

$sch->start();
