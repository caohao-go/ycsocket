<?php

class Zoneinfo
{
    static public $zoneinfo;
    static public $game_version;

    public static function update()
    {
        self::$zoneinfo = require(APP_ROOT . "/config/zoneinfo.php");
    }

    public static function zone_info()
    {
        return self::$zoneinfo['zone_info'];
    }

    public static function recommend_zone()
    {
        return self::$zoneinfo['recommend_zone'];
    }

    public static function game_version()
    {
        $game_version = MySQLPool::instance("default")->query("select * from game_version");
        self::$game_version = $game_version[0]['ver'];
    }
}
 