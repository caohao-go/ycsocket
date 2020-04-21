<?php
if (!function_exists('show_error')) {
    function show_error($message, $status_code = 500, $heading = 'An Error Was Encountered')
    {
        echo "$heading\n";
        echo "Code: $status_code\n";
        if (is_array($message)) {
            foreach ($message as $val) {
                if (!empty($val)) echo "$val\n";
            }
        } else {
            echo "Message: $message\n";
        }

        echo "\n";
        exit;
    }
}

if (!function_exists('show_404')) {
    function show_404($page = '', $log_error = FALSE)
    {
        $heading = "404 Page Not Found";
        $message = "The page you requested was not found.  --> " . $page;

        // By default we log this, but allow a dev to skip it
        if ($log_error) {
            log_message('error', '404 Page Not Found --> ' . $page);
        }

        echo "$heading\n";
        echo "Code: 404\n";
        echo "message: $message\n\n";
    }
}

if (!function_exists('log_message')) {
    function log_message($level = 'error', $msg)
    {
        $level = strtoupper($level);
        $_levels = array('ERROR' => 1, 'DEBUG' => 2, 'WARNING' => 3, 'NOTICE' => 4, 'INFO' => 5, 'ALL' => 6);

        if (!isset($_levels[$level]) OR $_levels[$level] > PHP_LOG_THRESHOLD) {
            return false;
        }

        $message = $level . ' - ' . date('Y-m-d H:i:s') . ' --> ' . $msg . "\n";
        @file_put_contents(LOG_PATH . "/Ycsocket-" . date('Y-m-d') . ".log.wf", $message, FILE_APPEND);
    }
}

if (!function_exists('_exception_handler')) {
    function _exception_handler($severity, $message, $filepath, $line)
    {
        if ($severity == E_STRICT) {
            return;
        }

        $levels = array(
            E_ERROR => 'Error',
            E_WARNING => 'Warning',
            E_PARSE => 'Parsing Error',
            E_NOTICE => 'Notice',
            E_CORE_ERROR => 'Core Error',
            E_CORE_WARNING => 'Core Warning',
            E_COMPILE_ERROR => 'Compile Error',
            E_COMPILE_WARNING => 'Compile Warning',
            E_USER_ERROR => 'User Error',
            E_USER_WARNING => 'User Warning',
            E_USER_NOTICE => 'User Notice',
            E_STRICT => 'Runtime Notice'
        );

        $filepath = str_replace("\\", "/", $filepath);

        if (FALSE !== strpos($filepath, '/')) {
            $x = explode('/', $filepath);
            $filepath = $x[count($x) - 2] . '/' . end($x);
        }

        $severity_desc = (!isset($levels[$severity])) ? $severity : $levels[$severity];

        if (($severity & error_reporting()) == $severity) {
            echo "A PHP Error was encountered\n";
            echo "Severity: $severity_desc\n";
            echo "Filename: $filepath\n";
            echo "Line Number: $line\n";
            echo "Message: $message\n\n";
        }

        log_message($severity_desc, "[$filepath|$line] $message");
    }
}
