<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');
/**
 * UserController Class
 *
 * @package       Ycsocket
 * @subpackage    Controller
 * @category      UserController
 * @author        caohao
 */
class UserController extends SuperController {
    public function init() {
        parent::init();

        $this->util_log = $this->loader->logger('user_log');
        $this->weixin_model = $this->loader->model('WeixinModel');
        $this->user_model = $this->loader->model('UserinfoModel');
    }

    //修改用户信息
    public function f1Action() {
        $nickname = $this->params['nickname'];
        $result = $this->user_model->getUserByName($nickname);
        return $this->response_success_to_me($result);
    }

    //登陆接口
    public function loginAction() {
        //$this->util_log->LogInfo("f10001000Action:" . createLinkstringUrlencode($this->params));
        $appid = $this->params['appid'];
        $code = $this->params['code'];

        $result = array();
		
        $ret = $this->getAppConfig($appid);
        if (empty($ret['app_id'])) {
        	return $ret;
        }
        
        $out = '';
        $login_info = $this->weixin_model->get_openid($ret['app_id'], $ret['secret'], $code, $out);
        
        $openid = $login_info['openid'];
        if (empty($openid)) {
            return $this->response_error(10001006, $out);
        }

        $result['openid'] = $openid;
        $userInfo = $this->user_model->get_one_user_info_by_key('open_id', $openid);
        if (empty($userInfo)) {
            $result['userid'] = $this->user_model->generate_userid();
            $result['token'] = $this->user_model->register_user($appid, $result['userid'], $result['openid'], $login_info['session_key']);
        } else {
            $this->clearUserCache($userInfo['user_id']); //清理缓存
            $result['userid'] = $userInfo['user_id'];
            $result['token'] = $this->user_model->login_user($result['userid'], $login_info['session_key']);
        }

        if (empty($result['token'])) {
            return $this->response_error(10000009, "登陆失败, 请重试");
        }
        
        return $this->response_success_to_me($result);
    }

    //修改用户信息
    public function modifyInfoAction() {
        //$this->util_log->LogInfo("f10001001Action:" . createLinkstringUrlencode($this->params));
        $userId = $this->params['userid'];
        $token = $this->params['token'];
        $rawData = $this->params['rawData'];
        $encryptedData = $this->params['encryptedData'];
        $iv = $this->params['iv'];

        if (empty($userId)) {
            return $this->response_error(10000021, "user_id is empty");
        }

        if (empty($rawData)) {
            return $this->response_error(10000022, "rawData is empty");
        }
        
        $this->clearUserCache($userId); //清理缓存

        $userInfo = $this->user_model->getUserinfoByUserid($userId);
        if (empty($userInfo)) {
            return $this->response_error(10000023, "未找到该用户");
        }

        if (empty($token) || $token != $userInfo['token']) {
            return $this->response_error(10000024, "token 校验失败");
        }

        //更新用户资料
        $update_user_info = array();
        $raw_data_array = json_decode($rawData, true);
        $update_user_info['nickname'] = $raw_data_array['nickName'];
        $update_user_info['gender'] = $raw_data_array['gender'];
        $update_user_info['city'] = $raw_data_array['city'];
        $update_user_info['province'] = $raw_data_array['province'];
        $update_user_info['country'] = $raw_data_array['country'];
        $update_user_info['avatar_url'] = $raw_data_array['avatarUrl'];

        /*
        //解密敏感数据
        $session_key = $userInfo['session_key'];
        $appid = $userInfo['appid'];
        $appid = 'wx4f4bc4dec97d474b';
        $session_key = 'tiihtNczf5v6AKRyjwEUhQ==';
        $encryptedData="CiyLU1Aw2KjvrjMdj8YKliAjtP4gsMZMQmRzooG2xrDcvSnxIMXFufNstNGTyaGS9uT5geRa0W4oTOb1WT7fJlAC+oNPdbB+3hVbJSRgv+4lGOETKUQz6OYStslQ142dNCuabNPGBzlooOmB231qMM85d2/fV6ChevvXvQP8Hkue1poOFtnEtpyxVLW1zAo6/1Xx1COxFvrc2d7UL/lmHInNlxuacJXwu0fjpXfz/YqYzBIBzD6WUfTIF9GRHpOn/Hz7saL8xz+W//FRAUid1OksQaQx4CMs8LOddcQhULW4ucetDf96JcR3g0gfRK4PC7E/r7Z6xNrXd2UIeorGj5Ef7b1pJAYB6Y5anaHqZ9J6nKEBvB4DnNLIVWSgARns/8wR2SiRS7MNACwTyrGvt9ts8p12PKFdlqYTopNHR1Vf7XjfhQlVsAJdNiKdYmYVoKlaRv85IfVunYzO0IKXsyl7JCUjCpoG20f0a04COwfneQAGGwd5oa+T8yO5hzuyDb/XcxxmK01EpqOyuxINew==";
        $iv = 'r7BXXKkLb8qrSNn05n0qiA==';

        include_once APPPATH . "/application/library/smallgame/wxBizDataCrypt.php"; //引入微信api
        $pc = new WXBizDataCrypt($appid, $session_key);
        $errCode = $pc->decryptData($encryptedData, $iv, $data);
        if($errCode == 0) {
            $data = json_decode($data, true);
            $update_user_info['union'] = $data['unionId'];
        } else {
            $this->util_log->LogError("f10001001Action decryptData error errCode=[$errCode]");
        }
        */
        
        $this->user_model->update_user($userId, $update_user_info);
        return $this->response_success_to_me($result);
    }
	
    //获取用户信息接口
    public function getInfoAction() {
    	//$this->util_log->LogInfo("f10001002Action:" . createLinkstringUrlencode($this->params));
        $userId = $this->params['userid'];
        $token = $this->params['token'];

        if (empty($userId)) {
            return $this->response_error(10000017, "user_id is empty");
        }

        if (empty($token)) {
            return $this->response_error(10000016, "token is empty");
        }

        $userInfo = $this->user_model->getUserinfoByUserid($userId);
        if (empty($userInfo)) {
            return $this->response_error(10000023, "未找到该用户");
        }

        if (empty($token) || $token != $userInfo['token']) {
            return $this->response_error(10000024, "token 校验失败");
        }

        $data = array();
        $data['amount'] = $userInfo['amount'];
        $data['gender'] = intval($userInfo['gender']);
        $data['avatarUrl'] = $userInfo['avatar_url'];
        $data['nickname'] = $userInfo['nickname'];
        $data['form_id'] = $userInfo['form_id'];
        $data['userId'] = $userInfo['user_id'];
        
        return $this->response_success_to_me($data);
    }

    //获取他人用户信息接口
    public function	getOtherInfoAction() {
    	//$this->util_log->LogInfo("f10001003Action:" . createLinkstringUrlencode($this->params));
        $userId = $this->params['userid'];
        $token = $this->params['token'];
        $toUserId = $this->params['toUserId'];

        if (empty($userId)) {
            return $this->response_error(10000017, "user_id is empty");
        }

        if (empty($toUserId)) {
            return $this->response_error(10000059, "toUserId is empty");
        }

        if (empty($token)) {
            return $this->response_error(10000016, "token is empty");
        }

        $myUserInfo = $this->user_model->getUserinfoByUserid($userId);
        if (empty($myUserInfo)) {
            return $this->response_error(10000023, "未找到用户信息");
        }

        if (empty($token) || $token != $myUserInfo['token']) {
            return $this->response_error(10000024, "token 校验失败");
        }


        $userInfo = $this->user_model->getUserinfoByUserid($toUserId);
        if (empty($userInfo)) {
            return $this->response_error(10000023, "未找到该用户");
        }

        $data = array();
        $data['amount'] = intval($userInfo['amount']);
        $data['gender'] = intval($userInfo['gender']);
        $data['avatarUrl'] = $userInfo['avatar_url'];
        $data['nickname'] = $userInfo['nickname'];
        $data['form_id'] = $userInfo['form_id'];
        $data['userId'] = intval($userInfo['user_id']);

        return $this->response_success_to_me($data);
    }

    //解密群数据
    public function getOpenGidAction() {
    	//$this->util_log->LogInfo("f10001006Action:" . createLinkstringUrlencode($this->params));
        $userId = $this->params['userid'];
        $token = $this->params['token'];
        $encryptedData = $this->params['encryptedData'];
        $iv = $this->params['iv'];

        if (empty($userId)) {
            return $this->response_error(10000021, "userid is empty");
        }

        $userInfo = $this->user_model->getUserinfoByUserid($userId);
        if (empty($userInfo)) {
            return $this->response_error(10000023, "未找到该用户");
        }

        if (empty($token) || $token != $userInfo['token']) {
            return $this->response_error(10000024, "token 校验失败");
        }

        //解密敏感数据
        $session_key = $userInfo['session_key'];
        $appid = $userInfo['appid'];

        //$encryptedData = "ch7Ky7H8sZ0n/fk/MP9d2shyz3AVt+IwZ/L867gk8cs9w3QP693aFfn460mTx76Y/adKo/ZwHVayBDOR61SWTtPx0s9/VsESEDCqmMvKi3loY0nfVgLLPqwjzGZT3CSdeePqe7pV+KTl2uVYjYxL8w==";

        include_once APPPATH . "/application/library/smallgame/wxBizDataCrypt.php"; //引入微信api
        $pc = new WXBizDataCrypt($appid, $session_key);
        $errCode = $pc->decryptData($encryptedData, $iv, $data);

        if ($errCode == 0) {
            $data = json_decode($data, true);
            $result['openGId'] = $data['openGId'];
        } else {
            $this->util_log->LogError("f10001006Action decryptData error errCode=[$errCode]");
            return $this->response_error(10003212, "解密失败 - errCode=[$errCode]");
        }

        return $this->response_success_to_me($result);
    }

    //ip位置
    public function locationAction() {
    	//$this->util_log->LogInfo("f10001009Action:" . createLinkstringUrlencode($this->params));
        $appid = $this->params['appid'];
        $userid = $this->params['userid'];
        $version = $this->params['version'];

        $result = array();
        $result['flag'] = 0;
        $result['type'] = 'none';

        $config_file = Loader::config("filter");
        $config = $config_file[$appid][$version];

        if (empty($appid) || empty($config)) {
            $filter_area = $config_file['filter_area']; //默认排除几个地域
            if (!empty($filter_area)) {
                $location = Ip_Location::find($this->ip);
                $area = $location[1];
                if (in_array($area, $filter_area)) {
                    $result['flag'] = 1;
                    $result['type'] = 'area';
                    return $this->response_success_to_me($result);
                }
            }
            return $this->response_success_to_me($result);
        }

        //反向用户
        $fanxiang = false;
        $include_users = $config['include_users'];

        if (!empty($config['fanxiang'])) {
            $fanxiang = true;
        } else if (!empty($include_users) && !empty($userid)) {
            if (in_array($userid, $include_users)) {
                $fanxiang = true;
            }
        }

        //排除时间
        $filter_times = $config['filter_times'];
        if (!empty($filter_times)) {
            $time = date('H:i');
            foreach($filter_times as $filter_time) {
                if ($time >= $filter_time[0] && $time <= $filter_time[1]) {
                    if ($fanxiang) {
                        $result['flag'] = 0;
                    } else {
                        $result['flag'] = 1;
                        $result['type'] = 'time';
                    }
                    return $this->response_success_to_me($result);
                }
            }
        }

        //排除用户
        $filter_users = $config['filter_users'];
        if (!empty($filter_users) && !empty($userid)) {
            if (in_array($userid, $filter_users)) {
                if ($fanxiang) {
                    $result['flag'] = 0;
                } else {
                    $result['flag'] = 1;
                    $result['type'] = 'user';
                }
                return $this->response_success_to_me($result);
            }
        }

        //排除地域
        $filter_area = $config['filter_area'];
        if (!empty($filter_area)) {
            $location = Ip_Location::find($this->ip);
            $area = $location[1];
            if (in_array($area, $filter_area)) {
                if ($fanxiang) {
                    $result['flag'] = 0;
                } else {
                    $result['flag'] = 1;
                    $result['type'] = 'area';
                }
                return $this->response_success_to_me($result);
            }
        }

        if ($fanxiang) {
            $result['flag'] = 1;
            $result['type'] = 'fanxiang';
        } else {
            $result['flag'] = 0;
        }
        return $this->response_success_to_me($result);
    }

    //游戏配置
    public function gameConfAction() {
    	//$this->util_log->LogInfo("f10001008Action:" . createLinkstringUrlencode($this->params));
        $appid = $this->params['appid'];
        $userid = $this->params['userid'];
        $version = $this->params['version'];

        $config_file = Loader::config("gameconf");
        $config = $config_file[$appid][$version];

        return $this->response_success_to_me($config);
    }
    
    //获取用户合区信息
    public function getZoneUserAction() {
    	//$this->util_log->LogInfo("f10001010Action:" . createLinkstringUrlencode($this->params));
    	$userId = $this->params['userid'];
        $token = $this->params['token'];
        $zone_id = intval($this->params['zone_id']);

        if (empty($userId)) {
            return $this->response_error(10000021, "userid is empty");
        }

        $userInfo = $this->user_model->getUserinfoByUserid($userId);
        if (empty($userInfo)) {
            return $this->response_error(10000023, "未找到该用户");
        }

        if (empty($token) || $token != $userInfo['token']) {
            return $this->response_error(10000024, "token 校验失败");
        }
        
        $zone_user_id = $this->user_model->getUserZoneUid($userId, $zone_id);
        
        if ($zone_user_id == 0) {
        	//新用户，插入新区
        	$zone_user_id = $this->user_model->generate_userid(); 
        	if($zone_user_id == 0) {
        		return $this->response_error(99, "system error");
        	}
        	
        	$ret = $this->user_model->insertUserZoneUid($userId, $zone_id, $zone_user_id);
        	if($ret <= 0) {
            	return $this->response_error(99, "system error");
            }
        }
        
        $zoneinfo = Loader::config('zoneinfo');
        
        $result = array();
        $result['zone_user_id'] = $zone_user_id;
        $result['zone_id'] = $zone_id;
        $result['hequ'] = $zoneinfo['zone_info'][$zone_id - 1]['hequ'];
        return $this->response_success_to_me($result);
    }

    private function getAppConfig($appid) {
    	global $wxAppConfig;
        if (empty($appid)) {
           	return $this->response_error(10000001, 'appid is empty');
        }
		
        if (!isset($wxAppConfig[$appid])) {
            return $this->response_error(10000002, 'access token config not find');
        }

        $appConfigInfo = $wxAppConfig[$appid];

        if (empty($appConfigInfo['app_id']) || empty($appConfigInfo['secret'])) {
            return $this->response_error(10000003, 'access token config error');
        }

        return $appConfigInfo;
    }

    private function clearUserCache($user_id) {
        $redis = $this->loader->redis("userinfo");
        if (!empty($redis)) {
            $userInfo = $redis->del("pre_redis_user_info_" . $user_id);
        }
    }
    
    
    
}
