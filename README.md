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

加入Actor模型，基于协程版unixsocket 和 协程channel 的高并发事务模型。

本框架 Actor 模型借鉴自EasySwoole 框架的Actor模块。 github网址： https://github.com/easy-swoole/actor ，做了些修改后融入本框架

# Actor 模型
   在高并发环境中，为了保证多个进程同时访问一个对象时的数据安全，我们通常采用两种策略，共享数据和消息传递，
   
   使用共享数据方式的并发编程面临的最大的一个问题就是数据条件竞争（data race）。处理各种锁的问题是让人十分头痛的一件事，锁限制了并发性, 调用者线程阻塞带来的浪费，用的不好，还容易造成死锁。
   
   和共享数据方式相比，消息传递机制最大的优点就是不会产生数据竞争状态（data race），除此之外，还有如下一些优点：
- 事件模型驱动--Actor之间的通信是异步的，即使Actor在发送消息后也无需阻塞或者等待就能够处理其他事情
- 强隔离性--Actor中的方法不能由外部直接调用，所有的一切都通过消息传递进行的，从而避免了Actor之间的数据共享，想要观察到另一个Actor的状态变化只能通过消息传递进行询问
- 位置透明--无论Actor地址是在本地还是在远程机上对于代码来说都是一样的
- 轻量性--Actor是非常轻量的计算单机，单个Actor仅包含一个 actorId 和 channel 对象，只需少量内存就能达到高并发
   
   本代码Actor模型主要基于swoole协程的channel来实现，进程间通过协程版 unix domain socket 进行通信，当然Actor不仅仅局限于单个节点上，也可以作为分布式集群运行。
   
## 基本原理
   Actor模型=数据+行为+消息
   
   Actor模型内部的状态由它自己维护即它内部数据只能由它自己修改(通过消息传递来进行状态修改)，Actor由状态(state)、行为(Behavior)和邮箱(mailBox)三部分组成

- 状态(state)：Actor中的状态指的是Actor对象的变量信息，状态由Actor自己管理，避免了并发环境下的锁和内存原子性等问题
- 行为(Behavior)：行为指定的是Actor中计算逻辑，通过Actor接收到消息来改变Actor的状态
- 邮箱(mailBox)：邮箱是Actor和Actor之间的通信桥梁，邮箱内部通过FIFO消息队列来存储发送方Actor消息，接受方Actor从邮箱队列中获取消息

   Actor的基础就是消息传递
   
## 示例代码剖析
   我们框架中的示例代码，是一个多人竞技的游戏服务器，代码中有3个 Actor : RoomLogic 、 PkLogic 、 GameLogic，分别用于存储所有房间的游戏大厅、单个房间逻辑、房间内每个玩家的游戏逻辑，还有一个非 Actor 类 AiLogic ，用于处理AI玩家逻辑，代码都存在于 application/logic 目录中。

### Actor的注册：
```php
//Application.php
function register_actor() {
	Actor::getInstance()->register(RoomLogic::class, 1);
	Actor::getInstance()->register(PkLogic::class, 1);
	Actor::getInstance()->register(GameLogic::class, 1);
}
```
   每一个Actor对于其他的Actor来说都是封闭的,他们拥有自己的变量，依附于一个特殊的进程ActorProcess，不同进程的Actor通过协程版unix domain socket 通讯，访问请求会被写入Actor信箱(channel)，也就是说Actor本身是进程安全的。
   
   每个Actor拥有一个唯一的id，所有进程对Actor的访问，都是通过该id来确定Actor的进程位置，然后发送消息来访问Actor，所有Actor在使用之前都需要注册，注册主要是初始化Actor名称、进程数、启动回调函数、销毁回调函数、定时任务等信息。

   在注册完成之后，我们将为每个Actor都创建对应的依附进程。并将进程挂到 swoole 服务器下。
   
```php
$ws = new swoole_server("0.0.0.0", 9508,  SWOOLE_PROCESS, SWOOLE_SOCK_TCP | SWOOLE_SSL);
Actor::getInstance()->attachToServer($ws);
```
   
### Actor的创建与信箱的监听
   Actor本质上是一个类，所有Actor都继承自ActorBean，该父类保存每个Actor的唯一编号actorId，和一些操作这些Actor对象的方法，比如创建Actor对象的new静态方法。
   
```php
//ActorBean.php
class ActorBean {
    protected $actorId;
    
    public static function new(...$args);
    public static function getBean(string $actorId);
    public function exist();
    public function bean();
    function onDestroy(...$arg);
    function getThis();
    function setActorId($actorId);
    function getActorId();
}
```
```php
//RoomLogic.php 单例
class RoomLogic extends ActorBean {
    private static $instance;

    public function __construct() {
    }

    public static function getInstance() {
        if (!isset(self::$instance)) {
            global $roomActorId; //通过一个全局变量在共享内存中存储 RoomLogic 的 ActorId
            $actorIdArray = $roomActorId->get("RoomActorId");
            if (empty($actorIdArray['id'])) {
                self::$instance = RoomLogic::new();
                $roomActorId->set("RoomActorId", ['id' => self::$instance->getActorId()]);
            } else {
                self::$instance = RoomLogic::getBean($actorIdArray['id']);
            }
        }

        return self::$instance;
    }
    ...
}
```
   PkLogic::new方法会通过协程版 unix domain socket 发送新建请求到ActorProcess，该进程会通过工厂类(ActorFactory)创建真实的Actor对象，即RoomLogic、 PkLogic、 GameLogic对象，同时在 startUp 函数里，创建一个信箱(channel)，并创建一个协程来监听信箱，一旦有请求，将会开启一个协程处理消息，该消息必定会被顺序依次处理，但是切记在处理逻辑中不要出现阻塞IO，请使用swoole协程版IO，包含数据库、redis等，否则效率会非常低下，
   如果消息是销毁Actor，工厂会删除真实的Actor对象，并在ActorProcess进程里销毁该工厂。还有就是通过 call_user_func 调用真实Actor对象的成员函数。
   
```php
class ActorFactory
{
    public function __construct($actorClass, $actorId, $args) {
        $this->realActor = new $actorClass(...$args);
        $this->realActor->setActorId($actorId);
        $this->actorId = $actorId;
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

### Actor 行为
PkLogic::new 方法返回的并不是真实的Actor对象，而是一个ActorClient，我们可以通过ActorClient来实现远程顺序调用真实Actor成员函数的目的，当然，这里的远程是指的跨进程，从业务进程到ActorProcess，如果扩展到分布式集群环境下，这里可以是集群中节点。

```php
class RoomLogic extends ActorBean {
    private $joiningRoom;

    public function joinRoom($userid, ... ) {
        $this->joiningRoom['pkLogic'] = PkLogic::new();
        $this->joiningRoom['pkLogic']->joinUser($userid);
        $this->joiningRoom['id'] = $this->joiningRoom['pkLogic']->getActorId();
    }
}

class PkLogic extends ActorBean {
    private $gameLogics = array();

    public function __construct() {
    }

    public function joinUser($uid) {
        $this->gameLogics[$uid] = GameLogic::new($this->actorId, $uid);
        $this->gameLogics[$uid]->createGame();

        return count($this->gameLogics);
    }
}
```

上面创建通过 PkLogic::new 创建Actor对象后，调用joinUser方法，由于 PkLogic::new() 返回的是 ActorClient 对象，然后ActorClient并没有 joinUser 方法，那么他会调用 \_\_call 魔术方法，该魔术方法会将请求通过 unixsocket 传到 ActorProcess 进程，并在该进程被 push 到ActorFactory 的信箱(channel)，ActorFactory 的监听协程会从信箱 pop 数据，并实现真正的函数调用，并返回结果。
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
   
### Actor的销毁
ActorClient有个 destroy 方法，用于销毁Actor。
```php
class ActorClient {
    private $actorId;
    
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
}

class ActorFactory {
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
}
```
Actor的销毁也是将消息destroy消息发送到ActorProcess，然后由工厂类做一些清理工作，最后删除真实的Actor对象，在delele之前，会调用真实Actor的onDestroy方法，这个函数在父类ActorBean是一个空函数，用户可以重写该函数以便加入自己的清理逻辑，例如下面的PkLogic在onDestroy方法里面，将销毁GameLogic对象来清理房间内所有玩家的游戏数据。
```php
class RoomLogic extends ActorBean {
    private $playingRooms;
    
    public function exitRoom($pkid) {
        $this->playingRooms[$pkid]['pkLogic']->destroy();
        unset($this->playingRooms[$pkid]);
    }
}


class PkLogic extends ActorBean {
    private $gameLogics = array();

    function onDestroy() {
        foreach($this->gameLogics as $gameLogics) {
            $gameLogics->destroy();
        }
    }
}
```

   
   
