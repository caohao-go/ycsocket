<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');

class GlobalEntity {
    static private $instance;
    public $table;

    private function __construct(& $table) {
        $this->table = & $table;
    }

    static public function & instance(& $table = null) {
        if (empty(self::$instance)) {
            if (empty($table)) {
                return false;
            }

            self::$instance = new GlobalEntity($table);
        }
        return self::$instance;
    }

    static public function set($id, $data, $expire = 0) {
        if (intval($expire) != 0) {
            $expire = intval($expire) + time();
        }

        return self::$instance->table->set($id, array("data"=> $data, "int_data" => 0, "expire" => intval($expire)));
    }

    static public function get($id) {
        $data = self::$instance->table->get($id);
        if (empty($data)) {
            return;
        } else {
            if ($data["expire"] != 0 && time() > $data["expire"]) {
                self::$instance->table->del($id);
                return;
            } else {
                return $data['data'];
            }
        }
    }

    static public function del($id) {
        self::$instance->table->del($id);
    }

    static public function incr($id, $incrby = 1) {
        $data = self::$instance->table->incr($id, 'data', $incrby);
        return $data;
    }

    static public function set_int($id, $data, $expire = 0) {
        if (intval($expire) != 0) {
            $expire = intval($expire) + time();
        }

        return self::$instance->table->set($id, array("data"=> "", "int_data" => $data, "expire" => intval($expire)));
    }

    static public function get_int($id) {
        $data = self::$instance->table->get($id);
        if (empty($data)) {
            return null;
        } else {
            if ($data["expire"] != 0 && time() > $data["expire"]) {
                self::$instance->table->del($id);
                return null;
            } else {
                return $data['int_data'];
            }
        }
    }
}

$table = new swoole_table(16384);
$table->column('int_data', swoole_table::TYPE_INT);
$table->column('data', swoole_table::TYPE_STRING);
$table->column('expire', swoole_table::TYPE_INT);
$table->create();
return $table;
