<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');
/**
 * ExampleModel Class
 *
 * @package        SuperCI
 * @subpackage    Model
 * @category      Example Model
 * @author        caohao
 */
class DabaojianModel extends Core_Model
{
    public function init(){
        $this->db = $this->loader->database('starfast');
        $this->util_log = $this->loader->logger('dabaojian_log');
    }
}