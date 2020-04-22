<?php
echo "加载配置 ...\n";
include(APP_ROOT . "/config/constants.php");
include(APP_ROOT . "/config/database.php");
include(APP_ROOT . "/config/redis.php");

echo "\n加载基础库 ...\n";

include(BASEPATH . "/Exceptions.php");
set_error_handler('_exception_handler'); //设置异常处理函数

include(BASEPATH . "/Loader.php");
include(BASEPATH . "/Logger.php");
include(BASEPATH . "/RedisPool.php");
include(BASEPATH . "/MySQLPool.php");
include(BASEPATH . "/SuperController.php");
include(BASEPATH . "/SuperService.php");
include(BASEPATH . "/SuperDao.php");
include(BASEPATH . "/Connector.php");
include(APP_ROOT . "/Filter.php");

echo "\n加载 Library ...\n";
include_file(APP_ROOT . "/library");

echo "\n加载 Dao ...\n";
include_file(APP_ROOT . "/dao");

echo "\n加载 Service ...\n";
include_file(APP_ROOT . "/service");

echo "\n加载 Controller ...\n";
include_file(APP_ROOT . "/controller");

class Application
{
    public function __construct()
    {
    }

    public function run(& $params, $clientInfo)
    {
        $ret = Filter::auth($params);
        if ($ret != 0) {
            return $ret;
        }

        foreach ($params as $k => $v) {
            $params[$k] = trim($v);
        }

        $controller = ucfirst($params['c']);
        $action = $params['m'] . "Action";
        $class_name = $controller . "Controller";

        try {
            $obj = new $class_name($params, $clientInfo);

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

            if ($e instanceof LogicException) { //业务异常
                $errorcode = $e->getCode() == 0 ? 8 : $e->getCode();
                return $this->response_error($errorcode, $e->getMessage());
            } else if ($e->getMessage() != 'swoole exit.') {
                Logger::error("Catch An Exception File=[" . $e->getFile() . "|" . $e->getLine() . "] Code=[" . $e->getCode() . "], Message=[" . $e->getMessage() . "]", "exception_log");

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

    public function response_error($code, $message)
    {
        $data = array("code" => $code, "msg" => $message);
        $result['send_user'] = "me";
        $result['msg'] = json_encode($data);
        return $result;
    }
}

function include_file($path)
{
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
