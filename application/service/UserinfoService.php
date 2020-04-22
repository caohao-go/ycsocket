<?php if (!defined('BASEPATH')) exit('No direct script access allowed');

/**
 * UserinfoService    Class
 *
 * @package           Ycsocket
 * @subpackage        Service
 * @category          UserinfoService
 * @author            caohao
 */
class UserinfoService extends SuperService
{

    public function init()
    {
        $this->user_dao = $this->loader->dao("UserDao");
    }

    public function getZoneUserAndAuth($zone_user_id, $token)
    {
        if (empty($zone_user_id)) {
            throw new LogicException("user id is empty", 99900031);
        }

        $userInfo = $this->user_dao->getUserinfoByUserid(self::getUidByZoneUserId($zone_user_id));
        if (empty($userInfo)) {
            throw new LogicException("not find user", 99900032);
        }

        if (empty($token) || $token != $userInfo['token']) {
            throw new LogicException("token is invalid", 99900033);
        }

        return $userInfo;
    }

    public function getUserAndAuth($user_id, $token)
    {
        if (empty($user_id)) {
            throw new LogicException("user id is empty", 99900031);
        }

        $userInfo = $this->user_dao->getUserinfoByUserid($user_id);
        if (empty($userInfo)) {
            throw new LogicException("not find user", 99900032);
        }

        if (empty($token) || $token != $userInfo['token']) {
            throw new LogicException("token is invalid", 99900033);
        }
        return $userInfo;
    }

    public function getUserInZoneUserids($zone_userids)
    {
        $ret = array();

        if (empty($zone_userids)) {
            return $ret;
        }

        $user_ids = array();
        foreach ($zone_userids as $v) {
            $user_ids[] = self::getUidByZoneUserId($v);
        }

        $userinfos = $this->user_dao->getUserinfos($user_ids);
        if (empty($userinfos)) {
            return array();
        }

        $userinfos_array = array();
        foreach ($userinfos as $value) {
            $userinfos_array[$value['user_id']] = $value;
        }

        foreach ($zone_userids as $zone_user_id) {
            $user_id = self::getUidByZoneUserId($zone_user_id);
            $userinfo = $userinfos_array[$user_id];
            $value['user_id'] = $zone_user_id;
            $value['nickname'] = $userinfo['nickname'];
            $value['avatar_url'] = $userinfo['avatar_url'];
            $ret[$zone_user_id] = $value;
        }

        return $ret;
    }

    public static function getUserZoneUid($userid, $zone_id)
    {
        $zone_id = $zone_id <= 0 ? 1 : $zone_id;
        return sprintf("%s%03d", $userid, intval($zone_id));
    }

    public static function getUidByZoneUserId($zone_user_id)
    {
        return intval($zone_user_id / 1000);
    }

    public static function getUserZoneid($zone_user_id)
    {
        return intval($zone_user_id) % 1000;
    }

}
