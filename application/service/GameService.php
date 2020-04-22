<?php if (!defined('BASEPATH')) exit('No direct script access allowed');

/**
 * GameService Class
 *
 * @package            Ycsocket
 * @subpackage         Service
 * @category           Example Service
 * @author             caohao
 */
class GameService extends SuperService
{
    public function init()
    {
        parent::init();
        $this->game_dao = $this->loader->dao("GameDao");
        $this->userinfo_service = $this->loader->dao("UserinfoService");
        $this->util_log = $this->loader->logger('game_log');
    }

    //用户充值
    public function get_user_vip_contents($userid)
    {
        $data = $this->game_dao->get_user_vip_contents($userid);

        if (empty($data['content'])) {
            $content = array();
            $content['leiji_xiaofei'] = 0;  //累计消费
            $content['leiji_chong'] = 0;  //累计充值
            $content['jijin']['status'] = 0;  //是否购买成长基金 0-未购买 1-已购买
            $this->game_dao->insert_user_vip_contents($userid, $content);
        } else {
            $content = json_decode($data['content'], true);
        }

        return $content;
    }

    //更新充值信息
    public function update_user_vip_contents($userid, $content)
    {
        return $this->game_dao->update_user_vip_contents($userid, $content);
    }

    //玩家英雄
    public function get_user_heros($userid)
    {
        return $this->game_dao->get_user_heros($userid);
    }

    //删除玩家英雄
    public function delete_user_heros($userid, $id)
    {
        return $this->game_dao->delete_user_heros($userid, $id);
    }

    //修改玩家英雄数据
    public function replace_user_heros($userid, $id)
    {
        $replace_data['id'] = $id;
        $replace_data['hero_id'] = '1336';
        $replace_data['hero'] = '雅典娜';
        $replace_data['lv'] = 1;
        return $this->game_dao->replace_user_heros($userid, $replace_data);
    }

    /////////////////// 排行榜  ///////////////////////
    //分数更新，修改排名
    public function modify_rank($rank_name, $userid, $score, $expire = 0)
    {
        RedisPool::instance('pika')->zadd("pre_{$rank_name}_rank", $score, $userid);
        $this->game_dao->clear_redis_cache("pre_{$rank_name}_rank_cache");

        if (!empty($expire)) {
            RedisPool::instance('pika')->expire("pre_{$rank_name}_rank", $expire);
        }

        return true;
    }

    //分数新增
    public function incr_rank_score($rank_name, $userid, $add_score, $expire = 0)
    {
        RedisPool::instance('pika')->zincrby("pre_{$rank_name}_rank", $add_score, $userid);
        $this->game_dao->clear_redis_cache("pre_{$rank_name}_rank_cache");

        if (!empty($expire)) {
            RedisPool::instance('pika')->expire("pre_{$rank_name}_rank", $expire);
        }

        return true;
    }

    //获取我的排名
    public function get_my_rank($rank_name, $userid)
    {
        $myRank = RedisPool::instance('pika')->zrevrank("pre_{$rank_name}_rank", $userid);
        return $myRank === false || $myRank === NULL ? 0 : $myRank + 1;
    }

    //获取我的分数
    public function get_my_rank_score($rank_name, $userid)
    {
        $score = RedisPool::instance('pika')->zscore("pre_{$rank_name}_rank", $userid);
        return intval($score);
    }

    //清理排名
    public function clear_rank($rank_name)
    {
        RedisPool::instance('pika')->del("pre_{$rank_name}_rank");
        $this->game_dao->clear_redis_cache("pre_{$rank_name}_rank_cache");
    }

    //清除我的排行
    public function clear_my_rank($rank_name, $userid)
    {
        RedisPool::instance('pika')->zrem("pre_{$rank_name}_rank", $userid);
        $this->game_dao->clear_redis_cache("pre_{$rank_name}_rank_cache");
    }

    //获取排名列表
    public function get_rank_list($rank_name, $return_userinfo_flag = true, $start = 0, $end = 99)
    {
        $pre_rank_cache = "pre_{$rank_name}_rank_cache";
        $result = $this->game_dao->hget_redis($pre_rank_cache, "{$start}_{$end}");
        if (!empty($result) && $result == SuperDao::EMPTY_STRING) {
            return array();
        }


        if (!empty($result)) {
            $result = unserialize($result);
        } else {
            $result = array();
            $score_ranks = RedisPool::instance('pika')->zrevrange("pre_{$rank_name}_rank", $start, $end, 1);
            if (!empty($score_ranks)) {
                $user_keys = array_keys($score_ranks);

                $userinfos = array();
                if ($return_userinfo_flag) {
                    $userinfos = $this->userinfo_service->getUserInZoneUserids($user_keys);
                }

                foreach ($score_ranks as $key => $value) {
                    if (!empty($userinfos[$key])) {
                        $tmp = $userinfos[$key];
                        $tmp['score'] = $value;
                    } else {
                        $tmp = array('user_id' => $key, 'zone_id' => 0, 'nickname' => '', 'avatar_url' => '');
                        $tmp['score'] = $value;
                    }

                    $result[] = $tmp;
                }
            }

            $this->game_dao->hset_redis($pre_rank_cache, "{$start}_{$end}", $result, 60);
        }

        return $result;
    }
}
