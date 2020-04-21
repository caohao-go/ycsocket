<?php

class MySQLPool
{
    const POOL_SIZE = 10;

    protected $pool;
    static private $instances;

    var $host = '';
    var $username = '';
    var $password = '';
    var $dbname = '';
    var $port = 3306;
    var $char_set = 'utf8';
    var $dbcollat = 'utf8_general_ci';

    static public function instance($params)
    {
        if (!isset(self::$instances[$params])) {

            $params = empty($params) ? 'default' : $params;

            global $util_db_config;

            if (!isset($util_db_config[$params])) {
                throw new RuntimeException("You have specified an invalid database connection group.");
            }

            $config = $util_db_config[$params];

            $pool_size = isset($config['pool_size']) ? intval($config['pool_size']) : MySQLPool::POOL_SIZE;
            $pool_size = $pool_size <= 0 ? MySQLPool::POOL_SIZE : $pool_size;

            self::$instances[$params] = new MySQLPool($config, $pool_size);
        }

        return self::$instances[$params];
    }

    /**
     * MySQLPool constructor.
     * @param int $size 连接池的尺寸
     */
    function __construct($params, $size)
    {
        foreach ($params as $key => $val) {
            $this->$key = $val;
        }

        $this->pool = new Swoole\Coroutine\Channel($size);

        for ($i = 0; $i < $size; $i++) {
            $mysql = new Swoole\Coroutine\MySQL();

            $ret = $this->connect($mysql);

            if ($ret) {
                $this->pool->push($mysql);
                $this->query("SET NAMES '" . $this->char_set . "' COLLATE '" . $this->dbcollat . "'");
            } else {
                throw new RuntimeException("MySQL connect error host={$this->host}, port={$this->port}, user={$this->username}, database={$this->dbname}, errno=[" . $mysql->errno . "], error=[" . $mysql->error . "]");
            }
        }
    }

    function insert($table = '', $data = NULL)
    {
        if (empty($table) || empty($data) || !is_array($data)) {
            throw new RuntimeException("insert_table_or_data_must_be_set");
        }

        $sql_stat = swoole_orm::insert($table, $data);
        if ($sql_stat === false) {
            throw new RuntimeException("insert_sql error [$table][" . json_encode($data) . "]");
        }

        $ret = $this->query($sql_stat['sql'], $sql_stat['bind_value'], $mysql);
        if (!empty($ret)) {
            return $mysql->insert_id;
        } else {
            return intval($ret);
        }
    }

    function replace($table = '', $data = NULL)
    {
        if (empty($table) || empty($data) || !is_array($data)) {
            throw new RuntimeException("replace_table_or_data_must_be_set");
        }

        $sql_stat = swoole_orm::replace($table, $data);
        if ($sql_stat === false) {
            throw new RuntimeException("replace_sql error [$table][" . json_encode($data) . "]");
        }

        $ret = $this->query($sql_stat['sql'], $sql_stat['bind_value']);
        return $ret;
    }

    function update($table = '', $where = NULL, $data = NULL)
    {
        if (empty($table) || empty($where) || empty($data) || !is_array($data)) {
            throw new RuntimeException("update_table_or_data_must_be_set table=[" . json_encode($table) . "], where=[" . json_encode($where) . "], data=[" . json_encode($data) . "]");
        }

        $sql_stat = swoole_orm::update($table, $data, $where);
        if ($sql_stat === false) {
            throw new RuntimeException("update_sql error [$table][" . json_encode($data) . "][" . json_encode($where) . "]");
        }

        $ret = $this->query($sql_stat['sql'], $sql_stat['bind_value']);
        return $ret;
    }

    function delete($table = '', $where = NULL)
    {
        if (empty($table) || empty($where)) {
            throw new RuntimeException("delete_table_or_where_must_be_set");
        }

        $sql_stat = swoole_orm::delete($table, $where);
        if ($sql_stat === false) {
            throw new RuntimeException("replace_sql error [$table][" . json_encode($where) . "]");
        }

        $ret = $this->query($sql_stat['sql'], $sql_stat['bind_value']);
        return $ret;
    }

    function get($table = '', $where = array(), $columns = "*")
    {
        if (empty($table) || empty($columns)) {
            throw new RuntimeException("select_table_or_columns_must_be_set");
        }

        $sql_stat = swoole_orm::select($table, $columns, $where);
        if ($sql_stat === false) {
            throw new RuntimeException("select_sql error [$table][" . json_encode($where) . "][" . json_encode($columns) . "]");
        }

        $ret = $this->query($sql_stat['sql'], $sql_stat['bind_value']);
        return $ret;
    }

    function get_one($table = '', $where = array(), $columns = "*")
    {
        $where['LIMIT'] = 1;
        $ret = $this->get($table, $where, $columns);
        if (empty($ret) || !is_array($ret)) {
            return array();
        }

        return $ret[0];
    }

    private function connect(& $mysql, $reconn = false)
    {
        if ($reconn) {
            $mysql->close();
        }

        $options = array();
        $options['host'] = $this->host;
        $options['port'] = intval($this->port) == 0 ? 3306 : intval($this->port);
        $options['user'] = $this->username;
        $options['password'] = $this->password;
        $options['database'] = $this->dbname;
        $ret = $mysql->connect($options);
        return $ret;
    }

    private function real_query(& $mysql, & $sql, & $map)
    {
        if (empty($map)) {
            return $mysql->query($sql);
        } else {
            $stmt = $mysql->prepare($sql);

            if ($stmt == false) {
                return false;
            } else {
                return $stmt->execute($map);
            }
        }
    }

    function query($sql, $map = null, & $mysql = null)
    {
        if (empty($sql)) {
            throw new RuntimeException("input_empty_query_sql");
        }

        try {
            $mysql = $this->pool->pop();
            $ret = $this->real_query($mysql, $sql, $map);

            if ($ret === false) {
                if ($mysql->errno != 2006 && $mysql->errno != 2013) {
                    Logger::error("MySQL QUERY FAIL [" . $mysql->errno . "][" . $mysql->error . "], sql=[{$sql}], map=[" . json_encode($map) . "]", 'mysql_log');
                }

                if ($mysql->errno == 2006 || $mysql->errno == 2013) {
                    //重连MySQL
                    $ret = $this->connect($mysql, true);
                    if ($ret) {
                        $ret = $this->real_query($mysql, $sql, $map);
                    } else {
                        throw new RuntimeException("reconnect fail: [" . $mysql->errno . "][" . $mysql->error . "], host={$this->host}, port={$this->port}, user={$this->username}, database={$this->dbname}");
                    }
                }
            }

            if ($ret === false) {
                throw new RuntimeException($mysql->errno . "|" . $mysql->error);
            }

            $this->pool->push($mysql);
            return $ret;
        } catch (Exception $e) {
            $this->pool->push($mysql);
            Logger::error("MySQL catch exception [" . $e->getMessage() . "], sql=[{$sql}], map=" . json_encode($map), 'mysql_log');
            throw new RuntimeException("MySQL catch exception [" . $e->getMessage() . "], sql=[{$sql}], map=" . json_encode($map));
        }
    }
}
