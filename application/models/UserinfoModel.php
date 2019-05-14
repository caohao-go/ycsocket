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
        $this->util_log = $this->loader->logger('userinfo_log');
    }

    function generate_userid() {
        $cur_time = time();
        $data = array('time' => $cur_time);

        $sequence_no = 0;
        for ($i = 0; $i < 3; $i++) {
            $sequence_no = MySQLPool::instance("default")->insert('sequence', $data);
            if ($sequence_no > 0) {
                break;
            }
        }

        if ($sequence_no) {
            $time_from_cur = $cur_time - 1554557274; // 2019-04-22
            $userid = sprintf("%s%03d", substr($time_from_cur, 0, -2), substr($sequence_no, -3));
            return intval($userid);
        }
    }

    function register_user($appid, $userid, $open_id) {
        $data = array();
        $data['appid'] = $appid;
        $data['user_id'] = $userid;
        $data['open_id'] = $open_id;
        $data['last_login_time'] = $data['regist_time'] = date('Y-m-d H:i:s', time());
        $data['token'] = md5(TOKEN_GENERATE_KEY . time() . $userid . $session_key);
        $ret = MySQLPool::instance("default")->insert("user_info", $data);
        if ($ret > 0) {
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

        $ret = MySQLPool::instance("default")->update("user_info", ["user_id" => $userid], $data);

        if ($ret) {
            return $data['token'];
        } else {
            $this->util_log->LogError("error to login_user, DATA=[".json_encode($data)."]");
            return false;
        }
    }

    function update_user($userid, $update_data) {
        RedisPool::instance("userinfo")->del("pre_redis_user_info_" . $userid);


        $ret = MySQLPool::instance("default")->update("user_info", ["user_id" => $userid], $update_data);
        if ($ret) {
            return true;
        } else {
            $this->util_log->LogError("error to update_user, DATA=[".json_encode($update_data)."]");
            return false;
        }
    }

    function get_one_user_info_by_key($key, $value) {
        $ret = MySQLPool::instance("default")->get_one("user_info", [$key => $value]);
        return $ret;
    }

    public function getUserinfoByUserid($user_id) {
        $redis_key = "pre_redis_user_info_" . $user_id;

        $userInfo = RedisPool::instance("userinfo")->get($redis_key);

        if (empty($userInfo)) {
            $userInfo = $this->get_one_user_info_by_key('user_id', $user_id);

            if (!empty($userInfo)) {
                RedisPool::instance("userinfo")->set($redis_key, serialize($userInfo));
                RedisPool::instance("userinfo")->expire($redis_key, 900);
            }
        } else {
            $userInfo = unserialize($userInfo);
        }

        return $userInfo;
    }

    public function getUserByName($nickname) {
        return MySQLPool::instance("default")->query("select * from user_info where nickname like '%$nickname%'");
    }

    function getUserInUserids($userids) {
        $ret = array();

        if (empty($userids)) {
            return $ret;
        }

        $result = MySQLPool::instance("default")->get('user_info', ['user_id' => $userids], "user_id,nickname,avatar_url,city");
        if (!empty($result)) {
            foreach($result as $value) {
                $ret[$value['user_id']] = $value;
            }
        }

        return $ret;
    }
}
