<?php

class Items extends Entity {
    const EMPTY_STRING = "-999999999";
    const GLOBAL_PRE_ITEMS = 'global_pre_items_';

    public function init() {
    }

    public function get_item_by_id($id) {
        $data = $this->getGlobal(Items::GLOBAL_PRE_ITEMS . $id);
        if ($data == self::EMPTY_STRING) {
            return;
        } else if (!empty($data)) {
            return json_decode($data, true);
        }

        $db = $this->loader->database('shine_light');
        $data = $db->get_one("items", ['id' => $id]);

        if ($data == -1) {
            return;
        } else if (!empty($data)) {
            $this->setGlobal(Items::GLOBAL_PRE_ITEMS . $id, json_encode($data), 300);
            return $data;
        } else {
            $this->setGlobal(Items::GLOBAL_PRE_ITEMS . $id, self::EMPTY_STRING, 100);
            return;
        }
    }
}
