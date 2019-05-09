# ycsocket
基于 swoole 和 swoole_orm 的 websocket 框架，各位可以自己扩展到 TCP/UDP，HTTP。

在ycsocket 中，采用的是全协程化，全池化的数据库、缓存IO，对于IO密集型型的应用，能够支撑较高并发。

环境：
PHP7+
swoole_orm   //一个C语言扩展的ORM，本框架协程数据库需要该扩展支持，https://github.com/swoole/ext-orm
swoole

我写推送后端的时候写的

客户端是一个聊天窗口

支持 Redis 协程线程池，源码位于 system/RedisPool，支持失败自动重连

支持 MySQL 协程连接池， 源码位于 system/MySQLPool，支持失败自动重连

支持共享内存 entity ,可以支持超时更新内容

加入Actor模型，基于unixsocket 和 channel 的高并发模型

# Actor 模型
   在高并发环境中，为了保证多个进程同时访问一个对象时的数据安全，我们经常采用互斥锁来实现，然而，互斥锁性能低下，用的不好，经常会造成死锁的情况发生，在这里，我们将采用 actor 模型来保证数据安全。
   
   基本原理是，将需要维护的对象注册到一个特定进程，所有对对象的数据修改，都必须通过一个信道(channel)来发送(push)给对象，由对象自己从信道读取(pop)修改请求，并创建一个协程来对自身进行维护，只要维护的函数中没有协程IO，则原则上每一次修改都是有序，并且唯一的。保证了数据安全。
   
   例如示例代码，是一个多人竞技的游戏服务器，代码中有3个Actor: RoomLogic、PkLogic、GameLogic，分别用于保存所有房间的容器、单个房间逻辑、以及房间内每个玩家的游戏逻辑，当然还有一个非Actor类AiLogic，用于处理AI玩家逻辑，代码存在于 application/logic 目录中。
   
## 简单从代码整理整个Actor模型思路：
#### Actor的注册：
```php
//Application.php
function register_actor() {
	Actor::getInstance()->register(RoomLogic::class, 1);
	Actor::getInstance()->register(PkLogic::class, 1);
	Actor::getInstance()->register(GameLogic::class, 1);
}
```
  每个Actor都是一个独立维护自身数据的个体，拥有一个唯一的id，他们本质上是一个类，继承自ActorBean，该父类拥有一些操作这些类对象的方法，比如创建Actor对象的new静态方法。创建后的Actor对象依附在一个特殊的进程 ActorProcess 中，业务进程可以通过 unixsocket 访问该对象。

  如果 PkLogic::new 将在 ActorProcess 进程创建一个ActorFactory对象，该对象拥有Actor的指针，例如RoomLogic，PkLogic，GameLogic，并创建一个信道channel，并监听信道，一旦有请求，将会开启一个协程处理消息，该消息必定会被顺序执行，但是切记在处理逻辑中不要出现阻塞方法，否则效率会非常低下，处理主要包含2种，一种是销毁Actor， 一种是调用实际的Actor方法，即RoomLogic，PkLogic，GameLogic的方法。
```php
class ActorFactory
{
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
                });
            }
        });
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
}
```

PkLogic::new 方法返回的并不是真正的Actor对象，而是一个ActorClient对象，我们可以通过该对象，来实现远程顺序调用Actor函数的目的，当然，这里的远程是指的跨进程，从业务进程到ActorProcess。
```php
class RoomLogic extends ActorBean {
    private $joiningRoom;

    public function joinRoom($userid, ... ) {
        $this->joiningRoom['pkLogic'] = PkLogic::new();
        $this->joiningRoom['pkLogic']->joinUser($userid);
        $this->joiningRoom['id'] = $this->joiningRoom['pkLogic']->getActorId();
    }
}
```
上面创建通过 PkLogic::new 创建Actor对象后，调用joinUser方法，由于 PkLogic::new() 返回的是 ActorClient 对象，然后ActorClient并没有 joinUser 方法，那么他会调用 ActorClient 的魔术方法，该魔术方法会将请求通过unixsocket传到 ActorProcess 进程，并push到ActorFactory的信道，然后由ActorFactory从信道获取数据，并实现真正的函数调用，并返回结果。
```php
class ActorClient
{
    private $tempDir;
    private $actorName;
    private $actorId;
    private $actorProcessNum;

    function __construct(ActorConfig $config, string $tempDir) {
        $this->tempDir = $tempDir;
        $this->actorName = $config->getActorName();
        $this->actorProcessNum = $config->getActorProcessNum();
    }

    function new($timeout, $arg);

    function exist(string $actorId, $timeout = 3.0);

    function destroy(...$arg);

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

    private function generateSocketByProcessIndex($processIndex):string {
        return $this->tempDir."/ActorProcess.".SERVER_NAME.".{$this->actorName}.{$processIndex}.sock";
    }

    public static function actorIdToProcessIndex(string $actorId):int {
        return intval(substr($actorId, 0, strpos($actorId, "0")));
    }
}
```
   
   
   
   
   
   
   
   
   
