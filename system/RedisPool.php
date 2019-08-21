<?php
/** RedisPool.php
 * SuperModel Class
 *
 * @package			Ycsocket  https://github.com/caohao-php/ycsocket
 * @subpackage		RedisPool
 * @category		RedisPool
 * @author			caohao
 */
class RedisPool {
    const POOL_SIZE = 10;

    protected $host;
    protected $port;
    protected $pool;
    protected $logger;
    static private $instances;

    static public function instance($redis_name) {
        if (!isset(self::$instances[$redis_name])) {
            global $util_redis_conf;
            if (!isset($util_redis_conf[$redis_name]['host'])) {
            	$logger = new Logger(array('file_name' => 'redis_log'));
            	$logger->LogError("Loader::redis:  redis config not exist [$redis_name]");
                throw new RuntimeException("Loader::redis:  redis config not exist");
            }

            $pool_size = isset($util_redis_conf[$redis_name]['pool_size']) ? intval($util_redis_conf[$redis_name]['pool_size']) : RedisPool::POOL_SIZE;
            $pool_size = $pool_size <= 0 ? RedisPool::POOL_SIZE : $pool_size;

            self::$instances[$redis_name] = new RedisPool($util_redis_conf[$redis_name]['host'], $util_redis_conf[$redis_name]['port'], $pool_size);
        }

        return self::$instances[$redis_name];
    }

    /**
     * RedisPool constructor.
     * @param int $size 连接池的尺寸
     */
    function __construct($host, $port, $size) {
        $this->logger = new Logger(array('file_name' => 'redis_log'));

        $this->host = $host;
        $this->port = $port;
        $this->pool = new Swoole\Coroutine\Channel($size);

        for ($i = 0; $i < $size; $i++) {
            $redis = new Swoole\Coroutine\Redis();
            $res = $redis->connect($host, $port);
            if ($res) {
                $this->pool->push($redis);
            } else {
                throw new RuntimeException("Redis connect error [$host] [$port]");
            }
        }
    }

    public function __call($func, $args) {
        $ret = null;
        try {
            $redis = $this->pool->pop();
            $ret = call_user_func_array(array($redis, $func), $args);
            if ($ret === false && $redis->errCode != 0) {
                $this->logger->LogError("redis reconnect [{$this->host}][{$this->port}]");

                //重连一次
                $redis->close();
                $res = $redis->connect($this->host, $this->port);
                if (!$res) {
                    throw new RuntimeException("Redis reconnect error [{$this->host}][{$this->port}]");
                }

                $ret = call_user_func_array(array($redis, $func), $args);

                if ($ret === false && $redis->errCode != 0) {
                    throw new RuntimeException("redis error after reconnect");
                }
            }

            $this->pool->push($redis);
        } catch (Exception $e) {
            $this->pool->push($redis);
            $this->logger->LogError("Redis catch exception [".$e->getMessage()."] [$func]");
            throw new RuntimeException("Redis catch exception [".$e->getMessage()."] [$func]");
        }

        return $ret;
    }
}