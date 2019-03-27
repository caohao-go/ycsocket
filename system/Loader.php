<?php
/**
 * Loader Class
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
    private $dbs;
    private $redises;

    public function __construct(& $controller) {
        $this->ip = $controller->get_ip();
        $this->params = $controller->get_params();
    }

    public static function & config($conf_name) {
        if (isset(self::$configs[$conf_name])) {
            return self::$configs[$conf_name];
        }

        self::$configs[$conf_name] = include(APP_ROOT . "/config/".$conf_name.".php");
        return self::$configs[$conf_name];
    }

    public function entity($entity_name) {
        if (!isset($this->entities[$entity_name])) {
            $this->entities[$entity_name] = new $entity_name($this);
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

    public function & database($params = '') {
        if (empty($params)) {
            $params = 'default';
        }

        if (isset($this->dbs[$params])) {
            return $this->dbs[$params];
        }

        global $util_db_config;

        $db_log = $this->logger("database");

        if ( ! isset($util_db_config) OR count($util_db_config) == 0) {
            $db_log->LogError('No database connection settings were found in the database config file.');
            die('No database connection settings were found in the database config file.');
        }

        if (! isset($util_db_config[$params])) {
            $db_log->LogError('You have specified an invalid database connection group.');
            die('You have specified an invalid database connection group.');
        }

        $config = $util_db_config[$params];

        $this->dbs[$params] = new DatabaseProxy($config, $db_log);

        if ($this->dbs[$params]->autoinit == TRUE) {
            $this->dbs[$params]->initialize();
        }

        return $this->dbs[$params];
    }

    public function & redis($redis_name) {
        if (empty($this->redises[$redis_name])) {
            $util_log = $this->logger('redis');

            global $util_redis_conf;

            if (!isset($util_redis_conf[$redis_name]['host'])) {
                $util_log->LogError("Loader::redis:  redis config not exist");
                return;
            }

            $this->redises[$redis_name] = new Redis();

            if (substr($util_redis_conf[$redis_name]['host'], 0, 1) == '/') {
                $flag = $this->redises[$redis_name]->connect($util_redis_conf[$redis_name]['host']);
            } else {
                $flag = $this->redises[$redis_name]->connect($util_redis_conf[$redis_name]['host'], $util_redis_conf[$redis_name]['port']);
            }

            if (!$flag) {
                $util_log->LogError("Loader::redis: redis connect error");
                return null;
            }

            if (!empty($util_redis_conf[$redis_name]['auth'])) {
                $suc = $this->redises[$redis_name]->auth($util_redis_conf[$redis_name]['auth']);
                if (!$suc) {
                    $util_log->LogError("Loader::redis:  redis auth error");
                    return null;
                }
            }
        }

        return $this->redises[$redis_name];
    }
}
