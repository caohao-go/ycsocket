<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');
/**
 * ShinelightController Class  https://github.com/caohao-php/ycsocket
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
                
        $joined = RoomLogic::getInstance()->isInRoom($userId);
        
        if(!empty($joined)) {
        	return $this->response_success_to_all(['data' => $joined]);
    	}
        
        $ret = RoomLogic::getInstance()->joinRoom($userId);
        
        return $this->response_success_to_all(['data' => $ret]);
    }

    public function cmdAction() {
        $userId = intval($this->params['userid']);
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
