<?php

class RedisPool
{
    const POOL_SIZE = 10;

    protected $host;
    protected $port;
    protected $auth;
    protected $pool;
    static private $instances;

    static public function instance($redis_name)
    {
        if (!isset(self::$instances[$redis_name])) {
            global $util_redis_conf;
            if (!isset($util_redis_conf[$redis_name]['host'])) {
                Logger::error("Loader::redis:  redis config not exist [$redis_name]", "redis_log");
                throw new RuntimeException("Loader::redis:  redis config not exist");
            }

            if (substr($util_redis_conf[$redis_name]['host'], 0, 1) == '/') {
                self::$instances[$redis_name] = new Redis();
                $flag = self::$instances[$redis_name]->connect($util_redis_conf[$redis_name]['host']);
                if (!$flag) {
                    Logger::error("Loader::redis:  redis connect unix domain error  [$redis_name]", "redis_log");
                    throw new RuntimeException("Loader::redis: redis connect unix domain error");
                }
            } else {
                $pool_size = isset($util_redis_conf[$redis_name]['pool_size']) ? intval($util_redis_conf[$redis_name]['pool_size']) : RedisPool::POOL_SIZE;
                $pool_size = $pool_size <= 0 ? RedisPool::POOL_SIZE : $pool_size;

                self::$instances[$redis_name] = new RedisPool($util_redis_conf[$redis_name]['host'], $util_redis_conf[$redis_name]['port'], $pool_size, $util_redis_conf[$redis_name]['auth']);
            }
        }

        return self::$instances[$redis_name];
    }

    /**
     * RedisPool constructor.
     * @param int $size 连接池的尺寸
     */
    function __construct($host, $port, $size, $auth = "")
    {
        $this->host = $host;
        $this->port = $port;
        $this->auth = $auth;
        $this->pool = new Swoole\Coroutine\Channel($size);

        for ($i = 0; $i < $size; $i++) {
            $redis = new Swoole\Coroutine\Redis();
            $redis->setOptions(['compatibility_mode' => true]);
            $res = $redis->connect($host, $port);
            if ($res) {
                if(!empty($auth)) {
                    $redis->auth($auth);
                }
                $this->pool->push($redis);
            } else {
                throw new RuntimeException("Redis connect error [$host] [$port]");
            }
        }
    }

    public function __call($func, $args)
    {
        $ret = null;
        try {
            $redis = $this->pool->pop();
            $ret = call_user_func_array(array($redis, $func), $args);
            if ($ret === false && $redis->errCode != 0) {
                Logger::error("redis reconnect [{$this->host}][{$this->port}]", "redis_log");

                //重连一次
                $redis->close();
                $res = $redis->connect($this->host, $this->port);
                if (!$res) {
                    throw new RuntimeException("Redis reconnect error [{$this->host}][{$this->port}]");
                }
                if(!empty($this->auth)) {
                    $redis->auth($this->auth);
                }

                $ret = call_user_func_array(array($redis, $func), $args);

                if ($ret === false && $redis->errCode != 0) {
                    throw new RuntimeException("redis error after reconnect");
                }
            }

            $this->pool->push($redis);
        } catch (Exception $e) {
            $this->pool->push($redis);
            Logger::error("Redis catch exception [" . $e->getMessage() . "] [$func]", "redis_log");
            throw new RuntimeException("Redis catch exception [" . $e->getMessage() . "] [$func]");
        }

        return $ret;
    }
}
