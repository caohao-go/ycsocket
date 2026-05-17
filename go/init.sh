rm -rf /usr/local/var/db/redis/appendonly.aof
rm -rf /usr/local/var/db/redis/dump.rdb
ps -ef | grep redis | grep -v grep |awk  '{print $2}'|xargs kill -9

ps -ef | grep redis

redis-server /usr/local/etc/redis.conf

DELETE FROM shine_world_1.guild where id>0;
DELETE FROM shine_world_1.guild_chapter_blood where id>0;
DELETE FROM shine_world_1.template_level where id>0;
DELETE FROM shine_world_1.user_climbtower_record where layer>0;
DELETE FROM shine_world_1.user_endless_help_hero where id>0;
DELETE FROM shine_world_1.user_expedition_help_hero where id>0;
DELETE FROM shine_world_1.user_hero where id>0;
DELETE FROM shine_world_1.user_mail where id>0;
DELETE FROM shine_world_1.user_nickname where user_id>0;
DELETE FROM shine_world_1.users_friends where id>0;
DELETE FROM shine_world_1.users_guild where id>0;
DELETE FROM shine_world_1.user_tongguan_reward where id>0;
DELETE FROM shine_world_1.user_tuteng_pk_detail where id>0;


## redis 清理
hdel user_achieve_$uid $field
hdel userattr_$uid $field
hdel daily_$today_$uid $field
hdel content_$uid $content_type
hdel items_1_$uid $item_id
hdel items_2_$uid $item_id
hdel items_3_$uid $item_id
hdel items_4_$uid $item_id
hdel items_5_$uid $item_id




