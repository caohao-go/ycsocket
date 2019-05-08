<?php
//日志目录
define('LOG_PATH', '/data/app/logs/super_server');
define('PHP_LOG_THRESHOLD', 1); //错误记录级别 ERROR=1, DEBUG=2, WARNING=3, NOTICE=4, INFO=5, ALL=6

echo "加载配置 ...\n";
include(APPROOT . "/config/constants.php");
include(APPROOT . "/config/database.php");
include(APPROOT . "/config/redis.php");

echo "\n加载基础库 ...\n";

include(BASEPATH . "/Exceptions.php");
set_error_handler('_exception_handler'); //设置异常处理函数

include(BASEPATH . "/Singleton.php");
include(BASEPATH . "/Loader.php");
include(BASEPATH . "/Logger.php");
include(BASEPATH . "/RedisPool.php");
include(BASEPATH . "/MySQLPool.php");
include(BASEPATH . "/SuperController.php");
include(BASEPATH . "/SuperModel.php");
include(APPROOT . "/core/CoreModel.php");

echo "\n创建跨进程全局实体类...\n";
require(BASEPATH . "/Global.php");
$globalTable = new swoole_table(40960);
$globalTable->column('int_data', swoole_table::TYPE_INT);
$globalTable->column('data', swoole_table::TYPE_STRING, 4096);
$globalTable->column('expire', swoole_table::TYPE_INT);
$globalTable->create();

$roomActorId = new swoole_table(1);
$roomActorId->column('id', swoole_table::TYPE_STRING, 32);
$roomActorId->create();

$fdUserTable = new swoole_table(102400);
$fdUserTable->column('fd', swoole_table::TYPE_INT);
$fdUserTable->create();

echo "\n加载 Actor ...\n";
include_file(BASEPATH . "/actor");

echo "\n加载 Controller ...\n";
include_file(APPROOT . "/controllers");

echo "\n加载 Model ...\n";
include_file(APPROOT . "/models");

echo "\n加载 Logic ...\n";
include_file(APPROOT . "/logic");

echo "\n加载 Library ...\n";
include_file(APPROOT . "/library");

echo "\n加载 entity ...\n";
include_file(APPROOT . "/entity");

echo "\n加载 helper ...\n";
include_file(APPROOT . "/helpers");

echo "\n注册 Actor ...\n";
register_actor();

class Application {
    var $input_fd;

    public function __construct($fd) {
        $this->input_fd = $fd;
    }

    public function run(& $params, $clientInfo, & $ws) {
        $ret = $this->_auth($params);
        if ($ret != 0) {
            return $ret;
        }

        $controller = ucfirst($params['c']);
        $action = $params['m'] . "Action";
        $class_name = $controller . "Controller";

        try {
            $obj = new $class_name($this->input_fd, $params, $clientInfo, $ws);

            if (!method_exists($obj, $action)) {
                unset($obj);
                show_404("$controller/$action");
                return $this->response_error(3, "route error");
            }

            $ret = $obj->$action();
            unset($obj);
            return $ret;
        } catch (Exception $e) {
            unset($obj);

            if ($e->getMessage() != 'swoole exit.') {
                $logger = new Logger(array('file_name' => 'exception_log'));
                $logger->LogError("Catch An Exception File=[".$e->getFile()."|".$e->getLine()."] Code=[".$e->getCode()."], Message=[".$e->getMessage()."]");

                echo "Catch An Exception \n";
                echo "File:" . $e->getFile() . "\n";
                echo "Line:" . $e->getLine() . "\n";
                echo "Code:" . $e->getCode() . "\n";
                echo "Message:" . $e->getMessage() . "\n";
                return $this->response_error(99, "system exception");
            } else {
                echo "swoole exit.\n";
                return $this->response_error(99, "application exit");
            }
        }
    }

    //验签过程
    protected function _auth(& $params) {
        return 0;
    }

    public function response_error($code, $message) {
        $data = array("code" => $code, "msg" => $message);
        $result['send_user'] = array($this->input_fd);
        $result['msg'] = json_encode($data);
        return $result;
    }
}

function include_file($path) {
    $handle = opendir($path);
    if ($handle) {
        while (($filename = readdir($handle)) !== false) {
            $pathfile = $path . "/" . $filename;
            if (!is_dir($pathfile) && substr_compare($filename, ".php", -strlen(".php")) === 0) {
                echo $pathfile . "\n";
                include($pathfile);
            } else if (is_dir($pathfile) && $filename != '.' && $filename != '..') {
                include_file($pathfile);
            }
        }
        closedir($handle);
    }
}

function register_actor() {
	Actor::getInstance()->register(RoomLogic::class, 1);
	Actor::getInstance()->register(PkLogic::class, 1);
	Actor::getInstance()->register(GameLogic::class, 1);
}
