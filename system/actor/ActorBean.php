<?php
/**
 * Created by PhpStorm.
 * User: yf
 * Date: 2018-12-27
 * Time: 12:10
 */

class ActorBean
{
	public static function new(...$args)
	{
		return Actor::getInstance()->actorCreate(static::class, $args);
	}
	
	public static function exist(string $actorId, $timeout = 3.0) {
		return Actor::getInstance()->actorExist(static::class, $actorId, $timeout);
	}
	
    function onDestroy(...$arg)
    {
    }
}