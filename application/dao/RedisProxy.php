<?php

class RedisProxy
{
    //新手礼包领取时间
    public static function get_new_gift_time($user_id)
    {
        $redis_key = "new_gift_time_" . $user_id;
        $time = intval(RedisPool::instance('pika')->get($redis_key));
        if ($time == 0) {
            $time = time() + 180;
            RedisPool::instance('pika')->set($redis_key, $time);
            RedisPool::instance('pika')->expire($redis_key, 86400);

        }

        return time() >= $time ? 0 : $time - time();
    }

    //今天是否已经领取，新手登录礼包
    public static function get_today_new_7day_gift($user_id)
    {
        $redis_key = "new_7day_gift_" . $user_id . "_" . date('Ymd');
        return intval(RedisPool::instance('pika')->get($redis_key));
    }


    //获取帮战状态数据
    public static function get_guild_fight_status()
    {
        return intval(RedisPool::instance('pika')->get("guild_fight_status"));
    }

    //设置帮战状态
    public static function set_guild_fight_status($status)
    {
        RedisPool::instance('pika')->set("guild_fight_status", intval($status));
    }

    //获取下周1的日期
    public static function get_next_monday()
    {
        $time = time();

        $w = date('w', $time);

        if ($w == 0) {
            $nextWeekTime = $time + 86400;
        } else {
            $nextWeekTime = $time + (8 - $w) * 86400;
        }

        return date('Ymd', $nextWeekTime);
    }

    //离下周一还有多久
    public static function left_time_to_next_monday()
    {
        return strtotime(self::get_next_monday()) - time();
    }

    //离明天还有多久
    public static function left_time_to_tomorrow()
    {
        return strtotime(date('Ymd', (time() + 86400))) - time();
    }
}
