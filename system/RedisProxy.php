<?php
/**
 * RedisPorxy class
 *
 * @package       SuperCI
 * @subpackage    Libraries
 * @category      RedisPorxy
 * @author        caohao
 */
class RedisPorxy {
	private $ip;
	private $port;
	private $conn;
	
	public function __construct(& $logger, $ip, $port = 6379) {
		$this->redis_log = & $logger;
		$this->ip = $ip;
		$this->port = $port;
		$this->conn = new Redis();
	}
	
	/**
	 * 魔术方法, 透传到redis连接
	 * @param $func
	 * @param $args
	 * @return mixed
	 * @throws RedisException
	 */
	public function __call($func, $args)
	{
		$ret = null;
		try {
			$ret = call_user_func_array(array($this->conn, $func), $args);
		} catch (Exception $e) {
			$this->redis_log->LogError("redis error [$func]: ip=[" . $this->ip . "] port=[" . $this->port . "] error=[".$e->getMessage()."]");
		}
		return $ret;
	}
}