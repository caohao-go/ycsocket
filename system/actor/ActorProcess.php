<?php
/**
 * Created by PhpStorm.
 * User: yf
 * Date: 2018-12-27
 * Time: 13:13
 */

namespace EasySwoole\Actor;

use Swoole\Coroutine\Channel;
use Swoole\Coroutine\Socket;

class ActorProcess extends AbstractProcess {
    protected $actorIndex = 2;//index是为了做actorId前缀标记
    protected $actorAtomic = 0;
    protected $processIndex;
    protected $actorClass;
    protected $actorList = [];

    /**
     * @var $config ProcessConfig
     */
    protected $config;

    public function run($processConfig) {
        $this->config = $processConfig;

        \Swoole\Runtime::enableCoroutine(true);
        \co::set(['max_coroutine' => 200000]);

        $this->processIndex = $processConfig->getIndex();
        $this->actorClass = $processConfig->getActorClass();

        if ($processConfig->getTick() > 0 && is_callable($processConfig->getOnTick())) {
            $onTick = $processConfig->getOnTick();
            $this->addTick($processConfig->getTick(), function () use ($onTick) {
                try {
                    call_user_func($onTick, $this);
                } catch (\Throwable $throwable) {
                    $this->onException($throwable);
                }
            }
                          );
        }

        $sockFile = $processConfig->getTempDir() . "/{$this->getProcessName()}.sock";
        if (file_exists($sockFile)) {
            unlink($sockFile);
        }

        $socketServer = new Socket(AF_UNIX,SOCK_STREAM,0);
        $socketServer->bind($sockFile);
        if (!$socketServer->listen($processConfig->getBacklog())) {
            trigger_error('listen '.$sockFile. ' fail');
            return;
        }

        go(function () use ($processConfig, $socketServer) {
            if (is_callable($this->config->getOnStart())) {
                try {
                    call_user_func($this->config->getOnStart(), $this);
                } catch (\Throwable $throwable) {
                    $this->onException($throwable);
                }
            }

            while (1) {
                $conn = $socketServer->accept(-1);
                if ($conn) {
                    go(function () use ($conn) {
                        //先取4个字节的头
                        $header = $conn->recv(4,1);
                        if (strlen($header) != 4) {
                            $conn->close();
                            return;
                        }

                        $allLength = Protocol::packDataLength($header);
                        $recvLeft = $allLength;
                        $data = '';
                        $tryTimes = 10;

                        while ($recvLeft > 0 && $tryTimes > 0) {
                            $temp = $conn->recv($allLength, 1);
                            if ($temp === false) {
                                break;
                            }
                            $data = $data.$temp;
                            $recvLeft = $recvLeft - strlen($temp);
                            $tryTimes--;
                        }

                        if (strlen($data) != $allLength) {
                            $conn->close();
                            return;
                        }

                        $fromPackage = ActorFactory::unpack($data);
                        if (!$fromPackage instanceof Command) {
                            $conn->close();
                            return;
                        }

                        switch ($fromPackage->getCommand()) {
                        case 'new': {
                            $actorId = $this->processIndex.str_pad($this->actorIndex, 11, '0', STR_PAD_LEFT);
                            $this->actorIndex++;
                            $this->actorAtomic++;
                            try {
                                /** @var $actor ActorFactory*/
                                $actor = new ActorFactory($this->actorClass, $actorId, $fromPackage->getArg());
                                $this->actorList[$actorId] = $actor;
                                $actor->__startUp();
                            } catch (\Throwable $throwable) {
                                $this->actorAtomic--;
                                unset($this->actorList[$actorId]);
                                $actorId = null;
                                $this->onException($throwable);
                            }
                            $conn->send(Protocol::pack(ActorFactory::pack($actorId)));
                            $conn->close();
                            break;
                        }

                        case 'exist': {
                            $actorId = $fromPackage->getArg();
                            if (isset($this->actorList[$actorId])) {
                                $conn->send(Protocol::pack(ActorFactory::pack(true)));
                            } else {
                                $conn->send(Protocol::pack(ActorFactory::pack(false)));
                            }
                            $conn->close();
                            break;
                        }

                        case 'call': {
                            $args = $fromPackage->getArg();

                            if (isset($args['id'])) {
                                $actorId = $args['id'];
                                if (isset($this->actorList[$actorId])) {
                                    $pushData = ['connection'=> $conn, 'msg' => 'call', 'func' => $args['func'], 'arg' => $args['arg'], 'reply'=> true];

                                    $this->actorList[$actorId]->getChannel()->push($pushData);
                                    break;
                                }
                            }
                            $conn->send(Protocol::pack(ActorFactory::pack(null)));
                            $conn->close();
                            break;
                        }

                        case 'destroy': {
                            $args = $fromPackage->getArg();
                            if (isset($args['id'])) {
                                $actorId = $args['id'];
                                if (isset($this->actorList[$actorId])) {
                                    //消息回复在actor中
                                    $this->actorList[$actorId]->getChannel()->push([
                                                'connection'=> $conn,
                                                'msg'=> 'destroy',
                                                'arg'=> $args['arg'],
                                                'reply'=> true
                                            ]);
                                    $this->actorAtomic--;
                                    unset($this->actorList[$actorId]);
                                    break;
                                }
                            }
                            $conn->send(Protocol::pack(ActorFactory::pack(null)));
                            $conn->close();
                            break;
                        }

                        case 'destroyAll': {
                            $this->actorAtomic = 0;
                            $args = $fromPackage->getArg();
                            foreach ($this->actorList as $actorId => $item) {
                                //单独多出arg字段
                                $item->getChannel()->push(['msg'=>'destroy', 'reply'=> false, 'arg'=> $args]);
                                unset($this->actorList[$actorId]);
                            }

                            $conn->send(Protocol::pack(ActorFactory::pack(true)));
                            $conn->close();
                            break;
                        }

                        default: {
                            $conn->send(Protocol::pack(ActorFactory::pack(null)));
                            $conn->close();
                            break;
                        }
                        }
                    }
                      );
                }
            }
        }
          );
    }

    public function onShutDown() {
        if (is_callable($this->config->getOnShutdown())) {
            try {
                call_user_func($this->config->getOnShutdown(), $this);
            } catch (\Throwable $throwable) {
                $this->onException($throwable);
            }
        }
    }

    public function onReceive(string $str) {
    }

    protected function onException(\Throwable $throwable) {
        if (is_callable($this->config->getProcessOnException())) {
            call_user_func($this->config->getProcessOnException(), $throwable);
        } else {
            parent::onException($throwable);
        }
    }

    public function status() {
        return [
                   'actorIndex'=>$this->actorIndex,
                   'actorAtomic'=>$this->actorAtomic,
                   'actorList'=>$this->actorList,
                   'processIndex'=>$this->processIndex
               ];
    }

    public function setStatus($status) {
        if (isset($status['actorIndex'])) {
            $this->actorIndex = $status['actorIndex'];
        }
        if (isset($status['actorAtomic'])) {
            $this->actorAtomic = $status['actorAtomic'];
        }
        if (isset($status['actorList'])) {
            $this->actorList = $status['actorList'];
        }
    }

    public function wakeUpActor(ActorFactory $actor) {
        $this->actorList[$actor->actorId()] = $actor;
        $actor->wakeUp();
    }

    /**
     * @return ProcessConfig
     */
    public function getConfig(): ProcessConfig {
        return $this->config;
    }

}
