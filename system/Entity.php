<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');

class Entity {
    public function getGlobal($id){
    	$data = GlobalEntity::get($id);
    	
    	if(!empty($data)) {
    		foreach($data as $key => $value) {
    			$this->$key = $value;
    		}
    	}
    	
    	return $data;
    }
    
    public function setGlobal($id) {
    	return GlobalEntity::set($id, $this);
    }
    
	public function getJson() {
		return json_encode($this, JSON_FORCE_OBJECT);
	}
	
	public function getArray() {
		$data = json_encode($this, JSON_FORCE_OBJECT);
		return json_decode($data, true);
	}
}
