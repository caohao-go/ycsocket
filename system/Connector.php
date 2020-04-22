<?php

class Connector
{
	const connect_expire_time = 300;  //超时关闭连接
    public static $conn_map = array();  //连接
    public static $heartbeat_times_map = array();  //心跳到期时间

    public static function is_online($fd)
    {
        return isset(self::$conn_map[$fd]) ? true : false;
    }

    //绑定 ws 到 fd
    public static function set_fd($fd, & $ws)
    {
        if (!empty(self::$conn_map[$fd])) { //同一账号不能同时登陆
            if ($ws->fd != self::$conn_map[$fd]->fd) {
                self::$conn_map[$fd]->push(json_encode(array("code" => "2", "msg" => "input error", "data" => "同一账号不能同时登陆")));
                self::$conn_map[$fd]->close();
            }
        }

        self::$conn_map[$fd] = & $ws;
        self::$heartbeat_times_map[$fd] = time() + Connector::connect_expire_time;
    }

    public static function set_connect_expire($fd)
    {
        self::$heartbeat_times_map[$fd] = time() + Connector::connect_expire_time;
    }

    public static function close($fd)
    {
    	if (!empty(self::$conn_map[$fd])) {
	        self::$conn_map[$fd]->close();
	    }
	    
		unset(self::$conn_map[$fd]);
		unset(self::$heartbeat_times_map[$fd]);
    }

    public static function send_all($msg)
    {
        go(function () use ($msg) {
            foreach (self::$conn_map as $fd => & $ws) {
                if ($fd % 1000 != GAME_ZONE_ID) {
                    continue;
                }

                $ws->push($msg);
            }
        });
    }

    public static function send_fds($fds, $msg)
    {
        go(function () use ($fds, $msg) {
            foreach ($fds as $fd) {
                if (!empty(self::$conn_map[$fd])) {
                    self::$conn_map[$fd]->push($msg);
                }
            }
        });
    }

    public static function send($fd, $msg)
    {
        if (empty(self::$conn_map[$fd])) {
            return false;
        }

        return self::$conn_map[$fd]->push($msg);
    }


    public static function connect_expire()
    {
        foreach (self::$heartbeat_times_map as $fd => $quit_time) {
            try {
                if (time() >= $quit_time) { //超过2分钟未收到心跳，关闭连接
                    if (isset(self::$conn_map[$fd])) {
                        self::$conn_map[$fd]->close();
                        unset(self::$conn_map[$fd]);
                        //修改下线时间
                        //MySQLPool::instance('game')->update('user_grade', ['user_id' => $fd], ['off_time' => date('Y-m-d H:i:s')]);
                    }
                    unset(self::$heartbeat_times_map[$fd]);
                } else {
                    return;
                }
            } catch (Exception $e) {
                Logger::error("Catch An Exception File=[" . $e->getFile() . "|" . $e->getLine() . "] Code=[" . $e->getCode() . "], Message=[" . $e->getMessage() . "]");
            }
        }
    }
}
