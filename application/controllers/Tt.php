<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');
/**
 * ShinelightController Class
 *
 * @package			Ycsocket
 * @subpackage		Controller
 * @category		ShinelightController
 * @author			caohao
 */
class TtController extends SuperController {
    public function init() {
    }
	
    public function joinAction() {
        $userId = intval($this->params['userid']);
        $token = $this->params['token'];
                
        $joined = RoomLogic::getInstance()->isUserJoined($userId);
        
        if(!empty($joined)) {
        	return $this->response_success_to_all(['data' => $joined]);
    	}
        
        $joined = RoomLogic::getInstance()->joinRoom($userId, 'nickname', 'avatar');
        $result = array();
        $result['id'] = $joined['id'];
        $result['state'] = GameLogic::STATE_JOIN;
        $result['createTime'] = $joined['createTime'];
        $result['users'] = array_values($joined['users']);
        
        return $this->response_success_to_all(['data' => $result]);
    }
    
    public function cmdAction() {
        $userId = intval($this->params['userid']);
        $token = $this->params['token'];
        $pkid = intval($this->params['pkid']);
        $cmd = $this->params['cmd'];
        
        $pkLogic = PkLogic::getBean($pkid);
        $gameLogic = $pkLogic->getGameLogicByUid($userId);
        if(empty($gameLogic)) {
        	return $this->response_error(99, 'system error');
        }
        
        $ret = $gameLogic->action($cmd);
        if($ret === false) {
        	return $this->response_error(99, 'system error');
        }
        
        return $this->response_success_to_all(['data' => $ret]);
    }
}
