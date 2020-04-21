<?php if (!defined('BASEPATH')) exit('No direct script access allowed');

/**
 * SuperDao Class
 *
 * @package        Ycsocket
 * @subpackage    Dao
 * @category      SuperDao
 * @author        caohao
 */
class SuperDao
{
    var $db_name;
    var $redis_name;
    const EMPTY_STRING = -999999999;  //防止数据库击穿模拟空串

    protected $loader;

    public function __construct(& $loader = null)
    {
        if (!empty($loader)) {
            $this->loader = &$loader;
        } else {
            $this->loader = new Loader();
        }

        $this->redis_name = "default";
        $this->db_name = "default";

        $this->init();
    }

    protected function init()
    {
    }

    /**
     * 根据key获取表记录
     * @param string redis_key redis 缓存键值
     */
    public function hget_redis($redis_key, $field)
    {
        if (empty($redis_key)) return;
        return RedisPool::instance($this->redis_name)->hget($redis_key, $field);
    }

    /**
     * 设置 redis 值
     * @param string redis_key redis 缓存键值, 可空， 非空时清理键值缓存
     * @param array data 表数据
     * @param int redis_expire redis 缓存到期时长(秒)
     * @param boolean set_empty_flag 是否缓存空值，如果缓存空值，在表记录更新之后，一定记得清理空值标记缓存
     */
    public function hset_redis($redis_key, $field, $data, $redis_expire = 600, $set_empty_flag = true)
    {
        if (empty($redis_key)) return;

        if (empty($data) && $set_empty_flag) {
            $ret = RedisPool::instance($this->redis_name)->hset($redis_key, $field, self::EMPTY_STRING);
        } else {
            $ret = RedisPool::instance($this->redis_name)->hset($redis_key, $field, serialize($data));
        }

        RedisPool::instance($this->redis_name)->expire($redis_key, $redis_expire);
        return $ret;
    }

    /**
     * 根据key获取表记录
     * @param string redis_key redis 缓存键值
     */
    public function get_redis($redis_key)
    {
        if (empty($redis_key)) return;
        return RedisPool::instance($this->redis_name)->get($redis_key);
    }

    /**
     * 设置 redis 值
     * @param string redis_key redis 缓存键值, 可空， 非空时清理键值缓存
     * @param array data 表数据
     * @param int redis_expire redis 缓存到期时长(秒)
     * @param boolean set_empty_flag 是否缓存空值，如果缓存空值，在表记录更新之后，一定记得清理空值标记缓存
     */
    public function set_redis($redis_key, $data, $redis_expire = 600, $set_empty_flag = true)
    {
        if (empty($redis_key)) return;

        if (empty($data) && $set_empty_flag) {
            RedisPool::instance($this->redis_name)->set($redis_key, self::EMPTY_STRING);
        } else {
            RedisPool::instance($this->redis_name)->set($redis_key, serialize($data));
        }

        RedisPool::instance($this->redis_name)->expire($redis_key, $redis_expire);
    }

    /**
     * 清理记录缓存
     * @param string redis_key redis 缓存键值
     */
    public function clear_redis_cache($redis_key = "")
    {
        if (empty($redis_key)) {
            return;
        }

        RedisPool::instance($this->redis_name)->del($redis_key);
    }

    /**
     * 插入表记录
     * @param string table 表名
     * @param array data 表数据
     * @param string redis_key redis 缓存键值, 可空， 非空时清理键值缓存
     */
    public function insert_table($table, $data, $redis_key = "")
    {
        $ret = MySQLPool::instance($this->db_name)->insert($table, $data);

        if (!empty($redis_key)) {
            $this->clear_redis_cache($redis_key);
        }

        if ($ret < 0) {
            Logger::error("error to insert_table $table , DATA=[" . json_encode($data) . "]");
            return 0;
        }

        return $ret;
    }

    /**
     * 更新表记录
     * @param string table 表名
     * @param array where 查询条件
     * @param array data 更新数据
     * @param string redis_key redis 缓存键值, 可空， 非空时清理键值缓存
     */
    public function update_table($table, $where, $data, $redis_key = "")
    {
        if (empty($where)) return;
        $ret = MySQLPool::instance($this->db_name)->update($table, $where, $data);

        if (!empty($redis_key)) {
            $this->clear_redis_cache($redis_key);
        }

        if ($ret) {
            return true;
        } else {
            Logger::error("error to update_table $table [" . json_encode($where) . "], DATA=[" . json_encode($data) . "]");
            return false;
        }
    }

    /**
     * 替换表记录
     * @param string table 表名
     * @param array data 替换数据
     * @param string redis_key redis 缓存键值, 可空， 非空时清理键值缓存
     */
    public function replace_table($table, $data, $redis_key = "")
    {
        $ret = MySQLPool::instance($this->db_name)->replace($table, $data);

        if (!empty($redis_key)) {
            $this->clear_redis_cache($redis_key);
        }

        if ($ret) {
            return true;
        } else {
            Logger::error("error to replace_table $table , DATA=[" . json_encode($data) . "]");
            return false;
        }
    }

    /**
     * 删除表记录
     * @param string table 表名
     * @param array where 查询条件
     * @param string redis_key redis缓存键值, 可空， 非空时清理键值缓存
     */
    public function delete_table($table, $where, $redis_key = "")
    {
        if (empty($where)) return;
        $ret = MySQLPool::instance($this->db_name)->delete($table, $where);

        if (!empty($redis_key)) {
            $this->clear_redis_cache($redis_key);
        }

        if ($ret) {
            return true;
        } else {
            Logger::error("error to delete_table $table [" . json_encode($where) . "]");
            return false;
        }
    }

    /**
     * 获取表数据
     * @param string table 表名
     * @param array where 查询条件
     * @param string redis_key redis 缓存键值, 可空， 非空时清理键值缓存
     * @param int redis_expire redis 缓存到期时长(秒)
     * @param boolean set_empty_flag 是否标注空值，如果标注空值，在表记录更新之后，一定记得清理空值标记缓存
     */
    public function get_table_data($table, $where = null, $redis_key = "", $column = "*", $redis_expire = 600, $set_empty_flag = true)
    {
        $data = $this->get_redis($redis_key);
        if (!empty($data)) {
            return $data == self::EMPTY_STRING ? array() : unserialize($data);
        }

        $data = MySQLPool::instance($this->db_name)->get($table, $where, $column);

        $this->set_redis($redis_key, $data, $redis_expire, $set_empty_flag);
        return $data;
    }

    /**
     * 获取一条表数据
     * @param string table 表名
     * @param array where 查询条件
     * @param string redis_key redis 缓存键值, 可空， 非空时清理键值缓存
     * @param int redis_expire redis 缓存到期时长(秒)
     * @param boolean set_empty_flag 是否标注空值，如果标注空值，在表记录更新之后，一定记得清理空值标记缓存
     */
    public function get_one_table_data($table, $where = null, $redis_key = "", $redis_expire = 600, $set_empty_flag = true)
    {
        $data = $this->get_redis($redis_key);
        if (!empty($data)) {
            return $data == self::EMPTY_STRING ? array() : unserialize($data);
        }

        $data = MySQLPool::instance($this->db_name)->get_one($table, $where);
        $this->set_redis($redis_key, $data, $redis_expire, $set_empty_flag);
        return $data;
    }
}
