<?php if (!defined('BASEPATH')) exit('No direct script access allowed');

/**
 * UserDao    Class
 *
 * @package            Ycsocket
 * @subpackage        Dao
 * @category        UserDao
 * @author            caohao
 */
class UserDao extends SuperDao
{
    public function init()
    {
        $this->util_log = $this->loader->logger('game_log');
    }

    //user_info 表
    public function getUserinfoByUserid($user_id)
    {
        $redis_key = "pre_redis_user_info_" . $user_id;
        $user_info = $this->get_one_table_data("user_info", ['user_id' => $user_id], $redis_key);
        return $user_info;
    }

    public function getUserinfos($user_ids)
    {
        return $this->get_table_data("user_info", ['user_id' => $user_ids], "", "user_id,nickname,avatar_url");
    }

    //user_login_zones 表
    public static function getLoginZone($userid)
    {
        $sql = "select * from user_login_zones where user_id={$userid} order by utime desc";
        return MySQLPool::instance('default')->query($sql);
    }
}
