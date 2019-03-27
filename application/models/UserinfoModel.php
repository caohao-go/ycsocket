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

    function register_user($appid, $userid, $open_id, $session_key) {
        $data = array();
        $data['appid'] = $appid;
        $data['user_id'] = $userid;
        $data['open_id'] = $open_id;
        $data['session_key'] = $session_key;
        $data['last_login_time'] = $data['regist_time'] = date('Y-m-d H:i:s', time());
        $data['token'] = md5(TOKEN_GENERATE_KEY . time() . $userid . $session_key);
        $ret = $this->db->insert("user_info", $data);
        if ($ret != -1) {
            return $data['token'];
        } else {
            $this->util_log->LogError("error to insert_trade_info, DATA=[".json_encode($data)."]");
            return false;
        }
    }

    function login_user($userid, $session_key) {
        $data = array();
        $data['user_id'] = $userid;
        $data['session_key'] = $session_key;
        $data['last_login_time'] = date('Y-m-d H:i:s', time());
        $data['token'] = md5(TOKEN_GENERATE_KEY . time() . $userid . $session_key);

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

    function getUserInZoneUserids($zone_userids) {
        $ret = array();

        if (empty($zone_userids)) {
            return $ret;
        }

        $zone_users = $this->db->get('zone_user', ['zone_user_id' => $zone_userids]);
        if (empty($zone_users) || $zone_users == -1) {
            return array();
        }

        $user_ids = array_column($zone_users, 'user_id');
        $userinfos = $this->db->get('user_info', ['user_id' => $user_ids], "user_id,nickname,avatar_url,city");
        if (empty($userinfos) || $userinfos == -1) {
            return array();
        }

        $userinfos_array = array();
        foreach($userinfos as $value) {
            $userinfos_array[$value['user_id']] = $value;
        }

        $zoneinfo = Loader::config('zoneinfo')['source_zone'];

        foreach($zone_users as $value) {
            $user_id = $value['user_id'];
            $zone_user_id = $value['zone_user_id'];
            $zone_id = $value['zone_id'];

            $userinfo = $userinfos_array[$user_id];
            if (!empty($userinfo['nickname'])) {
                $zone = $zoneinfo[$zone_id] == 0 ? "" : "[".$zone_id."区]";
                $userinfo['nickname'] = $zone . $userinfo['nickname'];
            }

            unset($value['id']);
            unset($value['updatetime']);
            unset($value['zone_user_id']);
            unset($value['zone_user_id']);
            $value['user_id'] = $zone_user_id;
            $value['nickname'] = $userinfo['nickname'];
            $value['avatar_url'] = $userinfo['avatar_url'];
            $value['city'] = $userinfo['city'];
            $ret[$zone_user_id] = $value;
        }

        return $ret;
    }

    function getUserZoneUid($userid, $zone_id) {
        $zone_id = $zone_id <= 0 ? 1 : $zone_id;

        $data = $this->db->get_one('zone_user', ['user_id' => $userid, 'zone_id' => $zone_id]);
        if ($data != -1) {
            return intval($data['zone_user_id']);
        }

        return 0;
    }

    function insertUserZoneUid($user_id, $zone_id, $zone_user_id) {
        $data = array();
        $data['user_id'] = $user_id;
        $data['zone_id'] = $zone_id;
        $data['zone_user_id'] = $zone_user_id;

        $ret = $this->db->insert('zone_user', $data);

        if ($ret == -1) {
            $this->util_log->LogError("error to insertUserZoneUid , DATA=[".json_encode($data)."]");
            return 0;
        }

        return intval($ret);
    }

    function getUidByZoneUserId($zone_user_id) {
        $redis = $this->loader->redis('userinfo');
        $redis_key = "pre_zone_userid" . $zone_user_id;
        $userid = intval($redis->get($redis_key));
        if ($userid != 0) {
            return $userid;
        }

        $ret = $this->db->get_one('zone_user', ['zone_user_id' => $zone_user_id]);
        $userid = intval($ret['user_id']);
        if ($userid == 0) { //2次获取
            $ret = $this->db->get_one('zone_user', ['zone_user_id' => $zone_user_id]);
            $userid = intval($ret['user_id']);
            if ($userid == 0) {
                return 0;
            }
        }

        $redis->set($redis_key, $userid);
        $redis->expire($redis_key, 86400);

        return $userid;
    }

    function getUserZoneid($zone_user_id) {
        $redis = $this->loader->redis('userinfo');
        $redis_key = "pre_user_zone_id" . $zone_user_id;
        $zone_id = intval($redis->get($redis_key));
        if ($zone_id != 0) {
            return $zone_id;
        }

        $ret = $this->db->get_one('zone_user', ['zone_user_id' => $zone_user_id]);
        $zone_id = intval($ret['zone_id']);
        if ($zone_id == 0) { //2次获取
            $ret = $this->db->get_one('zone_user', ['zone_user_id' => $zone_user_id]);
            $zone_id = intval($ret['zone_id']);
            if ($zone_id == 0) {
                return 0;
            }
        }

        $zoneinfo = Loader::config('zoneinfo');
        if ($zoneinfo['source_zone'][$zone_id] == 0) {
            $redis->set($redis_key, 0);
            $redis->expire($redis_key, 86400);
            return 0;
        }

        $redis->set($redis_key, $zone_id);
        $redis->expire($redis_key, 86400);
        return $zone_id;
    }

    function getZoneNickname($userId, $nickname) {
        $zone_id = $this->getUserZoneid($userId);
        $zone = $zone_id == 0 ? "" : "[".$zone_id."区]";
        return $zone . $nickname;
    }

    function getUidsByZoneUserids($zone_user_ids) {
        $redis = $this->loader->redis('userinfo');
        $redis_key = "pre_zone_userids_" . md5(serialize($zone_user_id));
        $userids = $redis->get($redis_key);
        if (!empty($userids)) {
            return unserialize($userids);
        }

        $userids = $this->db->get('zone_user', ['zone_user_id' => $zone_user_ids], 'user_id');
        if (empty($userids) || $userids == -1) {
            return array();
        }

        $redis->set($redis_key, serialize($userids));
        $redis->expire($redis_key, 3600);

        return $userids;
    }
}
