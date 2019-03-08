<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');
/**
 * TestController Class
 *
 * @package       SuperCI
 * @subpackage    Controller
 * @category      TestController
 * @author        caohao
 */
class DabaojianController extends SuperController {
	
    public function init() {
        $this->userinfo_model = $this->loader->model('UserinfoModel');
        $this->util_log = $this->loader->logger('dabaojian_log');
    }
    
    //公告
    public function gonggaoAction() {
    	$userId = $this->params['userid'];
    	$content = $this->params['content'];
    	
    	if(empty($content)) {
    		return $this->response_error(13342339, '内容不能为空');
    	}
    	
    	if($userId != 99999999) {
    		return $this->response_error(133425539, '内容不能为空');
    	}
    	
    	$result = array();
    	$result['content'] =  $content;
    	
    	return $this->response_success_to_all($result);
    }
    
    //聊天接口
    public function chatAction() {
    	$userId = $this->params['userid'];
    	$token = $this->params['token'];
    	$nickname = $this->params['nickname'];
    	$avatar_url = $this->params['avatar_url'];
    	$content = $this->params['content'];
    	$gender = $this->params['gender'];
    	
    	$userInfo = $this->userinfo_model->getUserAndAuth($this->userinfo_model->getUidByZoneUserId($userId), $token, $out);
        if (empty($userInfo)) {
            return $this->response_error($out['errno'], $this->params['c'] . '-' . $this->params['m'] . '-' . $out['errmsg']);
        }
    	
    	if(empty($content)) {
    		return $this->response_error(13342339, '内容不能为空');
    	}
    	
    	//是否黑名单
    	$exec_str = "grep $userId " . APPPATH . "/application/config/blacklist";
    	exec($exec_str, $res);
    	if(!empty($res)) {
    		$result['black'] = 1;
    		$result['content'] = "您已被禁言";
    		return $this->response_success_to_me($result);
    	}
    	
    	$sensitive = Loader::config("sensitive");
    	
    	$result = array();
    	$result['userid'] = $userId;
    	$result['nickname'] = $nickname;
    	$result['avatar_url'] = $avatar_url;
    	$result['gender'] = $gender;
    	$result['vip_level'] = $vip_level;
    	$result['content'] =  str_replace($sensitive, "*", $content);
    	
    	$redis = Loader::redis("default");
    	$redis->rpush("dabaojian_liaotian_history", serialize($result));
    	if(intval($redis->llen("dabaojian_liaotian_history")) > 50) {
    		$redis->lpop("dabaojian_liaotian_history");
    	}
    	
    	return $this->response_success_to_all($result);
    }
    
    
}
