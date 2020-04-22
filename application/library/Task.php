<?php

class Task
{
    static public $task_configs;
    static public $achieve_datas;

    static public function init_task()
    {
        /*
        $datas = MySQLPool::instance("game")->query("select * from task_config");
        foreach ($datas as $data) {
            $id = $data['id'];
            $type = $data['type'];
            self::$task_configs[$type][$id] = $data;
        }

        //成就任务
        $datas = MySQLPool::instance("game")->query("select * from task_achieve");
        foreach ($datas as $data) {
            $id_tmp = $data['id'];
            self::$achieve_datas[$id_tmp] = $data;
        }
        */
    }
}
