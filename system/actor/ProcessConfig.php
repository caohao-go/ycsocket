<?php
/**
 * Created by PhpStorm.
 * User: yf
 * Date: 2018-12-27
 * Time: 13:14
 */

namespace EasySwoole\Actor;

class ProcessConfig extends ActorConfig {
    protected $index;
    protected $tempDir;
    protected $actorClass;
    protected $processName;

    /**
     * @return mixed
     */
    public function getIndex() {
        return $this->index;
    }

    /**
     * @param mixed $index
     */
    public function setIndex($index): void {
        $this->index = $index;
    }

    /**
     * @return mixed
     */
    public function getTempDir() {
        return $this->tempDir;
    }

    /**
     * @param mixed $tempDir
     */
    public function setTempDir($tempDir): void {
        $this->tempDir = $tempDir;
    }

    /**
     * @return mixed
     */
    public function getActorClass() {
        return $this->actorClass;
    }

    /**
     * @param mixed $actorClass
     */
    public function setActorClass($actorClass): void {
        $this->actorClass = $actorClass;
    }

    /**
     * @return mixed
     */
    public function getProcessName() {
        return $this->processName;
    }

    /**
     * @param mixed $processName
     */
    public function setProcessName($processName): void {
        $this->processName = $processName;
    }
}
