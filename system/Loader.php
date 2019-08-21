<?php
/**
 * Loader Class  https://github.com/caohao-php/ycsocket
 *
 * @package			Ycsocket
 * @subpackage		Libraries
 * @category		Loader
 * @author			caohao
 */
class Loader {
    private static $configs = array();

    private $ip;
    private $params;

    private $entities;
    private $models;
    private $libraries;
    private $loggers;

    public function __construct(& $controller) {
        $this->ip = $controller->get_ip();
        $this->params = $controller->get_params();
    }

    public static function & config($conf_name) {
        if (isset(self::$configs[$conf_name])) {
            return self::$configs[$conf_name];
        }

        self::$configs[$conf_name] = include(APPROOT . "/config/".$conf_name.".php");
        return self::$configs[$conf_name];
    }

    public function entity($entity_name) {
        if (!isset($this->entities[$entity_name])) {
            $this->entities[$entity_name] = new $entity_name();
        }

        return $this->entities[$entity_name];
    }

    public function & library($library_name) {
        if (!isset($this->libraries[$library_name])) {
            $this->libraries[$library_name] = new $library_name($this);
        }

        return $this->libraries[$library_name];
    }

    public function & model($model_name) {
        if (!isset($this->models[$model_name])) {
            $this->models[$model_name] = new $model_name($this);
        }

        return $this->models[$model_name];
    }

    public function & logger($log_name, $config = array()) {
        $key = md5($log_name . json_encode($config));

        if (empty($this->loggers[$key])) {
            if (empty($config)) {
                $config = array('file_name' => $log_name);
            }

            $this->loggers[$key] = new Logger($config);
            $this->loggers[$key]->setClientIp($this->ip);
            $this->loggers[$key]->setParams($this->params);
        }

        return $this->loggers[$key];
    }
}
