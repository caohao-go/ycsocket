<?php
/**
 * Created by PhpStorm.
 * User: yf
 * Date: 2018-12-27
 * Time: 12:13
 */

use Swoole\Coroutine\Channel;

class ActorFactory
{
    protected $hasDoExit = false;
    protected $actorId;
    private $channel;
    private $tickList = [];
    private $replyChannel;
    protected $block = false;
    private $realActor;
    
    public function __construct($actorClass, $args)
    {
    	$this->realActor = new $actorClass(...$args);
    }
    
    function setBlock(bool $bool)
    {
        $this->block = $bool;
        return $this;
    }
    
   	public static function pack($data) {
   		return base64_encode(swoole_serialize::pack($data));
   	}
   	
   	public static function unpack($data) {
   		return swoole_serialize::unpack(base64_decode($data));
   	}
   	
	public static function configure(ActorConfig $actorConfig, $actorName)
    {
        $actorConfig->setActorName($actorName);
        $actorConfig->setOnStart(function (ActorProcess $actorProcess){
            $data = $actorProcess->status();
            $file = BASEPATH . "/actor/data/actorData.{$data['processIndex']}.data";
            if(file_exists($file)){
                $data = ActorFactory::unpack(file_get_contents($file));
                $actorProcess->setStatus($data);
                foreach ($data['actorList'] as $key => $actor){
                    $actorProcess->wakeUpActor($actor);
                }
            }
        });

        $saveFunc = function (ActorProcess $actorProcess){
            $data = $actorProcess->status();
            $str = ActorFactory::pack($data);
        	file_put_contents(BASEPATH . "/actor/data/actorData.{$data['processIndex']}.data",$str);
        };
        
        //每5s落地一次
        $actorConfig->setTick(5*1000);
        $actorConfig->setOnTick($saveFunc);
        //on shutdown 仅在正常关闭的情况下执行。可以改为定时器执行。
        $actorConfig->setOnShutdown($saveFunc);
    }
    
    function actorId()
    {
        return $this->actorId;
    }

    function setActorId($id)
    {
        $this->actorId = $id;
        return $this;
    }

    /*
     * 请用该方法来添加定时器，方便退出的时候自动清理定时器
     */
    function tick($time, callable $callback)
    {
        $id = swoole_timer_tick($time, function () use ($callback) {
            try {
                call_user_func($callback);
            } catch (\Throwable $throwable) {
                $this->onException($throwable);
            }
        });
        $this->tickList[$id] = $id;
        return $id;
    }

    /*
     * 请用该方法来添加定时器，方便退出的时候自动清理定时器
     */
    function after($time, callable $callback)
    {
        $id = swoole_timer_after($time, function () use ($callback) {
            try {
                call_user_func($callback);
            } catch (\Throwable $throwable) {
                $this->onException($throwable);
            }
        });
        return $id;
    }

    function deleteTick(int $timerId)
    {
        unset($this->tickList[$timerId]);
        return swoole_timer_clear($timerId);
    }

    function getChannel(): ?Channel
    {
        return $this->channel;
    }

    function __startUp(Channel $replyChannel)
    {
        $this->channel = new Channel(64);
        $this->replyChannel = $replyChannel;
        if ($this->block) {
            go(function () {
                $this->listen();
            });
        } else {
            $this->listen();
        }
    }
    
    private function exitHandler($arg)
    {
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
            $reply = $this->onException($throwable);
        }
        
        return $reply;
    }
    
    public function wakeUp(Channel $replyChannel)
    {
        $this->replyChannel = $replyChannel;
        $this->channel = new Channel(64);
        $this->listen();
    }

    private function listen()
    {
        go(function () {
            while (!$this->hasDoExit) {
                $array = $this->channel->pop();
                if (is_array($array)) {
                    if($this->block){
                        $this->handlerMsg($array);
                    }else{
                        go(function ()use($array){
                            $this->handlerMsg($array);
                        });
                    }
                }
            }
        });
    }

    private function handlerMsg(array $array)
    {
    	echo "=========handlerMsg==========\n";
    	var_dump($array);
        $msg = $array['msg'];
        if ($msg == 'destroy') {
            $reply = $this->exitHandler($array['arg']);
        } else {
            try {
                $reply = call_user_func([$this->realActor, $array['func']], ...$array['arg']);
            } catch (\Throwable $throwable) {
                $reply = $this->onException($throwable);
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
            $this->replyChannel->push($conn);
        }
    }
    
    protected function onException(\Throwable $throwable) {
		echo "ActorFactory::onException - Catch An Exception \n";
		echo "File:" . $throwable->getFile() . "\n";
		echo "Line:" . $throwable->getLine() . "\n";
		echo "Code:" . $throwable->getCode() . "\n";
		echo "Message:" . $throwable->getMessage() . "\n";
    }
}
