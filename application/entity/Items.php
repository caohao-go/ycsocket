<?php

class Items {
    const GLOBAL_PRE_ITEMS = 'global_pre_items_';

    public static function get_item_by_id($id) {
        $data = GlobalEntity::get(Items::GLOBAL_PRE_ITEMS . $id);
        if (!empty($data)) {
            return $data;
        }

        $data = MySQLPool::instance("shine_light")->get_one("items", ['id' => $id]);

        if (!empty($data)) {
            GlobalEntity::set(Items::GLOBAL_PRE_ITEMS . $id, $data, 300);
            return $data;
        }
    }
}
