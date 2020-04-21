<?php if (!defined('BASEPATH')) exit('No direct script access allowed');

/** 业务层基类
 * SuperService Class
 *
 * @package            Ycsocket
 * @subpackage        Service
 * @category        Service Base
 * @author            caohao
 */
class SuperService
{
    protected $loader;

    public function __construct(& $loader = null)
    {
        if (!empty($loader)) {
            $this->loader = &$loader;
        } else {
            $this->loader = new Loader();
        }

        $this->init();
    }

    protected function init()
    {
    }
}
