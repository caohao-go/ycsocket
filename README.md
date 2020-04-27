# ycsocket 概述
基于 swoole 和 swoole_orm 的极度轻量级 websocket 框架，我认为php就应该简单高效，各位可以自己扩展到 TCP/UDP，HTTP。

在ycsocket 中，采用的是全协程化，全池化的数据库、缓存IO，支持重连，对于IO密集型的应用，能够支撑较高并发。 <br><br>
支持 Redis 协程线程池，源码位于 system/RedisPool，支持失败自动重连<br>
支持 MySQL 协程连接池， 源码位于 system/MySQLPool，支持失败自动重连<br>

客户端chat.html是一个聊天窗口，用于发送测试 demo

# 环境
PHP7.1+ <br>
swoole 4.0 以上 <br>
swoole_orm   //一个C语言扩展的ORM，本框架协程数据库需要该扩展支持，https://github.com/swoole/ext-orm  <br>


# 应用场景
大型RPG游戏，纯php解决方案，php + swoole + swoole_orm + zephir ，这个游戏的战斗部分完全用 zephir 来实现，转化为 php 扩展，能做到同时兼顾性能和开发效率，（zephir 代码有机会我再开源出来，目前时机不成熟，游戏还比较火热），微信小游戏搜索："剑的传说" <br><br>
![Image](https://github.com/caohao-php/ycsocket/blob/master/image/1.jpeg)



# 代码结构
```php
———————————————— 
|--- server.php               //启动入口 
|--- system                   //框架系统代码
|--- application              //业务代码 
         |----- config        //配置目录
         |----- controller    //控制器目录
                |------ Game.php    //Game控制器
         |----- dao           //数据层
         |----- library       //公用类库
         |----- service       //业务层
```

# 请求路由
```
webSocket.send('{"c":"game","m":"ver", "userid":123593}');
```
输入参数为json， 根据 c 和 m 参数，路由到 controller/Game.php 下 verAction 函数。路由逻辑在 Application->run() 方法中，
路由之前，首先会调用 Filter::auth($params) 对参数验签，我们可以在该函数中加入自己的签名验证逻辑。

```php
//system/Application.php
class Application
{
    public function run(& $params, $clientInfo)
    {
        $ret = Filter::auth($params);
        if ($ret != 0) {
            return $ret;
        }

        foreach ($params as $k => $v) {
            $params[$k] = trim($v);
        }

        $controller = ucfirst($params['c']);
        $action = $params['m'] . "Action";
        $class_name = $controller . "Controller";

        try {
            $obj = new $class_name($params, $clientInfo);

            if (!method_exists($obj, $action)) {
                unset($obj);
                show_404("$controller/$action");
                return $this->response_error(3, "route error");
            }

            $ret = $obj->$action();
            unset($obj);
            return $ret;
        } catch (Exception $e) {
            unset($obj);

            if ($e instanceof LogicException) { //业务异常
                $errorcode = $e->getCode() == 0 ? 8 : $e->getCode();
                return $this->response_error($errorcode, $e->getMessage());
            } else if ($e->getMessage() != 'swoole exit.') {
                Logger::error("Catch An Exception File=[" . $e->getFile() . "|" . $e->getLine() . "] Code=[" . $e->getCode() . "], Message=[" . $e->getMessage() . "]", "exception_log");

                echo "Catch An Exception \n";
                echo "File:" . $e->getFile() . "\n";
                echo "Line:" . $e->getLine() . "\n";
                echo "Code:" . $e->getCode() . "\n";
                echo "Message:" . $e->getMessage() . "\n";
                return $this->response_error(99, "system exception");
            } else {
                echo "swoole exit.\n";
                return $this->response_error(99, "application exit");
            }
        }
    }

   	...
}
```

# 控制器Controller
所有控制器位于：application/controllers 目录下，继承自SuperController，父类SuperController 的构造函数中会调用$this->init()函数，所以你的控制器如果有初始化任务，请写在 init 函数里。

提供4个返回函数：<br>
	response_error 返回报错信息给自己<br>
	response_success_to_all 返回数据给当前所有玩家，例如世界聊天<br>
	response_success_to_me 返回数据给自己<br>
	response_success_to_uids 返回数据给指定uid，在 server.php ，数据接入的时候，我们会将 uid 绑定到 socket fd 上面去<br>

```php
//server.php
$uid = intval($input['userid']);
if($uid > 0) {
    Connector::set_fd($uid, $ws);
}
```

```php
class GameController extends SuperController
{
    var $game_service;
    var $userinfo_service;

    public function init()
    {
        $this->userinfo_service = $this->loader->service('UserinfoService');
        $this->game_service = $this->loader->service('GameService');

        $this->util_log = $this->loader->logger('game_log');
    }

    //聊天接口
    public function chatAction()
    {
        $userId = $this->params['userid'];
        $type = intval($this->params['type']);  //0-世界 1-私聊
        $token = $this->params['token'];
        $nickname = $this->params['nickname'];
        $avatar_url = $this->params['avatar_url'];
        $content = $this->params['content'];
        $to_userid = intval($this->params['to_userid']);

        $this->userinfo_service->getZoneUserAndAuth($userId, $token);

        if (empty($content)) {
            return $this->response_error(13342339, '内容不能为空');
        }

        $result = array();
        $result['userid'] = $userId;
        $result['type'] = $type;
        $result['nickname'] = $nickname;
        $result['avatar_url'] = $avatar_url;
        $result['gender'] = $this->params['gender'];
        $result['vip_level'] = $this->params['vip_level'];
        $result['lv'] = $this->params['lv'];
        $result['content'] = $content;

        if ($type == 0) {
            return $this->response_success_to_all($result);
        } else if ($type == 1) {
            return $this->response_success_to_uids([$userId, $to_userid], $result);
        }
    }
}
```

# 过滤验签
application/Filter.php ， 在 auth 中写入验签方法，所有接口都会在这里校验， 所有GET、POST等参数放在 $params 里。
```php
class Filter
{
    //验签过程
    public static function auth(& $params)
    {
        /*
        if($auth_error == false) { //验签失败
            return self::response_error(123, "auth error");
        } */

        //验签成功
        return 0;
    }

    public static function response_error($code, $message)
    {
        $data = array("code" => $code, "msg" => $message);
        $result['send_user'] = "me";
        $result['msg'] = json_encode($data);
        return $result;
    }
}
```
# 加载器
通过 Loader 加载器可以加载业务层，dao层，公共库，日志、配置等对象， Logger 为日志类。
```php
$this->game_service = $this->loader->service('GameService');
$this->game_dao = $this->loader->dao("GameDao");
$this->util_log = $this->loader->logger('game_log');
$this->util_lib = $this->loader->library('Utillib');
$this->conf = $this->loader->config('config');

```
# 业务层
通过 $this->game_service = $this->loader->service('GameService'); 去加载业务层。<br>
Service 继承自 SuperService，在 init() 函数里面实现对象初始化内容。
```php
class GameService extends SuperService
{
    public function init()
    {
        parent::init();
        $this->game_dao = $this->loader->dao("GameDao");
        $this->userinfo_service = $this->loader->dao("UserinfoService");
        $this->util_log = $this->loader->logger('game_log');
    }

    //用户充值
    public function get_user_vip_contents($userid)
    {
        $data = $this->game_dao->get_user_vip_contents($userid);

        if (empty($data['content'])) {
            $content = array();
            $content['leiji_xiaofei'] = 0;  //累计消费
            $content['leiji_chong'] = 0;  //累计充值
            $content['jijin']['status'] = 0;  //是否购买成长基金 0-未购买 1-已购买
            $this->game_dao->insert_user_vip_contents($userid, $content);
        } else {
            $content = json_decode($data['content'], true);
        }

        return $content;
    }

    //更新充值信息
    public function update_user_vip_contents($userid, $content)
    {
        return $this->game_dao->update_user_vip_contents($userid, $content);
    }
    
    ...
}
```

# Dao层
所有与Redis、MySQL等等存储介质打交道的逻辑，最好都放在Dao层，<br>
dao对象通过 $this->game_dao = $this->loader->dao("GameDao"); 加载。<br>
Dao层继承自 SuperDao，在 init() 函数里面实现对象初始化内容。
SuperDao 提供了许多快速操作数据库的方法，如果你需要用到 SuperDao 的快速操作数据库的函数，
你最好指定以下数据库、缓存配置，因为默认他们是 default， 这些配置位于application/config 目录下的 database.php 和 redis.php 中。<br>

$this->redis_name = "default";<br>
$this->db_name = "default";
 
```php
class GameDao extends SuperDao
{
    public function init()
    {
        $this->db_name = "game";
        $this->util_log = $this->loader->logger('game_log');
    }
    
    //user_vip_contents 表
    public function get_user_vip_contents($userid)
    {
        $key = 'pre_vip_contents_' . $userid;
        $data = $this->get_one_table_data('user_vip_contents', ['user_id' => $userid], $key);
        return $data;
    }

    public function insert_user_vip_contents($userid, $content)
    {
        $key = 'pre_vip_contents_' . $userid;
        return $this->insert_table('user_vip_contents', ['user_id' => $userid, 'content' => json_encode($content)], $key);
    }
    
    ...
}
```

```php
//数据库配置 database.php
$util_db_config['default']['host'] = '127.0.0.1';
$util_db_config['default']['username'] = 'test';
$util_db_config['default']['password'] = 'test';
$util_db_config['default']['dbname'] = 'user';
$util_db_config['default']['char_set'] = 'utf8';
$util_db_config['default']['dbcollat'] = 'utf8_general_ci';
$util_db_config['default']['pool_size'] = 10;

//redis配置 redis.php
$util_redis_conf['userinfo']['host'] = '127.0.0.1';
$util_redis_conf['userinfo']['port'] = 6381;
$util_redis_conf['userinfo']['auth'] = 'o01nc7vgd65xa';

//使用方法
MySQLPool::instance('default')->query($sql);
MySQLPool::instance('default')->get($table, $where, $column);
RedisPool::instance('userinfo')->set('test', 123);
RedisPool::instance('userinfo')->expire('test', 86400);
```

# library库
第三方类库都存在于 application/library 目录下 ，通过$this->utillib = $this->loader->library("Utillib"); 实例化。<br>

# 日志
日志可以通过 loader 实例化，实例化的日志会打印有请求参数和客户端IP等信息，也可以用得静态函数，不过静态函数无法获取则请求参数或者客户端IP等信息。<br>

日志路径在 server.php 中配置，记得把 /data/app/logs 的权限设置高些，define('LOG_PATH', '/data/app/logs/super_server'); //日志目录 <br>

日志分如下5个级别：<br>
const DEBUG = 'DEBUG';   /* 级别为 1 ,  调试日志,   当 DEBUG = 1 的时候才会打印调试 */<br>
const INFO = 'INFO';    /* 级别为 2 ,  应用信息记录,  与业务相关, 这里可以添加统计信息 */<br>
const NOTICE = 'NOTICE';  /* 级别为 3 ,  提示日志,  用户不当操作，或者恶意刷频等行为，比INFO级别高，但是不需要报告*/<br>
const WARN = 'WARN';    /* 级别为 4 ,  警告,   应该在这个时候进行一些修复性的工作，系统可以继续运行下去 */<br>
const ERROR = 'ERROR';   /* 级别为 5 ,  错误,     可以进行一些修复性的工作，但无法确定系统会正常的工作下去，系统在以后的某个阶段， 很可能因为当前的这个问题，导致一个无法修复的错误(例如宕机),但也可能一直工作到停止有不出现严重问题 */<br>

```
class GameService extends SuperService
{
    public function init()
    {
        parent::init();
        $this->util_log = $this->loader->logger('game_log');
    }
    
    public funciton test() 
    {
    	$this->util_log->LogInfo("info test");
	$this->util_log->LogNotice("notice test");
	$this->util_log->LogWarn("warning test");
	$this->util_log->LogError("error test");
    }
    
    public funciton static_test() 
    {
    	Logger::info("static info test");
	Logger::notice("static notice test");
	Logger::warn("static warning test");
	Logger::error("static error test");
	
    }
}
```


## 附录 - CoreModel 中的辅助极速开发函数（不关心可以跳过）

```php
/**
 * 根据key获取表记录
 * @param string redis_key redis 缓存键值
 */
public function hget_redis($redis_key, $field);
/**
 * 设置 redis 值
 * @param string redis_key redis 缓存键值, 可空， 非空时清理键值缓存
 * @param array data 表数据
 * @param int redis_expire redis 缓存到期时长(秒)
 * @param boolean set_empty_flag 是否缓存空值，如果缓存空值，在表记录更新之后，一定记得清理空值标记缓存
 */
public function hset_redis($redis_key, $field, $data, $redis_expire = 600, $set_empty_flag = true);
/**
 * 根据key获取表记录
 * @param string redis_key redis 缓存键值
 */
public function get_redis($redis_key)
/**
 * 设置 redis 值
 * @param string redis_key redis 缓存键值, 可空， 非空时清理键值缓存  
 * @param array data 表数据
 * @param int redis_expire redis 缓存到期时长(秒)
 * @param boolean set_empty_flag 是否缓存空值，如果缓存空值，在表记录更新之后，一定记得清理空值标记缓存
 */
public function set_redis($redis_key, $data, $redis_expire = 600, $set_empty_flag = true);
/**
 * 清理记录缓存
 * @param string redis_key redis 缓存键值
 */
public function clear_redis_cache($redis_key = "");
/**
 * 插入表记录
 * @param string table 表名
 * @param array data 表数据
 * @param string redis_key redis 缓存键值, 可空， 非空时清理键值缓存
 */
public function insert_table($table, $data, $redis_key = "");
/**
 * 更新表记录
 * @param string table 表名
 * @param array where 查询条件
 * @param array data 更新数据
 * @param string redis_key redis 缓存键值, 可空， 非空时清理键值缓存
 */
public function update_table($table, $where, $data, $redis_key = "");
/**
 * 替换表记录
 * @param string table 表名
 * @param array data 替换数据
 * @param string redis_key redis 缓存键值, 可空， 非空时清理键值缓存
 */
public function replace_table($table, $data, $redis_key = "");
/**
 * 删除表记录
 * @param string table 表名
 * @param array where 查询条件
 * @param string redis_key redis缓存键值, 可空， 非空时清理键值缓存
 */
public function delete_table($table, $where, $redis_key = "");
/**
 * 获取表数据
 * @param string table 表名
 * @param array where 查询条件
 * @param string redis_key redis 缓存键值, 可空， 非空时清理键值缓存
 * @param int redis_expire redis 缓存到期时长(秒)
 * @param string $column 数据库表字段，可空
 * @param boolean set_empty_flag 是否将空值写入缓存，防止数据库击穿，默认为是
 */
public function get_table_data($table, $where = array(), $redis_key = "", $redis_expire = 600, $column = "*", $set_empty_flag = true);
/**
 * 获取一条表数据
 * @param string table 表名
 * @param array where 查询条件
 * @param string redis_key redis 缓存键值, 可空， 非空时清理键值缓存
 * @param int redis_expire redis 缓存到期时长(秒)
 * @param string $column 数据库表字段，可空
 * @param boolean set_empty_flag 是否将空值写入缓存，防止数据库击穿，默认为是
 */
public function get_one_table_data($table, $where, $redis_key = "", $redis_expire = 600, $column = "*", $set_empty_flag = true);
```
