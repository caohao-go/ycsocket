<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');
/** Global.php
 * SuperModel Class
 *
 * @package			Ycsocket  https://github.com/caohao-php/ycsocket
 * @subpackage		Global
 * @category		Global
 * @author			caohao
 */
class GlobalEntity {
    static public function set($id, $data, $expire = 0) {
        if (intval($expire) != 0) {
            $expire = intval($expire) + time();
        }

        global $globalTable;
        return $globalTable->set($id, array("data"=> base64_encode(serialize($data)), "int_data" => 0, "expire" => intval($expire)));
    }

    static public function get($id) {
        global $globalTable;
        $data = $globalTable->get($id);
        if (empty($data)) {
            return;
        } else {
            if ($data["expire"] != 0 && time() > $data["expire"]) {
                $globalTable->del($id);
                return;
            } else {
                return unserialize(base64_decode($data['data']));
            }
        }
    }

    static public function del($id) {
        global $globalTable;
        $globalTable->del($id);
    }

    static public function incr($id, $incrby = 1) {
        global $globalIntTable;
        $data = $globalIntTable->incr($id, 'data', $incrby);
        return $data;
    }

    static public function set_int($id, $data, $expire = 0) {
        if (intval($expire) != 0) {
            $expire = intval($expire) + time();
        }

        global $globalIntTable;
        return $globalIntTable->set($id, array("data" => intval($data), "expire" => intval($expire)));
    }

    static public function get_int($id) {
        global $globalIntTable;
        $data = $globalIntTable->get($id);
        if (empty($data)) {
            return null;
        } else {
            if ($data["expire"] != 0 && time() > $data["expire"]) {
                $globalIntTable->del($id);
                return null;
            } else {
                return $data['data'];
            }
        }
    }

    static public function del_int($id) {
        global $globalIntTable;
        $globalIntTable->del($id);
    }
}
