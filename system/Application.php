<?php
//日志目录
define('LOG_PATH', '/data/app/logs/super_server');
define('PHP_LOG_THRESHOLD', 1); //错误记录级别 ERROR=1, DEBUG=2, WARNING=3, NOTICE=4, INFO=5, ALL=6

echo "加载配置 ...\n";
include(APP_ROOT . "/config/constants.php");
include(APP_ROOT . "/config/database.php");
include(APP_ROOT . "/config/redis.php");

echo "\n加载基础库 ...\n";

include(BASEPATH . "/Exceptions.php");
set_error_handler('_exception_handler'); //设置异常处理函数

include(BASEPATH . "/Loader.php");
include(BASEPATH . "/Logger.php");
include(BASEPATH . "/Entity.php");
include(BASEPATH . "/RedisProxy.php");
include(BASEPATH . "/SuperController.php");
include(BASEPATH . "/SuperModel.php");
include(BASEPATH . '/DatabaseProxy.php');
include(APP_ROOT . "/core/CoreModel.php");

echo "\n创建跨进程全局实体类...\n";
$global = require(BASEPATH . "/Global.php");
$instance = GlobalEntity::instance($global);
if(empty($instance)) {
	die('Create global entity failed');
}

echo "\n加载 Controller ...\n";
include_file(APP_ROOT . "/controllers");

echo "\n加载 Model ...\n";
include_file(APP_ROOT . "/models");

echo "\n加载 Library ...\n";
include_file(APP_ROOT . "/library");

echo "\n加载 entity ...\n";
include_file(APP_ROOT . "/entity");

echo "\n加载 helper ...\n";
include_file(APP_ROOT . "/helpers");

class Application {
    var $input_fd;
    
    public function __construct($fd) {
        $this->input_fd = $fd;
    }

    public function run(& $params, $clientInfo) {
        $ret = $this->_auth($params);
        if($ret != 0) {
            return $ret;
        }

        $controller = ucfirst($params['c']);
        $action = $params['m'] . "Action";
        
        $class_name = $controller . "Controller";
            
		if(!class_exists($class_name)) {
			show_404("$controller/$action");
			return $this->response_error(2, "route error");
		}
		
		$obj = new $class_name($this->input_fd, $params, $clientInfo);
		
		if (!method_exists($obj, $action)) {
			unset($obj);
			show_404("$controller/$action");
			return $this->response_error(2, "route error");
		}
		
        try {
            $ret = $obj->$action();
            unset($obj);
            return $ret;
        } catch (Exception $e) {
        	unset($obj);
			show_404("$controller/$action");
            return $this->response_error(2, "route error");
        }
    }

    //验签过程
    protected function _auth(& $params) {
        if ($params['no'] == "test") { //测试
            return 0;
        }

        if (empty($params['auth_rand'])) {
            return $this->response_error(99990002, 'params error');
        }

        if (empty($params['timestamp'])) {
            return $this->response_error(99990003, 'params error');
        }

        if (empty($params['signature'])) {
            return $this->response_error(99990005, 'params error');
        }

        $auth_params = $params;
        $c = $params['c'];
        $m = $params['m'];
        unset($auth_params['c']);
        unset($auth_params['m']);
        unset($auth_params['signature']);

        $str = "/" . $c . "/" . $m . "/" . $auth_params['token'] . "/"; // 加密串str = "/游戏名/接口/token/"

        unset($auth_params['token']);  // 去掉 token
        ksort($auth_params);  //数组按 key 排序
        reset($auth_params);  //重置数组指针指向第一个元素

        foreach ($auth_params as $param_value) {  //将有序串加入到加密串 str
            $str = $str . trim($param_value);
        }
        $signature = md5($str); //加密得到 signature
        if ($signature != $params['signature']) { //加密之后与上送的signature 比较，如果不一致则验证失败
            return $this->response_error(99990006, "params error");
        }
        
        return 0;
    }

    public function response_error($code, $message) {
        $data = array("tagcode" => "" . $code, "description" => $message);
        $result['send_user'] = array($this->input_fd);
        $result['msg'] = json_encode($data);
        return $result;
    }
}

function include_file($path) {
	$handle = opendir($path);
	if($handle){
	    while(($filename = readdir($handle)) !== false){
	    	$pathfile = $path . "/" . $filename;
	        if(!is_dir($pathfile) && substr_compare($filename, ".php", -strlen(".php")) === 0) {
	    		echo $pathfile . "\n";
	            include($pathfile);
	        } else if (is_dir($pathfile) && $filename != '.' && $filename != '..') {
	        	include_file($pathfile);
	        }
	    }
	    closedir($handle);
	}
}