<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');

include_once(APP_ROOT . "/config/global.php");

class GlobalEntity {
    static private $instance;
    public $global;

    private function __construct(& $global) {
        $this->global = & $global;
    }

    static public function & instance(& $global = null) {
        if (empty(self::$instance)) {
            if (empty($global)) {
                return false;
            }

            self::$instance = new GlobalEntity($global);
        }
        return self::$instance;
    }

    static public function set($id, $data) {
        if (is_array($data)) {
            $data = json_encode($data);
        } else if (is_object($data)) {
            $data = json_encode($data, JSON_FORCE_OBJECT);
        }

        $instance = self::instance();
        return $instance->global->set("index_{$id}", array("global_id" => $id, "data"=> $data));
    }

    static public function get($id) {
        $instance = self::instance();
        $data = $instance->global->get("index_{$id}");
        if (empty($data)) {
            return;
        } else {
            return json_decode($data['data'], true);
        }
    }

    static public function del($id) {
        $instance = self::instance();
        $data = $instance->global->del("index_{$id}");
    }

    static public function incr($id, $column, $incrby = 1) {
        $instance = self::instance();
        $data = $instance->global->incr("index_{$id}", $column, $incrby);
    }
}

$table = new swoole_table(1024);
$table->column('global_id', swoole_table::TYPE_INT);
$table->column('data', swoole_table::TYPE_STRING, 2048);
$table->create();
return $table;
