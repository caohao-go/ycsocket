<?php
/**
 * Created by PhpStorm.
 * User: yf
 * Date: 2018-12-27
 * Time: 13:38
 */
 
namespace EasySwoole\Actor;
 
class Command {
    protected $command;
    protected $arg;

    /**
     * @return mixed
     */
    public function getCommand() {
        return $this->command;
    }

    /**
     * @param mixed $command
     */
    public function setCommand($command): void {
        $this->command = $command;
    }

    /**
     * @return mixed
     */
    public function getArg() {
        return $this->arg;
    }

    /**
     * @param mixed $arg
     */
    public function setArg($arg): void {
        $this->arg = $arg;
    }
}
