<?php if (!defined('BASEPATH')) exit('No direct script access allowed');

/** 控制器基类
 * SuperController Class
 *
 * @package            Ycsocket
 * @subpackage        Controller
 * @category        Controller Base
 * @author            caohao
 */
class SuperController
{
    protected $ip;
    protected $params;
    protected $loader;

    public function __construct(& $params = array(), $clientInfo = array())
    {
        $this->ip = $clientInfo['remote_ip'];
        $this->params = & $params;

        $this->loader = new Loader($this);
        $this->init();
    }

    protected function init()
    {
    }

    public function & get_ip()
    {
        return $this->ip;
    }

    public function & get_params()
    {
        return $this->params;
    }

    private function get_result_success($message)
    {
        if (empty($message)) {
            $message = array('c' => $this->params['c'], 'm' => $this->params['m'], 'reqid' => $this->params['reqid'], 'code' => 0);
        } else {
            $message['c'] = $this->params['c'];
            $message['m'] = $this->params['m'];
            $message['reqid'] = $this->params['reqid'];
            $message['code'] = 0;
        }

        return json_encode($message);
    }

    private function get_result_error($code, $message)
    {
        $data = array('c' => $this->params['c'], 'm' => $this->params['m'], 'reqid' => $this->params['reqid'], "code" => $code, "msg" => $message);
        return json_encode($data);
    }

    /**
     * json输出
     * @param array $data
     */
    protected function response_success_to_all($message)
    {
        $data = array();
        $data["send_user"] = "all";
        $data["msg"] = $this->get_result_success($message);
        return $data;
    }

    /**
     * json输出
     * @param array $data
     */
    protected function response_success_to_me($message)
    {
        $data = array();
        $data["send_user"] = "me";
        $data["msg"] = $this->get_result_success($message);
        return $data;
    }

    /**
     * json输出
     * @param array $data
     */
    protected function response_success_to_uids($uids, $message)
    {
        $data = array();
        $data["send_user"] = array_values($uids);
        $data["msg"] = $this->get_result_success($message);
        return $data;
    }

    /**
     * 返回错误code以及错误信息
     * @param sting $message 返回错误的提示信息
     * @param int $type 返回的方式
     */
    protected function response_error($code, $message)
    {
        $data = array();
        $data["send_user"] = "me";
        $data["msg"] = $this->get_result_error($code, $message);
        return $data;
    }
}
