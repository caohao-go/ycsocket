<?php if (!defined('BASEPATH')) exit('No direct script access allowed');

/**
 * ExampleDao Class
 *
 * @package            Ycsocket
 * @subpackage        Dao
 * @category        Example Dao
 * @author            caohao
 */
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

    public function update_user_vip_contents($userid, $content)
    {
        $key = 'pre_vip_contents_' . $userid;
        return $this->update_table('user_vip_contents', ['user_id' => $userid], ['content' => json_encode($content)], $key);
    }

    //user_hero 表
    public function get_user_heros($userid)
    {
        $key = 'pre_user_heros' . $userid;
        $data = $this->get_table_data('user_hero', ['user_id' => $userid], $key);
        return $data;
    }

    public function delete_user_heros($userid, $id)
    {
        $key = 'pre_user_heros' . $userid;
        return $this->delete_table('user_hero', ['id' => $id], $key);
    }

    public function replace_user_heros($userid, $replace_data)
    {
        $key = 'pre_user_heros' . $userid;
        return $this->replace_table('user_hero', $replace_data, $key);
    }
}
