<?php

class Ip_Location
{

    private static $ip = null;

    private static $fp = null;

    private static $offset = null;

    private static $index = null;

    private static $cached = array();

    public static function find($ip)
    {
        if (empty( $ip ) === true) {
            return 'N/A';
        }
        
        $nip   = gethostbyname($ip);
        $ipdot = explode('.', $nip);

        if ($ipdot[0] < 0 || $ipdot[0] > 255 || count($ipdot) !== 4) {
            return 'N/A';
        }

        if (isset( self::$cached[$nip] ) === true) {
            return self::$cached[$nip];
        }

        if (self::$fp === null) {
            $ret = self::init();
            if(!$ret) {
                return false;
            }
        }

        $nip2 = pack('N', ip2long($nip));
        $tmp_offset = (int) $ipdot[0] * 4;
        //echo $tmp_offset."-";
        
        //$start = unpack('Vlen', self::$index[$tmp_offset].self::$index[$tmp_offset + 1].self::$index[$tmp_offset + 2].self::$index[$tmp_offset + 3]);
        $start = unpack('Vlen', self::get_unpack_info($tmp_offset, 4));
        $index_offset = $index_length = null;
        $max_comp_len = self::$offset['len'] - 1024 - 4;
        for ($start = $start['len'] * 8 + 1024; $start < $max_comp_len; $start += 8) {
            //if (self::$index{$start}.self::$index{$start + 1}.self::$index{$start + 2}.self::$index{$start + 3} >= $nip2) {
            if(self::get_unpack_info($start, 4) >= $nip2) {
                //$index_offset = unpack('Vlen', self::$index{$start + 4}.self::$index{$start + 5}.self::$index{$start + 6}."\x0");
                $index_offset = unpack('Vlen', self::get_unpack_info($start + 4, 3) . "\x0");
                
                //$index_length = unpack('Clen', self::$index{$start + 7});
                $index_length = unpack('Clen', self::get_unpack_info($start + 7, 1));
                break;
            }
        }

        if ($index_offset === null) {
            return 'N/A';
        }
        
        fseek(self::$fp, self::$offset['len'] + $index_offset['len'] - 1024);
        //echo fread(self::$fp, $index_length['len']);exit;
        self::$cached[$nip] = explode("\t", fread(self::$fp, $index_length['len']));

        return self::$cached[$nip];
    }
    
    private static function get_unpack_info($offset, $num) {
        fseek(self::$fp, $offset+4);
        return fread(self::$fp, $num);
    }

    private static function init()
    {
        if (self::$fp === null) {
            self::$ip = new self();

            self::$fp = fopen(__DIR__.'/ip.dat', 'rb');
            if (self::$fp === false) {
                return false;
            }
            
            self::$offset = unpack('Nlen', fread(self::$fp, 4));
            if(self::$offset < 4) {
                return false;
            }
            
            //self::$index = fread(self::$fp, self::$offset['len'] - 4);
        }
        
        return true;
    }

    public function __destruct()
    {
        if (self::$fp !== null) {
            fclose(self::$fp);
        }
    }
}
