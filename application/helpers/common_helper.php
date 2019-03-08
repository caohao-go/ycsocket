<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');
/**
 * @package        SuperCI
 * @subpackage    Function
 * @category      Common Function
 * @author        caohao
 */
if (!function_exists('http_open'))
{
function http_open($url, $post = '', &$errmsg = '', $refer = '', $cookie = '', $out_time = 2)
{
    if (empty($url)) {
        $errmsg = 'URL参数不能为空';
        return false;
    }
        
    if (!function_exists('curl_init')) {
        $errmsg = 'CURL模块没有加载';
        return false;
    }
        
    $ch = curl_init();
    if ($ch === false) {
            $errmsg = '初始化句柄错误';
            return false;
    }
        
    curl_setopt($ch, CURLOPT_URL, $url);
    curl_setopt($ch, CURLOPT_HEADER, false);
        
    if ($post) {
        curl_setopt($ch, CURLOPT_POST, 1);
        curl_setopt($ch, CURLOPT_POSTFIELDS, $post);
    }
        
    if ($refer) {
        curl_setopt($ch, CURLOPT_REFERER, $refer);
    }
        
    if ($cookie) {
        curl_setopt($ch, CURLOPT_COOKIE, $cookie);
    }
            
    if(isset($_SERVER['HTTP_USER_AGENT']))
    {
        curl_setopt($ch, CURLOPT_USERAGENT, $_SERVER['HTTP_USER_AGENT']);
    }
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
    curl_setopt($ch, CURLOPT_TIMEOUT, $out_time);
        
    $data = curl_exec($ch);
        
    if ($data === false) {
        $errmsg = curl_error($ch);
        curl_close($ch);
        return false;
    }
    return $data;
} 
}

if ( ! function_exists('createLinkstringUrlencode'))
{
/**
 * 把数组所有元素，按照“参数=参数值”的模式用“&”字符拼接成字符串，并对字符串做urlencode编码
 * @param $para 需要拼接的数组
 * return 拼接完成以后的字符串
 */
function createLinkstringUrlencode($para) {
    $arg  = "";
    
    if(is_array($para)) {
        foreach($para as $key => $val) {
            $arg.=$key."=".urlencode($val)."&";
        }
    }
    
    //去掉最后一个&字符
    $arg = rtrim(trim($arg), "&");
    
    //如果存在转义字符，那么去掉转义
    if(get_magic_quotes_gpc()){$arg = stripslashes($arg);}
    
    return $arg;
}
}

if ( ! function_exists('https_request'))
{
function https_request($url, $data=''){
    $ch = curl_init();

    curl_setopt($ch, CURLOPT_URL, $url);
    curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, false);
    curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, false);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
    if($data){
        if(is_array($data)) {
            $data = http_build_query($data);
        }
        curl_setopt($ch, CURLOPT_POST, 1);
        curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
    }
    if(!defined('CURL_TIMEOUT_MAX')){
        define('CURL_TIMEOUT_MAX', 20);
    }
    curl_setopt($ch, CURLOPT_TIMEOUT, CURL_TIMEOUT_MAX);
    $result = curl_exec($ch);
    curl_close($ch);
        
    return $result;
}
}

if ( ! function_exists('pkcs5_pad'))
{
function pkcs5_pad ($text, $blocksize) {
    $pad = $blocksize - (strlen($text) % $blocksize);
    return $text . str_repeat(chr($pad), $pad);
}
}

if ( ! function_exists('encrypt'))
{
function encrypt($input, $key) {
        $size = mcrypt_get_block_size(MCRYPT_RIJNDAEL_128, MCRYPT_MODE_ECB);
        $input = pkcs5_pad($input, $size);
        $td = mcrypt_module_open(MCRYPT_RIJNDAEL_128, '', MCRYPT_MODE_ECB, '');
        $iv = mcrypt_create_iv(mcrypt_enc_get_iv_size($td), MCRYPT_RAND);
        mcrypt_generic_init($td, $key, $iv);
        $data = mcrypt_generic($td, $input);
        mcrypt_generic_deinit($td);
        mcrypt_module_close($td);
        return base64_encode($data);
}
}
 
if ( ! function_exists('decrypt'))
{
function decrypt($sStr, $sKey) {
        $sStr = base64_decode($sStr);
        $decrypted= mcrypt_decrypt(
            MCRYPT_RIJNDAEL_128,
            $sKey,
            $sStr,
            MCRYPT_MODE_ECB
        );
        $dec_s = strlen($decrypted);
        $padding = ord($decrypted[$dec_s-1]);
        $decrypted = substr($decrypted, 0, -$padding);
        return $decrypted;
}
}