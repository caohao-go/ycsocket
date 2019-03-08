<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');

/** 模型基类
 * SuperModel Class
 *
 * @package			Ycsocket
 * @subpackage		Controller
 * @category		Controller Base
 * @author			caohao
 */

class SuperModel {
    protected $loader;

    public function __construct(& $loader) {
        $this->loader = & $loader;
        $this->init();
    }

    protected function init() {
    }
}
