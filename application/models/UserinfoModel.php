<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');
/**
 * UserinfoModel 	Class
 *
 * @package			Ycsocket
 * @subpackage		Model
 * @category		UserinfoModel
 * @author			caohao
 */
class UserinfoModel extends SuperModel {
    public function init() {
        $this->db = $this->loader->database('default');
        $this->util_log = $this->loader->logger('userinfo_log');
    }

    function generate_userid() {
        $cur_time = time();
        $data = array('time' => $cur_time);

        $sequence_no = 0;
        for ($i = 0; $i < 3; $i++) {
            $sequence_no = $this->db->insert('sequence', $data);
            if ($sequence_no != -1) {
                break;
            }
        }

        if ($sequence_no) {
            $userid = sprintf("%s%03d", $cur_time, substr($sequence_no, -3));
            return intval($userid);
        }
    }

    function register_user($appid, $userid, $open_id) {
        $data = array();
        $data['appid'] = $appid;
        $data['user_id'] = $userid;
        $data['open_id'] = $open_id;
        $data['last_login_time'] = $data['regist_time'] = date('Y-m-d H:i:s', time());
        $data['token'] = md5(TOKEN_GENERATE_KEY . time() . $userid);
        $ret = $this->db->insert("user_info", $data);
        if ($ret != -1) {
            return $data['token'];
        } else {
            $this->util_log->LogError("error to insert_trade_info, DATA=[".json_encode($data)."]");
            return false;
        }
    }

    function login_user($userid) {
        $data = array();
        $data['user_id'] = $userid;
        $data['last_login_time'] = date('Y-m-d H:i:s', time());
        $data['token'] = md5(TOKEN_GENERATE_KEY . time() . $userid);

        $ret = $this->db->update("user_info", ["user_id" => $userid], $data);

        if ($ret != -1) {
            return $data['token'];
        } else {
            $this->util_log->LogError("error to login_user, DATA=[".json_encode($data)."]");
            return false;
        }
    }

    function update_user($userid, $update_data) {
        $redis = $this->loader->redis("userinfo");
        $redis->del("pre_redis_user_info_" . $userid);


        $ret = $this->db->update("user_info", ["user_id" => $userid], $update_data);
        if ($ret != -1) {
            return true;
        } else {
            $this->util_log->LogError("error to update_user, DATA=[".json_encode($update_data)."]");
            return false;
        }
    }

    function get_one_user_info_by_key($key, $value) {
        $ret = $this->db->get_one("user_info", [$key => $value]);
        if ($ret == -1) {
            return array();
        }
        return $ret;
    }

    public function getUserinfoByUserid($user_id) {
        $redis_key = "pre_redis_user_info_" . $user_id;
        $redis = $this->loader->redis("userinfo");
        if (!empty($redis)) {
            $userInfo = $redis->get($redis_key);
        }

        if (empty($userInfo)) {
            $userInfo = $this->get_one_user_info_by_key('user_id', $user_id);

            if (!empty($userInfo)) {
                $redis->set($redis_key, serialize($userInfo));
                $redis->expire($redis_key, 900);
            }
        } else {
            $userInfo = unserialize($userInfo);
        }

        return $userInfo;
    }

    public function getUserByName($nickname) {
        return $this->db->query("select * from user_info where nickname like '%$nickname%'");
    }

    function getUserInUserids($userids) {
        $ret = array();

        if (empty($userids)) {
            return $ret;
        }

        $result = $this->db->get('user_info', ['user_id' => $userids], "user_id,nickname,avatar_url,city");
        if (!empty($result) && $result != -1) {
            foreach($result as $value) {
                $ret[$value['user_id']] = $value;
            }
        }

        return $ret;
    }
}
