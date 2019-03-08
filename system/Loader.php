<?php
/**
 * Loader Class
 *
 * @package        SuperCI
 * @subpackage    Libraries
 * @category      Loader
 * @author        caohao
 */
class Loader
{
    private static $configs = array();
    private static $redis = array();
    
    private $instance;
    
    private $models;
    private $controllers;
    private $libraries;
    private $loggers;
    private $dbs;
    private $redises;
	
	public function __construct(& $obj) {
		$this->controllers = & $obj;
	}
    
    public static function & config($conf_name) {
        if(isset(self::$configs[$conf_name])) {
            return self::$configs[$conf_name];
        }
        
        self::$configs[$conf_name] = include(APP_ROOT . "/config/".$conf_name.".php");
        return self::$configs[$conf_name];
    }
    
    public static function entity($entity_name, $params = null) {
		if(empty($params)) {
			return new $entity_name();
		} else {
			return new $entity_name($params);
		}
    }
    
    public function library($library_name) {
    	if(!isset($this->libraries[$library_name])) {
    		$this->libraries[$library_name] = new $library_name($this);
    	}
    	return $this->libraries[$library_name];
    }
    
    public function & model($model_name) {
    	if(!isset($this->models[$model_name])) {
    		$this->models[$model_name] = new $model_name($this);
    	}
    	
		return $this->models[$model_name];
    }
    
    public function & logger($log_name, $config = array()) {
    	$key = md5($log_name . json_encode($config));
    	
    	if(empty($this->loggers[$key])) {
    		if (empty($config)) {
                $config = array('file_name' => $log_name);
            }
            
            $this->loggers[$key] = new Logger($config);
            $this->loggers[$key]->setClientIp($this->controllers->get_ip());
            $this->loggers[$key]->setParams($this->controllers->get_params());
    	}
    	
    	return $this->loggers[$key];
    }
    
    public function & database($params = '')
    {
    	if(empty($params)) {
			$params = 'default';
		}
		
		if(isset($this->dbs[$params])) {
			return $this->dbs[$params];
		}
		
    	global $db_config;
    	
    	$db_log = $this->logger("database");
    	
	    if ( ! isset($db_config) OR count($db_config) == 0) {
	    	$db_log->LogError('No database connection settings were found in the database config file.');
	    	die('No database connection settings were found in the database config file.');
	    }
		
	    if (! isset($db_config[$params])) {
	    	$db_log->LogError('You have specified an invalid database connection group.');
	    	die('You have specified an invalid database connection group.');
	    }
	    
	    $config = $db_config[$params];
	    
	    $this->dbs[$params] = new DatabaseProxy($config, $db_log);
	    
	    if ($this->dbs[$params]->autoinit == TRUE) {
			$this->dbs[$params]->initialize();
		}
		
	    return $this->dbs[$params];
    }

    public function & redis($redis_name) {
        if(empty($this->redises[$redis_name])) {
            $util_log = $this->logger('redis');
            
    		global $redis_conf;
    		
            if(!isset($redis_conf[$redis_name]['host'])) {
                $util_log->LogError("Loader::redis:  redis config not exist");
                return;
            }
            
            $redis_conf = $redis_conf[$redis_name];
            
            $this->redises[$redis_name] = new RedisPorxy($util_log, $redis_conf['host'], $redis_conf['port']);

            if(substr($redis_conf['host'], 0, 1) == '/') {
                $flag = $this->redises[$redis_name]->connect($redis_conf['host']);
            } else {
                $flag = $this->redises[$redis_name]->connect($redis_conf['host'], $redis_conf['port']);
            }

            if(!$flag) {
                $util_log->LogError("Loader::redis: redis connect error");
                $this->redises[$redis_name] = null;
                return $this->redises[$redis_name];
            }

            if(!empty($redis_conf['auth'])){
                $suc = $this->redises[$redis_name]->auth($redis_conf['auth']);
                if(!$suc) {
                    $util_log->LogError("Loader::redis:  redis auth error");
                	$this->redises[$redis_name] = null;
                	return $this->redises[$redis_name];
                }
            }
        }
        
        return $this->redises[$redis_name];
    }
}
