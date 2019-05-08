<?php
/**
 * Created by PhpStorm.
 * User: yf
 * Date: 2018-12-27
 * Time: 12:10
 */

class ActorBean {
    protected $actorId;

    public function __construct() {
    }

    public static function new(...$args) {
        return Actor::getInstance()->actorCreate(static::class, $args);
    }

    public static function exist(string $actorId, $timeout = 3.0) {
        return Actor::getInstance()->actorExist(static::class, $actorId, $timeout);
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