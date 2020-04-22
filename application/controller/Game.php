<?php if (!defined('BASEPATH')) exit('No direct script access allowed');

/**
 * GameController Class
 *
 * @package            Ycsocket
 * @subpackage        Controller
 * @category        GameController
 * @author            caohao
 */
class GameController extends SuperController
{
    var $game_service;
    var $userinfo_service;

    public function init()
    {
        $this->userinfo_service = $this->loader->service('UserinfoService');
        $this->game_service = $this->loader->service('GameService');

        $this->util_log = $this->loader->logger('game_log');
    }

    //版本
    public function verAction()
    {
        $userId = $this->params['userid'];
        return $this->response_success_to_me(['ver' => Zoneinfo::$game_version]);
    }

    //聊天接口
    public function chatAction()
    {
        $userId = $this->params['userid'];
        $type = intval($this->params['type']);  //0-世界 1-私聊
        $token = $this->params['token'];
        $nickname = $this->params['nickname'];
        $avatar_url = $this->params['avatar_url'];
        $content = $this->params['content'];
        $to_userid = intval($this->params['to_userid']);

        $this->userinfo_service->getZoneUserAndAuth($userId, $token);

        if (empty($content)) {
            return $this->response_error(13342339, '内容不能为空');
        }

        $result = array();
        $result['userid'] = $userId;
        $result['type'] = $type;
        $result['nickname'] = $nickname;
        $result['avatar_url'] = $avatar_url;
        $result['gender'] = $this->params['gender'];
        $result['vip_level'] = $this->params['vip_level'];
        $result['lv'] = $this->params['lv'];
        $result['content'] = $content;

        if ($type == 0) {
            return $this->response_success_to_all($result);
        } else if ($type == 1) {
            return $this->response_success_to_uids([$userId, $to_userid], $result);
        }
    }

    //用户信息
    public function userGameInfoAction()
    {
        $userId = $this->params['userid'];
        $token = $this->params['token'];

        $this->userinfo_service->getZoneUserAndAuth($userId, $token);

        $data['vipinfo'] = $this->game_service->get_user_vip_contents($userId);  //玩家会员信息
        $data['heros'] = $this->game_service->get_user_heros($userId); //玩家英雄
        $data['time'] = RedisProxy::get_new_gift_time($userId);  //新礼物等待时间
        $data['today_lingqu'] = RedisProxy::get_today_new_7day_gift($userId);  //新玩家7天登陆礼

        return $this->response_success_to_me($data);
    }

    //获取排行榜
    public function rankListAction()
    {
        $userId = $this->params['userid'];
        $token = $this->params['token'];
        $type = $this->params['type'];  //copy-剧情进度 guild-工会排行 pk-竞技场 fight-战力排行

        $this->userinfo_service->getZoneUserAndAuth($userId, $token);

        $data['my_rank'] = $this->game_service->get_my_rank($type, $userId);
        $data['my_score'] = $this->game_service->get_my_rank_score($type, $userId);
        $data['rank_list'] = $this->game_service->get_rank_list($type, true, 0, 99);

        return $this->response_success_to_me($data);
    }

    //返回区信息
    public function zoneInfoAction()
    {
        $userId = $this->params['userid']; //用户id，非 zone_uid
        $token = $this->params['token'];

        $userInfo = $this->userinfo_service->getUserAndAuth($userId, $token);

        $config = Zoneinfo::zone_info();
        $result = array();
        foreach ($config as $zone_info) {
            $zone_id = $zone_info['zone_id'];
            $result["A" . $zone_id][] = $zone_info;
        }

        krsort($result);

        //login zone
        $login_zone = UserDao::getLoginZone($userId);
        foreach ($login_zone as $k => $v) {
            $login_zone[$k]['level'] = $v['lv'];
            $login_zone[$k]['nickname'] = $userInfo['nickname'];
        }

        return $this->response_success_to_me(array('nickname' => $userInfo['nickname'],
            'zone_list' => $result,
            'login_zone' => $login_zone,
            'recommend_zone' => Zoneinfo::recommend_zone()));
    }
}
