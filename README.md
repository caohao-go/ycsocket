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

   在高并发环境中，为了保证多个进程同时访问一个对象时的数据安全，我们经常采用互斥锁来实现，然而，互斥锁性能低下，用的不好，经常会造成死锁的情况发生，在这里，我们将采用 actor 模型来保证数据安全，
   基本原理是，将需要维护的对象注册到一个特定进程，所有对对象的数据修改，都必须通过一个信道(channel)来发送(push)给对象，由对象自己从信道获取(pop)修改请求，并创建一个协程来对自身进行维护，只要维护的函数中没有协程IO，则原则上每一次修改都是有序，并且唯一的。保证了数据安全。
