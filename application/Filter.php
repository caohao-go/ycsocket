<?php if (!defined('BASEPATH')) exit('No direct script access allowed');

/**
 * Filter    Class
 *
 * @package            Ycsocket
 * @subpackage        Filter
 * @category        Filter
 * @author            caohao
 */
class Filter
{
    //验签过程
    public static function auth(& $params)
    {
        /*
        if($auth_error == false) { //验签失败
            return self::response_error(123, "auth error");
        } */

        //验签成功
        return 0;
    }

    public static function response_error($code, $message)
    {
        $data = array("code" => $code, "msg" => $message);
        $result['send_user'] = "me";
        $result['msg'] = json_encode($data);
        return $result;
    }
}
