<?php
/** Singleton.php
 * SuperModel Class
 *
 * @package			Ycsocket  https://github.com/caohao-php/ycsocket
 * @subpackage		Singleton
 * @category		Singleton
 * @author			caohao
 */
trait Singleton
{
    private static $instance;

    static function getInstance(...$args)
    {
        if(!isset(self::$instance)){
            self::$instance = new static(...$args);
        }
        return self::$instance;
    }
}
