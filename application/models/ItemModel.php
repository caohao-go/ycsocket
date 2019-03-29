<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');
/**
 * ExampleModel Class
 *
 * @package			Ycsocket
 * @subpackage		Model
 * @category		Example Model
 * @author			caohao
 */
class ItemModel extends CoreModel {
    const ERROR_ITEM_NOT_SO_MUCH = "NOT_SO_MUCH"; //item不够

    public function init() {
        $this->util_log = $this->loader->logger('item_log');
        $this->item = $this->loader->entity("Items");
        $this->db_name = "shine_light";
    }

    public function sequence() {
        return RedisPool::instance("sequence")->incr("useritem_sequence");
    }

    public function get_user_items($user_id) {
        $redis_key = 'user_items_' . $user_id;

        $hkeys = RedisPool::instance("useritem")->hKeys($redis_key);

        if (empty($hkeys)) {
            return array();
        }

        $result = array();
        foreach($hkeys as $key) {
            $key_array = explode("_", $key);
            $key_data = RedisPool::instance("useritem")->hGet($redis_key, $key);
            $val_array = json_decode($key_data, true);

            foreach($val_array as $k => $v) {
                $result[] = ["item_id" => $key_array[0],  "open" => $key_array[1], "num" => $v['num'], "id" => $v['id']];
            }
        }

        return $result;
    }

    public function insert_user_items($user_id, $item_id, $num = 1, $open_flag = 1) {
        $item_info = $this->item->get_item_by_id($item_id);

        if (empty($item_info)) {
            $this->util_log->LogError("insert_user_items, not find item_id [$item_id]");
            return false;
        }

        $stack = intval($item_info['stack']);

        if ($stack == 0) {
            $this->util_log->LogError("insert_user_items, stack is zero [$item_id]");
            return false;
        }

        $ret = true;

        $redis_key = 'user_items_' . $user_id;
        $redis_hkey = $item_id . "_" . $open_flag;

        $data = RedisPool::instance("useritem")->hGet($redis_key, $redis_hkey);

        if (empty($data)) {
            $data = array();
        } else {
            $data = json_decode($data, true);
        }

        if ($stack == 1) { //若叠加上限＞1，则代表玩家持有这个物品id最多拥有的数量限制，若叠加上限 = 1，则代表玩家可持有多个物品，但每个物品在背包内会占一个格子。
            for ($i = 0; $i < $num; $i ++) {
                $data[] = ["num" => 1, "id" => $this->sequence()];
            }
        } else {
            if (!empty($data)) {
                foreach($data as $key => $val) {
                    if ($val['num'] < $stack) {
                        if ($val['num'] + $num <= $stack) {
                            $data[$key]['num'] = $val['num'] + $num;
                            $num = 0;
                            break;
                        } else {
                            $data[$key]['num'] = $stack;
                            $num = $num - $stack + $val['num'];
                        }
                    }
                }
            }

            if ($num > 0) {
                $n = intval($num / $stack);
                if ($n > 0) {
                    for ($i = 0; $i < $n; $i ++) {
                        $data[] = ["num" => $stack, "id" => $this->sequence()];
                    }
                }

                $m = $num % $stack;
                if ($m > 0) {
                    $data[] = ["num" => $m, "id" => $this->sequence()];
                }
            }
        }

        $ret = RedisPool::instance("useritem")->hSet($redis_key, $redis_hkey, json_encode($data));

        return true;
    }

}
