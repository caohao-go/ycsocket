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
        $appid = $this->params['appid'];
        $openid = $this->params['open_id'];

        $result['openid'] = $openid;
        $userInfo = $this->user_model->get_one_user_info_by_key('open_id', $openid);
        if (empty($userInfo)) {
            $result['userid'] = $this->user_model->generate_userid();
            $result['token'] = $this->user_model->register_user($appid, $result['userid'], $result['openid']);
        } else {
            $this->clearUserCache($userInfo['user_id']); //清理缓存
            $result['userid'] = $userInfo['user_id'];
            $result['token'] = $this->user_model->login_user($result['userid']);
        }

        if (empty($result['token'])) {
            return $this->response_error(10000009, "登陆失败, 请重试");
        }
        
        return $this->response_success_to_me($result);
    }

    //修改用户信息
    public function modifyInfoAction() {
        $this->util_log->LogInfo("f10001001Action:" . createLinkstringUrlencode($this->params));
        $userId = $this->params['userid'];
        $rawData = $this->params['rawData'];

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

        $this->user_model->update_user($userId, $update_user_info);
        return $this->response_success_to_me(array());
    }
	
    //获取用户信息接口
    public function getInfoAction() {
        $userId = $this->params['userid'];

        if (empty($userId)) {
            return $this->response_error(10000017, "user_id is empty");
        }

        $userInfo = $this->user_model->getUserinfoByUserid($userId);
        if (empty($userInfo)) {
            return $this->response_error(10000023, "未找到该用户");
        }

        $data = array();
        $data['amount'] = $userInfo['amount'];
        $data['gender'] = intval($userInfo['gender']);
        $data['avatarUrl'] = $userInfo['avatar_url'];
        $data['nickname'] = $userInfo['nickname'];
        $data['form_id'] = $userInfo['form_id'];
        
        return $this->response_success_to_me($data);
    }

    //ip位置
    public function locationAction() {
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
        $appid = $this->params['appid'];
        $version = $this->params['version'];

        $config_file = Loader::config("gameconf");
        $config = $config_file[$appid][$version];

        return $this->response_success_to_me($config);
    }

    private function clearUserCache($user_id) {
    	MySQLPool::instance("default")->del("pre_redis_user_info_" . $user_id);
    }
    
}
