<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');

/** 模型基类
 * SuperModel Class  https://github.com/caohao-php/ycsocket
 *
 * @package			Ycsocket
 * @subpackage		Controller
 * @category		Controller Base
 * @author			caohao
 */

class SuperLibrary {
    protected $loader;

    public function __construct(& $loader) {
        $this->loader = & $loader;
        $this->init();
    }

    protected function init() {
    }
}
