<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');

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

        global $globalTable;
        return $globalTable->set($id, array("data"=> "", "int_data" => $data, "expire" => intval($expire)));
    }

    static public function get_int($id) {
        global $globalTable;
        $data = $globalTable->get($id);
        if (empty($data)) {
            return null;
        } else {
            if ($data["expire"] != 0 && time() > $data["expire"]) {
                $globalTable->del($id);
                return null;
            } else {
                return $data['int_data'];
            }
        }
    }
}
