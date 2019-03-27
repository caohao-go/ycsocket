<?php

class RedisPool {
    const REDIS_POOL_SIZE = 10;

    protected $host;
    protected $port;

    protected $pool;
    protected $logger;

    static private $instances;

    static public function instance($redis_name) {
        if (!isset(self::$instances[$redis_name])) {
            global $util_redis_conf;

            if (!isset($util_redis_conf[$redis_name]['host'])) {
                $this->logger->LogError("Loader::redis:  redis config not exist");
                echo "Loader::redis:  redis config not exist";
                exit;
            }

            self::$instances[$redis_name] = new RedisPool($util_redis_conf[$redis_name]['host'], $util_redis_conf[$redis_name]['port']);
        }

        return self::$instances[$redis_name];
    }

    /**
     * RedisPool constructor.
     * @param int $size 连接池的尺寸
     */
    function __construct($host, $port, $size = RedisPool::REDIS_POOL_SIZE) {
        $this->logger = new Logger(array('file_name' => 'redis_log'));

        $this->host = $host;
        $this->port = $port;

        $this->pool = new Swoole\Coroutine\Channel($size);
        for ($i = 0; $i < $size; $i++) {
            $redis = new Swoole\Coroutine\Redis();
            $res = $redis->connect($host, $port);
            if ($res == false) {
                $this->logger->LogError("failed to connect redis server.[$host][$port]");
            } else {
                $this->pool->push($redis);
            }
        }
    }

    public function __call($func, $args) {
        $ret = null;
        try {
            $redis = $this->pool->pop();

            $ret = call_user_func_array(array($redis, $func), $args);
            if ($ret === false) {
                $this->logger->LogError("redis error [$func], reconnect [{$this->host}][{$this->port}]");

                $redis->close();
                $redis->connect($this->host, $this->port);
                $ret = call_user_func_array(array($redis, $func), $args);
            }

            $this->pool->push($redis);
        } catch (Exception $e) {
            $this->logger->LogError("redis error [$func], [".$e->getMessage()."], reconnect [{$this->host}][{$this->port}]");
            $redis->close();
            $redis->connect($this->host, $this->port);
            $ret = call_user_func_array(array($redis, $func), $args);
            $this->pool->push($redis);
        }

        return $ret;
    }
}