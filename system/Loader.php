<?php

/**
 * Loader Class
 *
 * @package            Ycsocket
 * @subpackage        Loader
 * @category        Loader
 * @author            caohao
 */
class Loader
{
    private static $configs = array();

    private $ip = '';
    private $params = array();

    private $daos;
    private $services;
    private $libraries;
    private $loggers;

    public function __construct(& $controller = null)
    {
        if (!empty($controller)) {
            $this->ip = $controller->get_ip();
            $this->params = $controller->get_params();
        }
    }

    public static function & config($conf_name)
    {
        if (isset(self::$configs[$conf_name])) {
            return self::$configs[$conf_name];
        }

        self::$configs[$conf_name] = include(APP_ROOT . "/config/" . $conf_name . ".php");
        return self::$configs[$conf_name];
    }

    public function & dao($dao_name)
    {
        if (!isset($this->daos[$dao_name])) {
            $this->daos[$dao_name] = new $dao_name($this);
        }

        return $this->daos[$dao_name];
    }

    public function & library($library_name, $params = null)
    {
        if (!isset($this->libraries[$library_name])) {
            if($params === null) {
                $this->libraries[$library_name] = new $library_name();
            } else {
                $this->libraries[$library_name] = new $library_name($params);
            }
        }

        return $this->libraries[$library_name];
    }

    public function & service($service_name)
    {
        if (!isset($this->services[$service_name])) {
            $this->services[$service_name] = new $service_name($this);
        }

        return $this->services[$service_name];
    }

    public function & logger($log_name)
    {
        if (empty($this->loggers[$log_name])) {
            $this->loggers[$log_name] = new Logger($log_name);
            $this->loggers[$log_name]->setClientIp($this->ip);
            $this->loggers[$log_name]->setParams($this->params);
        }

        return $this->loggers[$log_name];
    }
}
