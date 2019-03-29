<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');

class Entity {
    protected $loader;

    public function __construct(& $loader) {
        $this->loader = & $loader;
        $this->init();
    }

    protected function init() {
    }

    public function getGlobal($id) {
        $data = GlobalEntity::get($id);
        return $data;
    }

    public function setGlobal($id, $data, $expire = 0) {
        return GlobalEntity::set($id, $data, $expire);
    }
}
