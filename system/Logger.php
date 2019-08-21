<?php
/**
 * Logger Class  https://github.com/caohao-php/ycsocket
 *
 * @package       Ycsocket
 * @subpackage    Libraries
 * @category      Logger
 * @author        caohao
 */
define('DEFAULT_LOG_FILE_NAME', 'default');

define('DEBUG', 1);  /* 是否调试  0-不打印调试日志  1-打印调试日志 */

class Logger {
    private $LogPath;
    private $LogFileName;
    private $m_InitOk = false;
    private $ip;
    private $params;

    /**
     * @__construct 初始化
     * @param $config 日志配置:
    $config['log_path'];  -- 日志目录, 一般采用默认 LOG_PATH
    $config['file_name']; -- 日志文件名, 不写默认为 DEFAULT_LOG_FILE_NAME
     * @param
     * @return
     */
    public function __construct($config = 0) {
        /* 日志目录, 不写默认为 LOG_PATH */
        $this->LogPath = empty($config['log_path']) ? LOG_PATH : $config['log_path'];

        /* 删除目录最后一个'/' */
        if (substr($this->LogPath, -1) == '/') {
            $this->LogPath = substr($this->LogPath, 0, -1);
        }

        /* 日志文件, 不写默认为 DEFAULT_LOG_FILE_NAME */
        $this->LogFileName = empty($config['file_name']) ? DEFAULT_LOG_FILE_NAME : $config['file_name'];
        $this->LogFileName = $this->LogPath . "/" . $this->LogFileName . "." . date('Ymd') . ".log";
    }

    /**
     * @abstract 写入日志
     * @param String $log 内容
     */
    public function LogDebug($log, $error_code=0) {
        if (DEBUG) $this->Log(Logger::DEBUG, $log, $error_code);
    }

    public function LogInfo($log) {
        $this->Log(Logger::INFO, $log, 0);
    }

    public function LogNotice($log) {
        $this->Log(Logger::NOTICE, $log, 0);
    }

    public function LogWarn($log, $error_code=0) {
        $this->Log(Logger::WARN, $log, $error_code);
    }

    public function LogError($log, $error_code=0) {
        $this->Log(Logger::ERROR, $log, $error_code);
    }

    public function LogFatal($log, $error_code=0) {
        $this->Log(Logger::FATAL, $log, $error_code);
    }

    private function Log($privity, $error_msg, $error_code) {
        if ($this->m_InitOk == false) {
            if (is_dir($this->LogPath)) {
                $this->m_InitOk = true;
            } else {
                $this->m_InitOk =  @mkdir($this->LogPath, 0777, true);
                if (!$this->m_InitOk) {
                    return false;
                }
            }
        }

        $datestr = strftime("%Y-%m-%d %H:%M:%S");
        $uri = json_encode($this->params);
        $referer = "";
        $cookie = "";

        /* 日志格式: [日志级别] [时间] [错误代码] [文件|行数] [ip] [uri] [referer] [cookie] [统计信息] "内容" */
        if ($privity === Logger::INFO) { //INFO日志
            $log = sprintf( "[%s] [%s] - - [%s] - - - [%s] \"%s\"\n",
                            $privity,
                            $datestr,
                            $this->ip,
                            '',
                            $error_msg);
            file_put_contents($this->LogFileName, $log, FILE_APPEND);
        } else if ($privity === Logger::NOTICE) { //提示日志
            $log = sprintf( "[%s] [%s] - - [%s] [%s] [%s] [%s] - \"%s\"\n",
                            $privity,
                            $datestr,
                            $this->ip,
                            $uri,
                            $referer,
                            $cookie,
                            $error_msg);

            file_put_contents($this->LogFileName, $log, FILE_APPEND);
        } else if ($privity === Logger::DEBUG || $privity === Logger::NOTICE) { //调试日志
            $log = sprintf( "[%s] [%s] - [%s] [%s] [%s] [%s] [%s] - \"%s\"\n",
                            $privity,
                            $datestr,
                            $this->GetCallerInfo(),
                            $this->ip,
                            $uri,
                            $referer,
                            $cookie,
                            $error_msg);

            file_put_contents($this->LogFileName.".debug", $log, FILE_APPEND);
        } else {
            $log = sprintf( "[%s] [%s] [%d] [%s] [%s] [%s] [%s] [%s] - \"%s\"\n",
                            $privity,
                            $datestr,
                            $error_code,
                            $this->GetCallerInfo(),
                            $this->ip,
                            $uri,
                            $referer,
                            $cookie,
                            $error_msg);

            file_put_contents($this->LogFileName.".wf", $log, FILE_APPEND);
        }
    }

    private function GetCallerInfo() {
        $ret = debug_backtrace();

        $call_info = array();
        foreach ($ret as $item) {
            if (isset($item['class']) && 'Logger' == $item['class']) {
                $last_item = $item;
                continue;
            } else {
                $call_info[] = basename($last_item['file']). "|".$last_item['line'];
                $last_item = $item;
            }
        }

        $call_info[] = basename($last_item['file']). "|".$last_item['line'];
        $call_info = array_reverse($call_info);
        return implode($call_info, " => ");
    }

    public function setParams(& $params) {
        $this->params = & $params;
    }

    public function setClientIp($ip) {
        $this->ip = $ip;
    }

    const DEBUG  = 'DEBUG';   /* 级别为 1 ,  调试日志,   当 DEBUG = 1 的时候才会打印调试 */
    const INFO   = 'INFO';    /* 级别为 2 ,  应用信息记录,  与业务相关, 这里可以添加统计信息 */
    const NOTICE = 'NOTICE';  /* 级别为 3 ,  提示日志,  用户不当操作，或者恶意刷频等行为，比INFO级别高，但是不需要报告*/
    const WARN  = 'WARN';    /* 级别为 4 ,  警告,   应该在这个时候进行一些修复性的工作，系统可以继续运行下去 */
    const ERROR   = 'ERROR';   /* 级别为 5 ,  错误,     可以进行一些修复性的工作，但无法确定系统会正常的工作下去，系统在以后的某个阶段， 很可能因为当前的这个问题，导致一个无法修复的错误(例如宕机),但也可能一直工作到停止有不出现严重问题 */
    const FATAL  = 'FATAL';   /* 级别为 6 ,  严重错误,  这种错误已经无法修复，并且如果系统继续运行下去的话，可以肯定必然会越来越乱, 这时候采取的最好的措施不是试图将系统状态恢复到正常，而是尽可能的保留有效数据并停止运行 */
}
