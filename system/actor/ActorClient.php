<?php
/**
 * Created by PhpStorm.
 * User: yf
 * Date: 2018-12-27
 * Time: 13:57
 */

use Swoole\Coroutine\Channel;

class ActorClient {
    private $tempDir;
    private $actorName;
    private $actorId;
    private $actorProcessNum;

    function __construct(ActorConfig $config, string $tempDir) {
        $this->tempDir = $tempDir;
        $this->actorName = $config->getActorName();
        $this->actorProcessNum = $config->getActorProcessNum();
    }

    function setActorId($id) {
        $this->actorId = $id;
        return $this;
    }

    function getActorId() {
        return $this->actorId;
    }

    function new($timeout, $arg) {
        $command = new Command();
        $command->setCommand('new');
        $command->setArg($arg);

        $i = rand(1, $this->actorProcessNum);

        $this->actorId = UnixClient::sendAndRecv($command, $timeout, $this->generateSocketByProcessIndex($i));
        return $this->actorId;
    }

    function exist(string $actorId, $timeout = 3.0) {
        $command = new Command();
        $command->setCommand('exist');
        $command->setArg($actorId);

        return UnixClient::sendAndRecv($command, $timeout, $this->generateSocketByProcessIndex(self::actorIdToProcessIndex($actorId)));
    }

    function destroy(...$arg) {
        $processIndex = self::actorIdToProcessIndex($this->actorId);
        $command = new Command();
        $command->setCommand('destroy');
        $command->setArg([
                             'id' => $this->actorId,
                             'arg' => $arg
                         ]);

        return UnixClient::sendAndRecv($command, 3.0, $this->generateSocketByProcessIndex($processIndex));
    }

    function destroyAll(...$arg) {
        $command = new Command();
        $command->setCommand('destroyAll');
        $command->setArg($arg);
        return $this->broadcast($command, 3.0);
    }

    function __call($func, $args) {
        $processIndex = self::actorIdToProcessIndex($this->actorId);
        $command = new Command();
        $command->setCommand('call');
        $command->setArg([
                             'id' => $this->actorId,
                             'func'=> $func,
                             'arg'=> $args
                         ]);

        return UnixClient::sendAndRecv($command, 3.0, $this->generateSocketByProcessIndex($processIndex));
    }

    private function broadcast(Command $command,$timeout = 3.0) {
        $info = [];
        $channel = new Channel($this->actorProcessNum+1);
        for ($i = 1; $i <= $this->actorProcessNum; $i++) {
            go(function ()use($command,$channel,$i,$timeout) {
                $ret = UnixClient::sendAndRecv($command,$timeout,$this->generateSocketByProcessIndex($i));
                $channel->push([
                                   $i => $ret
                               ]);
            }
              );
        }
        $start = microtime(true);
        while (1) {
            if (microtime(true) - $start > $timeout) {
                break;
            }
            $temp = $channel->pop($timeout);
            if (is_array($temp)) {
                $info += $temp;
                if (count($info) == $this->actorProcessNum) {
                    break;
                }
            }
        }

        return $info;
    }

    private function generateSocketByProcessIndex($processIndex):string {
        return $this->tempDir."/ActorProcess.".SERVER_NAME.".{$this->actorName}.{$processIndex}.sock";
    }

    public static function actorIdToProcessIndex(string $actorId):int {
        return intval(substr($actorId, 0, strpos($actorId, "0")));
    }
}
