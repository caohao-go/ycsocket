<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');
/**
 * ExampleModel Class  https://github.com/caohao-php/ycsocket
 *
 * @package        Ycsocket
 * @subpackage    Model
 * @category      Example Model
 * @author        caohao
 */
class CoreModel extends SuperModel {
    var $db_name;
    var $redis_name;
    const EMPTY_STRING = -999999999;

    public function init() {
        $this->util_log = $this->loader->logger('model_log');
        $this->redis_name = "default";
        $this->db_name = "default";
    }

    /**
     * 根据key获取表记录
     * @param string redis_key redis 缓存键值
     */
    public function get_redis($redis_key) {
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
    public function set_redis($redis_key, $data, $redis_expire, $set_empty_flag) {
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
    public function clear_redis_cache($redis_key = "") {
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
    public function insert_table($table, $data, $redis_key = "") {
        $ret = MySQLPool::instance($this->db_name)->insert($table, $data);

        if (!empty($redis_key)) {
            $this->clear_redis_cache($redis_key);
        }

        if ($ret <= 0) {
            $this->util_log->LogError("error to insert_table $table , DATA=[".json_encode($data)."]");
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
    public function update_table($table, $where, $data, $redis_key = "") {
        if (empty($where)) return;
        $ret = MySQLPool::instance($this->db_name)->update($table, $where, $data);

        if (!empty($redis_key)) {
            $this->clear_redis_cache($redis_key);
        }

        if ($ret) {
            return true;
        } else {
            $this->util_log->LogError("error to update_table $table [".json_encode($where)."], DATA=[".json_encode($data)."]");
            return false;
        }
    }

    /**
     * 替换表记录
     * @param string table 表名
     * @param array data 替换数据
     * @param string redis_key redis 缓存键值, 可空， 非空时清理键值缓存
     */
    public function replace_table($table, $data, $redis_key = "") {
        $ret = MySQLPool::instance($this->db_name)->replace($table, $data);

        if (!empty($redis_key)) {
            $this->clear_redis_cache($redis_key);
        }

        if ($ret) {
            return true;
        } else {
            $this->util_log->LogError("error to replace_table $table , DATA=[".json_encode($data)."]");
            return false;
        }
    }

    /**
     * 删除表记录
     * @param string table 表名
     * @param array where 查询条件
     * @param string redis_key redis缓存键值, 可空， 非空时清理键值缓存
     */
    public function delete_table($table, $where, $redis_key = "") {
        if (empty($where)) return;
        $ret = MySQLPool::instance($this->db_name)->delete($table, $where);

        if (!empty($redis_key)) {
            $this->clear_redis_cache($redis_key);
        }

        if ($ret) {
            return true;
        } else {
            $this->util_log->LogError("error to delete_table $table [".json_encode($where)."]");
            return false;
        }
    }

    /**
     * 获取表数据
     * @param string table 表名
     * @param array where 查询条件
     * @param string redis_key redis 缓存键值, 可空， 非空时清理键值缓存
     * @param int redis_expire redis 缓存到期时长(秒)
     * @param string $column 数据库表字段，可空
     * @param boolean set_empty_flag 是否将空值写入缓存，防止数据库击穿，默认为是
     */
    public function get_table_data($table, $where = null, $redis_key = "", $redis_expire = 600, $column = "*", $set_empty_flag = true) {
        $data = $this->get_redis($redis_key);
        if (!empty($data)) {
            return $data == self::EMPTY_STRING ? array() : unserialize($data);
        }

        $data = MySQLPool::instance($this->db_name)->get($table, $where, $column);

        $this->set_redis($redis_key, $data, $redis_expire, $set_empty_flag);
        return $data;
    }

    /**
     * 分页获取表数据
     * @param string table 表名
     * @param array where 查询条件
     * @param array page - 页数，从 1 开始
     * @param array page_size - 每页条数，默认为 10 条
     * @param string redis_key redis 缓存键值, 可空， 非空时清理键值缓存
     * @param int redis_expire redis 缓存到期时长(秒)
     * @param array column 请求列
     * @param boolean set_empty_flag 是否将空值写入缓存，防止数据库击穿，默认为是
     */
    public function get_table_data_by_page($table, $where = null, $page = 1, $page_size = 10, $redis_key = "", $redis_expire = 600, $column = "*", $set_empty_flag = true) {
        if($page < 1 || $page_size <= 0) {
            return array();
        }

        if(!empty($redis_key)) {
            $redis_key = $redis_key . "_{$page_size}_{$page}";
        }

        $data = $this->get_redis($redis_key);
        if (!empty($data)) {
            if ($data == self::EMPTY_STRING) {
                return;
            } else {
                return unserialize($data);
            }
        }

        $where = empty($where) ? array() : $where;
        $start = ($page - 1) * $page_size;
        $where['LIMIT'] = [$start, $page_size];

        $data = MySQLPool::instance($this->db_name)->get($table, $where, $column);
        if($data != -1) {
            $this->set_redis($redis_key, $data, $redis_expire, $set_empty_flag);
            return $data;
        }
        return array();
    }

    /**
     * 获取一条表数据
     * @param string table 表名
     * @param array where 查询条件
     * @param string redis_key redis 缓存键值, 可空， 非空时清理键值缓存
     * @param int redis_expire redis 缓存到期时长(秒)
     * @param string $column 数据库表字段，可空
     * @param boolean set_empty_flag 是否将空值写入缓存，防止数据库击穿，默认为是
     */
    public function get_one_table_data($table, $where = null, $redis_key = "", $redis_expire = 600, $column = "*", $set_empty_flag = true) {
        $data = $this->get_redis($redis_key);
        if (!empty($data)) {
            return $data == self::EMPTY_STRING ? array() : unserialize($data);
        }
        
        $data = MySQLPool::instance($this->db_name)->get_one($table, $where, $column);
        $this->set_redis($redis_key, $data, $redis_expire, $set_empty_flag);
        return $data;
    }
}
