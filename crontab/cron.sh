#!/bin/sh
/usr/bin/php /data/app/super_server/crontab/send_rewards.php 1 &
/usr/bin/php /data/app/super_server/crontab/send_rewards.php 2 &
