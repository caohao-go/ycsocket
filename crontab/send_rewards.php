<?php
date_default_timezone_set('Asia/Shanghai');
header('Content-Type: text/html; charset=UTF-8');
ini_set('display_errors', 'On');
error_reporting(E_ERROR); /*E_ALL,E_ERROR*/

define("ENABLE_COROUTINE", true);

define("APPPATH", realpath(dirname(__FILE__) . '/../'));
define("BASEPATH", APPPATH . '/system');
define("APP_ROOT", APPPATH . '/application');

Swoole\Runtime::enableCoroutine(SWOOLE_HOOK_ALL | SWOOLE_HOOK_CURL);

$input_zone_id = intval($argv[1]);
if ($input_zone_id < 1) {
    die("请输入区号\n");
}

//根据游戏不同区设置端口
define('GAME_ZONE_ID', $input_zone_id);  //游戏区

include(BASEPATH . "/Application.php");

Co\run(function () {
    Zoneinfo::update();

    //奖品
    $rewards = array();
    $data = MySQLPool::instance("game")->query("select * from copy_reward where copy_id in (10001, 20100)");

    foreach ($data as $v) {
        $copy_id = $v['copy_id'];
        $rewards_tmp = array();
        $reward_array = json_decode($v['reward'], true);
        foreach ($reward_array as $reward) {
            $rewards_tmp[] = ['type' => $reward[0], 'num' => $reward[1]];
        }

        if (is_numeric($v['rank_count'])) {
            $rewards[$copy_id][$v['rank_count']] = $rewards_tmp;
        } else {
            $rank_count = json_decode($v['rank_count'], true);
            for ($i = $rank_count[0]; $i <= $rank_count[1]; $i++) {
                $rewards[$copy_id][$i] = $rewards_tmp;
            }
        }
    }

    //竞技场排行日常奖励
    go(function () use ($rewards) {
        $score_ranks = RedisPool::instance("pika")->zrevrange('pk_rank_keys', 0, 1999, 1);

        $rank = 1;
        foreach ($score_ranks as $userid => $score) {
            //GameService::insert_user_mail($userid, '竞技场日常奖励', '恭喜您在竞技场日常排行中取得第' . $rank . '名', $rewards[10001][$rank]);
            $rank++;
        }
    });

    //无尽试炼排行榜
    go(function () use ($rewards) {
        $score_ranks = RedisPool::instance("pika")->zrevrange('pre_endless_layer_rank_rank', 0, 19, 1);
        $rank = 1;
        foreach ($score_ranks as $userid => $score) {
            //GameService::insert_user_mail($userid, '无尽试炼排行奖励', '恭喜您在无尽试炼中取得第' . $rank . '名', $rewards[20100][$rank]);
            $rank++;
        }
    });

});
