<?php
/**
 * Logger Class
 *
 * @package       Ycsocket
 * @subpackage    Libraries
 * @category      Logger
 * @author        caohao
 */
define('DEBUG', 1);  /* 是否调试  0-不打印调试日志  1-打印调试日志 */

class Logger
{
    public static $trace;
    private $LogFileName;
    private $m_InitOk = false;
    private $ip = "";
    private $params = array();

    /**
     * @__construct 初始化
     * @param $config 日志配置:
     * $config['log_path'];  -- 日志目录, 一般采用默认 LOG_PATH
     * $config['file_name']; -- 日志文件名, 不写默认为 DEFAULT_LOG_FILE_NAME
     * @param
     * @return
     */
    public function __construct($file_name = 'default')
    {
        /* 日志文件, 不写默认为 DEFAULT_LOG_FILE_NAME */
        $this->LogFileName = $file_name;
        $this->LogFileName = LOG_PATH . "/" . $this->LogFileName . "." . date('Ymd') . ".log";
    }

    public static function debug($log, $file_name = 'trace') {
        if(empty(self::$trace[$file_name])) {
            self::$trace[$file_name] = new Logger($file_name);
        }

        self::$trace[$file_name]->Log(Logger::DEBUG, $log);
    }

    public static function info($log, $file_name = 'trace') {
        if(empty(self::$trace[$file_name])) {
            self::$trace[$file_name] = new Logger($file_name);
        }

        self::$trace[$file_name]->Log(Logger::INFO, $log, 0);
    }

    public static function notice($log, $file_name = 'trace') {
        if(empty(self::$trace[$file_name])) {
            self::$trace[$file_name] = new Logger($file_name);
        }

        self::$trace[$file_name]->Log(Logger::NOTICE, $log, 0);
    }

    public static function warn($log, $file_name = 'trace') {
        if(empty(self::$trace[$file_name])) {
            self::$trace[$file_name] = new Logger($file_name);
        }

        self::$trace[$file_name]->Log(Logger::WARN, $log);
    }

    public static function error($log, $file_name = 'trace') {
        if(empty(self::$trace[$file_name])) {
            self::$trace[$file_name] = new Logger($file_name);
        }

        self::$trace[$file_name]->Log(Logger::ERROR, $log);
    }

    /**
     * @abstract 写入日志
     * @param String $log 内容
     */
    public function LogDebug($log)
    {
        if (DEBUG) $this->Log(Logger::DEBUG, $log);
    }

    public function LogInfo($log)
    {
        $this->Log(Logger::INFO, $log, 0);
    }

    public function LogNotice($log)
    {
        $this->Log(Logger::NOTICE, $log, 0);
    }

    public function LogWarn($log)
    {
        $this->Log(Logger::WARN, $log);
    }

    public function LogError($log)
    {
        $this->Log(Logger::ERROR, $log);
    }

    private function Log($privity, $error_msg)
    {
        if ($this->m_InitOk == false) {
            if (is_dir(LOG_PATH)) {
                $this->m_InitOk = true;
            } else {
                $this->m_InitOk = @mkdir(LOG_PATH, 0777, true);
                if (!$this->m_InitOk) {
                    return false;
                }
            }
        }

        $datestr = strftime("%Y-%m-%d %H:%M:%S");
        $uri = empty($this->params) ? "" : json_encode($this->params);
        $referer = "";
        $cookie = "";

        /* 日志格式: [日志级别] [时间] [区] [文件|行数] [ip] [uri] [referer] [cookie] "内容" */
        if ($privity === Logger::INFO) { //INFO日志
            $log = sprintf("[%s] [%s] [%s] - [%s] - - - \"%s\"\n",
                $privity,
                $datestr,
                GAME_ZONE_ID,
                $this->ip,
                $error_msg);
            file_put_contents($this->LogFileName, $log, FILE_APPEND);
        } else if ($privity === Logger::NOTICE) { //提示日志
            $log = sprintf("[%s] [%s] [%s] - [%s] [%s] [%s] [%s] \"%s\"\n",
                $privity,
                $datestr,
                GAME_ZONE_ID,
                $this->ip,
                $uri,
                $referer,
                $cookie,
                $error_msg);

            file_put_contents($this->LogFileName, $log, FILE_APPEND);
        } else if ($privity === Logger::DEBUG) { //调试日志
            $log = sprintf("[%s] [%s] [%s] - [%s] [%s] [%s] [%s] [%s] \"%s\"\n",
                $privity,
                $datestr,
                GAME_ZONE_ID,
                $this->GetCallerInfo(),
                $this->ip,
                $uri,
                $referer,
                $cookie,
                $error_msg);

            file_put_contents($this->LogFileName . ".debug", $log, FILE_APPEND);
        } else {
            $log = sprintf("[%s] [%s] [%s] [%s] [%s] [%s] [%s] [%s] \"%s\"\n",
                $privity,
                $datestr,
                GAME_ZONE_ID,
                $this->GetCallerInfo(),
                $this->ip,
                $uri,
                $referer,
                $cookie,
                $error_msg);

            file_put_contents($this->LogFileName . ".wf", $log, FILE_APPEND);
        }
    }

    private function GetCallerInfo()
    {
        $ret = debug_backtrace();

        $call_info = array();
        foreach ($ret as $item) {
            if (isset($item['class']) && 'Logger' == $item['class']) {
                $last_item = $item;
                continue;
            } else {
                $call_info[] = basename($last_item['file']) . "|" . $last_item['line'];
                $last_item = $item;
            }
        }

        $call_info[] = basename($last_item['file']) . "|" . $last_item['line'];
        $call_info = array_reverse($call_info);
        return implode($call_info, " => ");
    }

    public function setParams(& $params)
    {
        $this->params = &$params;
    }

    public function setClientIp($ip)
    {
        $this->ip = $ip;
    }

    const DEBUG = 'DEBUG';   /* 级别为 1 ,  调试日志,   当 DEBUG = 1 的时候才会打印调试 */
    const INFO = 'INFO';    /* 级别为 2 ,  应用信息记录,  与业务相关, 这里可以添加统计信息 */
    const NOTICE = 'NOTICE';  /* 级别为 3 ,  提示日志,  用户不当操作，或者恶意刷频等行为，比INFO级别高，但是不需要报告*/
    const WARN = 'WARN';    /* 级别为 4 ,  警告,   应该在这个时候进行一些修复性的工作，系统可以继续运行下去 */
    const ERROR = 'ERROR';   /* 级别为 5 ,  错误,     可以进行一些修复性的工作，但无法确定系统会正常的工作下去，系统在以后的某个阶段， 很可能因为当前的这个问题，导致一个无法修复的错误(例如宕机),但也可能一直工作到停止有不出现严重问题 */
}
