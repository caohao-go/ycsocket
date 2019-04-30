<?php
class Userfd {
	use Singleton;
	
    private $ws;

    private function __construct($ws) {
        $this->ws = $ws;
    }
    
    public function set($uid, $fd) {
    	global $fdUserTable;
        return $fdUserTable->set("u_" . $uid, array("fd"=> $fd));
    }

    public function get($uid) {
    	global $fdUserTable;
        $fd = $fdUserTable->get("u_" . $uid);
        return intval($fd['fd']);
    }

    public function del($uid) {
    	global $fdUserTable;
		$fdUserTable->del("u_" . $uid);
    }
    
    public function send($uids, $msg) {
    	go(function() use ($uids, $msg) {
    		if(is_array($msg)) {
    			$msg['code'] = 0;
    			$msg = json_encode($msg);
    		}
    		
	    	if(is_array($uids)) {
	    		foreach($uids as $uid) {
	    			$this->_send($uid, $msg);
	    		}
	    	} else {
	    		$this->_send($uids, $msg);
	    	}
	    });
    }
    
    private function _send($uid, $msg) {
    	$fd = $this->get($uid);
    	if($fd != 0) {
    		$this->ws->push($fd, $msg);
    	}
    }
}
