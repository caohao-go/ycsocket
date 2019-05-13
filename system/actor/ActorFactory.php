<?php
/**
 * Created by PhpStorm.
 * User: yf
 * Date: 2018-12-27
 * Time: 12:13
 */

namespace EasySwoole\Actor;

use Swoole\Coroutine\Channel;

class ActorFactory {
    private $hasDoExit = false;
    private $actorId;
    private $channel;
    private $tickList = [];
    private $realActor;

    public static function configure(ActorConfig $actorConfig, $actorName) {
        $actorConfig->setActorName($actorName);
        $actorConfig->setOnStart(function (ActorProcess $actorProcess) {
            /*
            $data = $actorProcess->status();
            $file = BASEPATH . "/actor/data/actorData.{$data['processIndex']}.data";
            if(file_exists($file)) {
                $data = ActorFactory::unpack(file_get_contents($file));
                $actorProcess->setStatus($data);
                foreach ($data['actorList'] as $key => $actor){
                    $actorProcess->wakeUpActor($actor);
                }
            }*/
        });

        $saveFunc = function (ActorProcess $actorProcess) {
            /*
            $data = $actorProcess->status();
            $str = ActorFactory::pack($data);
            file_put_contents(BASEPATH . "/actor/data/actorData.{$data['processIndex']}.data",$str);
            */
        };

        //每5s落地一次
        $actorConfig->setTick(5*1000);
        $actorConfig->setOnTick($saveFunc);
        //on shutdown 仅在正常关闭的情况下执行。可以改为定时器执行。
        $actorConfig->setOnShutdown($saveFunc);
    }

    public function __construct($actorClass, $actorId, $args) {
        $this->realActor = new $actorClass(...$args);
        $this->realActor->setActorId($actorId);
        $this->actorId = $actorId;
    }

    public static function pack($data) {
        return base64_encode(serialize($data));
    }

    public static function unpack($data) {
        return unserialize(base64_decode($data));
    }

    function actorId() {
        return $this->actorId;
    }

    function setActorId($id) {
        $this->actorId = $id;
        return $this;
    }

    /*
     * 请用该方法来添加定时器，方便退出的时候自动清理定时器
     */
    function tick($time, callable $callback) {
        $id = swoole_timer_tick($time, function () use ($callback) {
            try {
                call_user_func($callback);
            } catch (\Throwable $throwable) {
                $this->onException($throwable);
            }
        }
                               );
        $this->tickList[$id] = $id;
        return $id;
    }

    /*
     * 请用该方法来添加定时器，方便退出的时候自动清理定时器
     */
    function after($time, callable $callback) {
        $id = swoole_timer_after($time, function () use ($callback) {
            try {
                call_user_func($callback);
            } catch (\Throwable $throwable) {
                $this->onException($throwable);
            }
        }
                                );
        return $id;
    }

    function deleteTick(int $timerId) {
        unset($this->tickList[$timerId]);
        return swoole_timer_clear($timerId);
    }

    function getChannel(): ?Channel {
        return $this->channel;
    }

    public function wakeUp() {
        $this->channel = new Channel(64);
        $this->listen();
    }

    function __startUp() {
        $this->channel = new Channel(64);
        $this->listen();
    }

    private function listen() {
        go(function () {
            while (!$this->hasDoExit) {
                $array = $this->channel->pop();

                go(function ()use($array) {
                    $this->handlerMsg($array);
                }
                  );
            }
        }
          );
    }

    private function handlerMsg(array $array) {
        $msg = $array['msg'];
        if ($msg == 'destroy') {
            $reply = $this->exitHandler($array['arg']);
        } else {
            try {
                $reply = call_user_func([$this->realActor, $array['func']], ...$array['arg']);
            } catch (\Throwable $throwable) {
                $reply = $throwable;
            }
        }

        if ($array['reply']) {
            $conn = $array['connection'];
            $string = Protocol::pack(ActorFactory::pack($reply));
            for ($written = 0; $written < strlen($string); $written += $fwrite) {
                $fwrite = $conn->send(substr($string, $written));
                if ($fwrite === false) {
                    break;
                }
            }
            $conn->close();
        }
    }

    private function exitHandler($arg) {
        $reply = null;

        try {
            //清理定时器
            foreach ($this->tickList as $tickId) {
                swoole_timer_clear($tickId);
            }

            $this->hasDoExit = true;
            $this->channel->close();

            $reply = $this->realActor->onDestroy(...$arg);
            if ($reply === null) {
                $reply = true;
            }
        } catch (\Throwable $throwable) {
            $reply = $throwable;
        }

        return $reply;
    }

    protected function onException(\Throwable $throwable) {
        echo "ActorFactory::onException \n";
        echo "File:" . $throwable->getFile() . "\n";
        echo "Line:" . $throwable->getLine() . "\n";
        echo "Code:" . $throwable->getCode() . "\n";
        echo "Message:" . $throwable->getMessage() . "\n";
        return $throwable;
    }
}
