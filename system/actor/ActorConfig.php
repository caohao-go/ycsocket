<?php
/**
 * Created by PhpStorm.
 * User: yf
 * Date: 2018-12-27
 * Time: 12:14
 */
 
namespace EasySwoole\Actor;

class ActorConfig {
    protected $actorName;
    protected $actorProcessNum = 3;
    protected $maxActorNum = 10000;
    protected $backlog = 256;
    /**
     * @var $onStart callable
     */
    protected $onStart;
    /**
     * @var $onShutdown callable
     */
    protected $onShutdown;

    protected $tick = 0;

    protected $onTick;

    protected $processOnException;

    /**
     * @return mixed
     */
    public function getActorName() {
        return $this->actorName;
    }

    public function setActorName($actorName): ActorConfig {
        $this->actorName = $actorName;
        return $this;
    }

    /**
     * @return int
     */
    public function getActorProcessNum(): int {
        return $this->actorProcessNum;
    }


    public function setActorProcessNum(int $actorProcessNum): ActorConfig {
        $this->actorProcessNum = $actorProcessNum;
        return $this;
    }

    /**
     * @return int
     */
    public function getMaxActorNum(): int {
        return $this->maxActorNum;
    }


    public function setMaxActorNum(int $maxActorNum): ActorConfig {
        $this->maxActorNum = $maxActorNum;
        return $this;
    }

    /**
     * @return int
     */
    public function getBacklog(): int {
        return $this->backlog;
    }

    /**
     * @param int $backlog
     */
    public function setBacklog(int $backlog): void {
        $this->backlog = $backlog;
    }

    /**
     * @return callable
     */
    public function getOnStart(): ?callable {
        return $this->onStart;
    }

    /**
     * @param callable $onStart
     */
    public function setOnStart(?callable $onStart): void {
        $this->onStart = $onStart;
    }

    /**
     * @return callable
     */
    public function getOnShutdown(): ?callable {
        return $this->onShutdown;
    }

    /**
     * @param callable $onShutdown
     */
    public function setOnShutdown(?callable $onShutdown): void {
        $this->onShutdown = $onShutdown;
    }

    /**
     * @return int
     */
    public function getTick(): int {
        return $this->tick;
    }

    /**
     * @param int $tick
     */
    public function setTick(int $tick): void {
        $this->tick = $tick;
    }

    /**
     * @return mixed
     */
    public function getOnTick() {
        return $this->onTick;
    }

    /**
     * @param mixed $onTick
     */
    public function setOnTick($onTick): void {
        $this->onTick = $onTick;
    }

    /**
     * @return mixed
     */
    public function getProcessOnException() {
        return $this->processOnException;
    }

    /**
     * @param mixed $processOnException
     */
    public function setProcessOnException($processOnException): void {
        $this->processOnException = $processOnException;
    }
}
