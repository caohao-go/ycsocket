<?php
/**
 * DatabaseProxy class
 *
 * @package			Ycsocket
 * @subpackage		database
 * @category		DatabaseProxy
 * @author			caohao
 */
class DatabaseProxy {
    var $ycdb = null;
    var $unix_socket = '';
    var $host = '';
    var $username = '';
    var $password = '';
    var $dbname = '';
    var $port = 3306;
    var $pconnect = FALSE;
    var $db_debug = FALSE;
    var $char_set = 'utf8';
    var $dbcollat = 'utf8_general_ci';
    var $autoinit = FALSE;
    var $init_flag = FALSE;
    var $db_log;

    function __construct($params, & $db_log) {
        if (!defined("DB_FAILURE")) define("DB_FAILURE", -1);

        if (is_array($params)) {
            foreach ($params as $key => $val) {
                $this->$key = $val;
            }
        }
        
        $this->db_log = & $db_log;
    }

    function initialize() {
        if ($this->init_flag && is_object($this->ycdb)) {
            return TRUE;
        }

        if (!empty($this->unix_socket)) {
            $this->ycdb = new ycdb(["unix_socket" => $this->unix_socket]);
        } else {
            $options = array();
            $options['host'] = $this->host;
            $options['username'] = $this->username;
            $options['password'] = $this->password;
            $options['dbname'] = $this->dbname;
            $options['port'] = intval($this->port) == 0 ? 3306 : intval($this->port);
            $options['option'] = array(PDO::ATTR_CASE => PDO::CASE_NATURAL, PDO::ATTR_TIMEOUT => 2);
            if ($this->pconnect) $options['option'][PDO::ATTR_PERSISTENT] = true;

            $this->ycdb = new ycdb($options);
        }

        try {
            $this->ycdb->initialize();
        } catch (PDOException $e) {
            return $this->handle_error("ycdb initialize error : " . $e->getMessage());
        }

        $set_charset_sql = "SET NAMES '".$this->char_set."' COLLATE '".$this->dbcollat."'";
        $ret = $this->ycdb->exec($set_charset_sql);
        if ($ret == DB_FAILURE) {
            return $this->handle_error();
        }

        $this->init_flag = TRUE;
        return TRUE;
    }

    function query($sql) {
        $this->initialize();

        if (empty($sql)) {
            return $this->handle_error("input_empty_query_sql");
        }

        try {
            if ($this->is_write_type($sql)) {
                $ret = $this->ycdb->exec($sql);
            } else {
                $ret = $this->ycdb->query($sql);
            }
        } catch (PDOException $e) {
            return $this->handle_error("ycdb query error : " . $e->getMessage());
        }

        if ($ret == DB_FAILURE) {
            return $this->handle_error();
        }

        return $ret;
    }

    function insert($table = '', $data = NULL, $ignore = false) {
        $this->initialize();

        if (empty($table) || empty($data) || !is_array($data)) {
            return $this->display_error('insert_table_or_data_must_be_set');
        }

        try {
            $ret = $this->ycdb->insert($table, $data);
        } catch (PDOException $e) {
            return $this->handle_error("ycdb insert error : " . $e->getMessage());
        }

        if ($ret == DB_FAILURE) {
            return $this->handle_error();
        }

        return $ret;
    }

    function replace($table = '', $data = NULL) {
        $this->initialize();

        if (empty($table) || empty($data) || !is_array($data)) {
            return $this->display_error('replace_table_or_data_must_be_set');
        }

        try {
            $ret = $this->ycdb->replace($table, $data);
        } catch (PDOException $e) {
            return $this->handle_error("ycdb replace error : " . $e->getMessage());
        }

        if ($ret == DB_FAILURE) {
            return $this->handle_error();
        }

        return $ret;
    }

    function update($table = '', $where = NULL, $data = NULL) {
        $this->initialize();

        if (empty($table) || empty($where) || empty($data) || !is_array($data)) {
            return $this->display_error('update_table_or_data_must_be_set');
        }

        try {
            $ret = $this->ycdb->update($table, $data, $where);
        } catch (PDOException $e) {
            return $this->handle_error("ycdb update error : " . $e->getMessage());
        }

        if ($ret == DB_FAILURE) {
            return $this->handle_error();
        }

        return $ret;
    }

    function delete($table = '', $where = NULL) {
        $this->initialize();

        if (empty($table) || empty($where)) {
            return $this->display_error('delete_table_or_where_must_be_set');
        }

        try {
            $ret = $this->ycdb->delete($table, $where);
        } catch (PDOException $e) {
            return $this->handle_error("ycdb delete error : " . $e->getMessage());
        }

        if ($ret == DB_FAILURE) {
            return $this->handle_error();
        }

        return $ret;
    }

    function get($table = '', $where = array(), $columns = "*") {
        $this->initialize();

        if (empty($table) || empty($columns)) {
            return $this->display_error('select_table_or_columns_must_be_set');
        }

        try {
            $ret = $this->ycdb->select($table, $columns, $where);
        } catch (PDOException $e) {
            return $this->handle_error("ycdb delete error : " . $e->getMessage());
        }

        if ($ret == DB_FAILURE) {
            return $this->handle_error();
        }

        return $ret;
    }

    function get_one($table = '', $where = array(), $columns = "*") {
        $this->initialize();

        if (empty($table) || empty($columns)) {
            return $this->display_error('select_table_or_columns_must_be_set');
        }

        $where['LIMIT'] = 1;

        try {
            $ret = $this->ycdb->select($table, $columns, $where);
        } catch (PDOException $e) {
            return $this->handle_error("ycdb delete error : " . $e->getMessage());
        }

        if ($ret == DB_FAILURE) {
            return $this->handle_error();
        }

        if (empty($ret[0])) {
            return array();
        }

        return $ret[0];
    }

    function handle_error($input_error = null) {
        if (empty($input_error)) {
            $info = $this->ycdb->errorInfo();
            $error_code = $info[0];
            $error_no = $info[1];
            $error_msg = $info[2];

            $this->db_log->LogError("Query Error: [$error_code][$error_no][$error_msg]");

            if ($this->db_debug) {
                $this->display_error(array("DBErrorNo: $error_no", "Message: [$error_code] $error_msg"));
            }
        } else {
            $this->db_log->LogError($input_error);

            if ($this->db_debug) {
                $this->display_error($input_error);
            }
        }

        return DB_FAILURE;
    }

    /**
     * Display an error message
     *
     * @access    public
     * @param    string    the error message
     * @param    string    any "swap" values
     * @param    boolean    whether to localize the message
     * @return    string    sends the application/error_db.php template
     */
    function display_error($error = '', $swap = '', $native = FALSE) {
        $heading = "A DB Error Occured";
        if ($native == TRUE) {
            $message = $error;
        } else {
            if (!is_array($error)) {
                $message[] = $error;
            } else {
                $message = $error;
            }
        }

        // Find the most likely culprit of the error by going through
        // the backtrace until the source file is no longer in the
        // database folder.
        $trace = debug_backtrace();

        foreach ($trace as $call) {
            if (isset($call['file']) && strpos($call['file'], BASEPATH.'/DatabaseProxy') === FALSE) {
                // Found it - use a relative path for safety
                $message[] = 'Filename: '.str_replace(array(BASEPATH, APPPATH), '', $call['file']);
                $message[] = 'Line Number: '.$call['line'];

                break;
            }
        }

        show_error($message, 500, $heading);
    }

    function is_write_type($sql) {
        if ( ! preg_match('/^\s*"?(SET|INSERT|UPDATE|DELETE|REPLACE|CREATE|DROP|TRUNCATE|LOAD DATA|COPY|ALTER|GRANT|REVOKE|LOCK|UNLOCK)\s+/i', $sql)) {
            return FALSE;
        }
        return TRUE;
    }
}
