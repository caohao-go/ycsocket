<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');
/**
 * ExampleModel Class
 *
 * @package        SuperCI
 * @subpackage    Model
 * @category      Example Model
 * @author        caohao
 */
class UserinfoModel extends SuperModel
{
	public $test = "";
	
    public function init() {
        $this->db = $this->loader->database('default');
        $this->util_log = $this->loader->logger('userinfo_log');
    }
    
    function generate_userid() {
        $cur_time = time();
        $data = array('time' => $cur_time);
            
        $ret = 0;
        for($i = 0; $i < 3; $i++) {
            $ret = $this->db->insert('sequence', $data);
            if($ret) {
                break;
            }
        }

        if($ret) {
            $query = $this->db->query('SELECT LAST_INSERT_ID() as sequence_no');
            $sequence_no = $query->result_array();
            if(intval($sequence_no[0]['sequence_no']) == 0) return;
            $time_from_cur = $cur_time - 1529596800; // 2018-06-22
            $userid = sprintf("%s%03d", substr($time_from_cur, 0, -2), substr($sequence_no[0]['sequence_no'], -3));
            return $userid;
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
        if ($ret && $this->db->affected_rows()) {
            return $data['token'];
        }else{
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
        
        $ret = $this->db->where("user_id", $userid)
                        ->update("user_info", $data);
                        
        if ($ret && $this->db->affected_rows()) {
            return $data['token'];
        }else{
            $this->util_log->LogError("error to login_user, DATA=[".json_encode($data)."]");
            return false;
        }
    }
    
    function update_user($userid, $update_data) {
        $redis = $this->loader->redis("userinfo");
        $redis->del("pre_redis_user_info_" . $userid);
        
        $ret = $this->db->where("user_id", $userid)
                        ->update("user_info", $update_data);
        
        if ($ret) {
            return true;
        }else{
            $this->util_log->LogError("error to login_user, DATA=[".json_encode($update_data)."]");
            return false;
        }
    }
    
    function get_one_user_info_by_key($key, $value) {
        $data = $this->db->where($key, $value)
                        ->get("user_info")
                        ->row_array();
		return $data;
    }
    
    public function getUserinfoByUserid($user_id) {
        $redis_key = "pre_redis_user_info_" . $user_id;
        $redis = $this->loader->redis("userinfo");
        if(!empty($redis)) {
            $userInfo = $redis->get($redis_key);
        }
        
        if(empty($userInfo)) {
            $userInfo = $this->get_one_user_info_by_key('user_id', $user_id);
            
            if(!empty($userInfo)) {
                $redis->set($redis_key, serialize($userInfo));
                $redis->expire($redis_key, 900);
            }
        } else {
            $userInfo = unserialize($userInfo);
        }
        
        return $userInfo;
    }
    
    function getUserAndAuth($user_id, $token,& $out) {
        if(empty($user_id)) {
            $out = array('errno' => 99900031, 'errmsg' => 'user id is empty');
            return;
        }
        
        $userInfo = $this->getUserinfoByUserid($user_id);
        if(empty($userInfo)) {
            $out = array('errno' => 99900032, 'errmsg' => 'not find user');
            return;
        }
        
        if(empty($token) || $token != $userInfo['token']) {
            $out = array('errno' => 99900033, 'errmsg' => 'token is invalid');
            return;
        }
        return $userInfo;
    }
    
    function getUserInUserids($userids) {
    	$ret = array();
    	$result = $this->db->select('user_id,nickname,avatar_url,city')->where_in('user_id', $userids)->get('user_info')->result_array();
    	if(!empty($result)) {
    		foreach($result as $value) {
    			$ret[$value['user_id']] = $value;
    		}
    	}
    	
    	return $ret;
   	}
    
    public function get_user_rand($num = 10, $userId) {
    	return $this->db->query("select * from user_info where user_id!=$userId order by rand() limit $num")->result_array();
    }
    
    function getUidByZoneUserId($zone_user_id) {
        $redis = $this->loader->redis('userinfo');
        $redis_key = "pre_zone_userid" . $zone_user_id;
        $userid = intval($redis->get($redis_key));
        if($userid != 0) {
        	return $userid;
    	}
    	
    	$ret = $this->db->get_one('zone_user2', ['zone_user_id' => $zone_user_id]);
    	$userid = intval($ret['user_id']);
    	if($userid == 0) {
    		$ret = $this->db->get_one('zone_user', ['zone_user_id' => $zone_user_id]);
    		$userid = intval($ret['user_id']);
    		if($userid == 0) {
    			return 0;
    		}
    	}
    	
        $redis->set($redis_key, $userid);
        $redis->expire($redis_key, 86400);
        
        return $userid;
    }
    
}