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
    在高并发环境中，为了保证多个进程同时访问一个对象时的数据安全，我们通常采用两种策略，共享数据和消息传递，
   
    使用共享数据方式的并发编程面临的最大的一个问题就是数据条件竞争（data race）。处理各种锁的问题是让人十分头痛的一件事，锁限制了并发性, 调用者线程阻塞带来的浪费，用的不好，还会造成死锁
   
    和共享数据方式相比，消息传递机制最大的优点就是不会产生数据竞争状态（data race）。实现消息传递有两种常见的类型：基于channel的消息传递和基于Actor的消息传递。
   
    本代码Actor模型主要基于swoole协程的channel来实现，进程间通过协程版 unix domain socket 进行通信
   
## 基本原理
    Actor模型=数据+行为+消息
   
    Actor模型内部的状态由它自己维护即它内部数据只能由它自己修改(通过消息传递来进行状态修改)，所以使用Actors模型进行并发编程可以很好地避免这些问题，Actor由状态(state)、行为(Behavior)和邮箱(mailBox)三部分组成

- 状态(state)：Actor中的状态指的是Actor对象的变量信息，状态由Actor自己管理，避免了并发环境下的锁和内存原子性等问题
- 行为(Behavior)：行为指定的是Actor中计算逻辑，通过Actor接收到消息来改变Actor的状态
- 邮箱(mailBox)：邮箱是Actor和Actor之间的通信桥梁，邮箱内部通过FIFO消息队列来存储发送方Actor消息，接受方Actor从邮箱队列中获取消息

Actor的基础就是消息传递
   
## 代码剖析
    我们框架中的示例代码，是一个多人竞技的游戏服务器，代码中有3个Actor: RoomLogic、PkLogic、GameLogic，分别用于存储所有房间、单个房间逻辑、房间内每个玩家的游戏逻辑，当然还有一个非Actor类AiLogic，用于处理AI玩家逻辑，代码存在于 application/logic 目录中。

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
   
   
   
   
   
   
   
   
   
