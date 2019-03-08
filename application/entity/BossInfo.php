<?php

class BossInfo extends Entity
{
	var $status = 0; //Boss状态 0-未开始 1-已开始
	var $blood = 1000; //血量
	var $start_time = 0; //开始时间
	var $is_test = 0; //test
	var $last_cut_time = 0; //上次砍BOSS时间
	
	const start_time_1 = 8;
	const start_time_2 = 13;
	
	public function init() {
		$this->status = 0;
		$this->blood = 1000;
		$this->start_time = date('H');
		
	}
	
	public function get_boss_info(& $dabaojian_model) {
		$this->getGlobal(GLOBAL_WORLD_BOSS);
		
		if($this->start_time == 0) { // 初始化boss
			if($this->is_test == 0) { //测试初始化为 18:03:00
				$this->start_time = "2018-09-12 18:03:00";
				$this->is_test = 1;
			} else {
				$this->next_start_time();
			}
		} else {
			if($this->status == 0) {
				if(time() > strtotime($this->start_time)) { //当前时间大于开始时间，激活boss
					echo "clear dabaojian_current_world_boss\n";
					$dabaojian_model->clear_rank("dabaojian_current_world_boss");
					$this->status = 1;
					$this->blood = 1000;
				}
			}
		}
		
		//echo "get_boss_info\n";
		//print_r($this);
		
		$this->setGlobal(GLOBAL_WORLD_BOSS);
		return $this->getArray();
	}
	
	public function next_start_time() {
		$hour = date('H');
		if($hour > self::start_time_1 && $hour <= self::start_time_2) { //8点-13点之间，就13点开始
			$this->start_time = date('Y-m-d '.self::start_time_2.':00:00');
		} else if(date('H') <= self::start_time_1){ //小于8点，就当天8点开始，
			$this->start_time = date('Y-m-d '.self::start_time_1.':00:00');
		} else { //次日8点开始
			$this->start_time = date('Y-m-d '.self::start_time_1.':00:00', time() + 86400);
		}
	}
	
	public function cut_boss($userid, $cut_blood, & $dabaojian_model,  & $left_blood = array()) {
		$left_blood = $this->get_boss_info($dabaojian_model);
		
		if($this->blood <= 0) {
			return "BOSS已经死亡";
		}
		
		if($this->status == 0) {
			return "BOSS还没开始";
		}
		
		if($this->blood == 1000) {
			$this->last_cut_time = $this->get_current_microtime();
			$del_blood = rand(2, 10);
		} else {
			$time_between = $this->get_current_microtime() - $this->last_cut_time;
			$this->last_cut_time = $this->get_current_microtime();
			$del_blood = intval(rand(2, 10) * $time_between);
		}
		
		$is_last_cut = 0;
		if($this->blood <= $del_blood) {
			$this->next_start_time();
			//$this->start_time = date('Y-m-d H:i:s', time() + 30);  //测试
			$this->status = 0;
			$this->blood = 0;
			$is_last_cut = 1;
			$redis = Loader::redis("default");
			$redis->set("dabaojian_world_boss_last_cut", $userid); //最后一击的人
		} else {
			$this->blood = $this->blood - $del_blood;
		}
		
		
		$this->setGlobal(GLOBAL_WORLD_BOSS);
		
		$dabaojian_model->add_rank_score("dabaojian_current_world_boss", $userid, $cut_blood);
		$left_blood = $this->get_boss_info($dabaojian_model);
		$left_blood['is_last_cut'] = $is_last_cut;
		unset($left_blood['last_cut_time']);
		echo "cut_boss : $del_blood\n";
		//print_r($this);
		return 0;
	}
	
	public function get_current_microtime() {
		list($msec, $sec) = explode(' ', microtime());
		return $sec . substr($msec, 1, 3); 
	}
}
