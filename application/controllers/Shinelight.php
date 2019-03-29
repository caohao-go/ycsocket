<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');
/**
 * ShinelightController Class
 *
 * @package			Ycsocket
 * @subpackage		Controller
 * @category		ShinelightController
 * @author			caohao
 */
class ShinelightController extends SuperController {
    public function init() {
        $this->userinfo_model = $this->loader->model('UserinfoModel');
        $this->item_model = $this->loader->model('ItemModel');
        $this->util_log = $this->loader->logger('shinelight_log');
    }

    //公告
    public function gonggaoAction() {
        $content = $this->params['content'];

        if (empty($content)) {
            return $this->response_error(13342339, '内容不能为空');
        }

        $result = array();
        $result['content'] =  $content;
		
        return $this->response_success_to_all($result);
    }

    //聊天接口
    public function chatAction() {
        $userId = $this->params['userid'];
        $nickname = $this->params['nickname'];
        $avatar_url = $this->params['avatar_url'];
        $content = $this->params['content'];
        $gender = $this->params['gender'];
		

        if (empty($content)) {
            return $this->response_error(13342339, '内容不能为空');
        }

        $result = array();
        $result['userid'] = $userId;
        $result['nickname'] = $nickname;
        $result['avatar_url'] = $avatar_url;
        $result['gender'] = $gender;
        $result['content'] = $content;

        return $this->response_success_to_all($result);
    }
    
    //返回用户道具接口
    public function userItemAction() {
        $userId = $this->params['userid'];
        $token = $this->params['token'];
    	
        
        $data = $this->item_model->get_user_items($userId);
        return $this->response_success_to_all(['list' => $data]);
    }
    
    //测试
    public function testAction() {
        $userId = $this->params['userid'];
    	
        
        $ret = $this->item_model->insert_user_items($userId, 1, 80, 1);
        if(!$ret) {
        	return $this->response_error(99, 'system error');
        }
        
        $data = $this->item_model->get_user_items($userId);
        
        return $this->response_success_to_all(['list' => $data]);
    }
}
