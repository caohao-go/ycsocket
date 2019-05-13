<?php
/**
 * Created by PhpStorm.
 * User: yf
 * Date: 2018-12-27
 * Time: 12:10
 */

namespace EasySwoole\Actor;

class ActorBean {
    protected $actorId;

    public function __construct() {
    }

    public static function new(...$args) {
        return Actor::getInstance()->actorCreate(static::class, $args);
    }

    public function exist() {
        return Actor::getInstance()->actorExist(static::class, $this->actorId, 3.0);
    }

    public static function getBean(string $actorId) {
        return Actor::getInstance()->getActorById(static::class, $actorId);
    }

    public function bean() {
        return Actor::getInstance()->getActorById(static::class, $this->actorId);
    }

    function onDestroy(...$arg) {
    }

    function getThis() {
        return $this;
    }

    function setActorId($actorId) {
        $this->actorId = $actorId;
    }

    function getActorId() {
        return $this->actorId;
    }
}