<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');

use Swoole\Coroutine\Channel;

/**
 * ExampleModel Class
 *
 * @package			Ycsocket
 * @subpackage		Model
 * @category		Example Model
 * @author			caohao
 */
class RoomLogic extends ActorBean {
    private static $instance;

    private $joiningRoom;
    private $playingRooms;
    private $joinedUser;

    public function __construct() {
    }

    public static function getInstance() {
        if (!isset(self::$instance)) {
            global $roomActorId;
            $actorIdArray = $roomActorId->get("RoomActorId");
            if (empty($actorIdArray['id'])) {
                self::$instance = RoomLogic::new();
                $roomActorId->set("RoomActorId", ['id' => self::$instance->getActorId()]);
            } else {
                self::$instance = RoomLogic::getBean($actorIdArray['id']);
            }
        }

        return self::$instance;
    }

    public function isUserJoined($userid) {
        $result = array();
        if (isset($this->joiningRoom['users'][$userid])) {
            $result['id'] = $this->joiningRoom['id'];
            $result['state'] = GameLogic::STATE_JOIN;
            $result['createTime'] = $this->joiningRoom['createTime'];
            $result['users'] = array_values($this->joiningRoom['users']);
            $this->joiningRoom['pkLogic']->setGameProxy($userid, false);
            return $result;
        }

        if (isset($this->joinedUser[$userid])) {
            $result['id'] = $pkid = $this->joinedUser[$userid];
            $result['state'] = $state = $this->playingRooms[$pkid]['pkLogic']->getState($userid);
            $result['createTime'] = $this->playingRooms[$pkid]['createTime'];
            $result['starttime'] = $this->playingRooms[$pkid]['starttime'];
            $result['users'] = array_values($this->playingRooms[$pkid]['users']);
            $this->playingRooms[$pkid]['pkLogic']->setGameProxy($userid, false);

            if ($state > GameLogic::STATE_INIT) {
                $result['tetris'] = $this->playingRooms[$pkid]['pkLogic']->getAllTetris();
            }

            return $result;
        }
    }

    public function joinRoom($userid, $nickname, $avatar_url) {
        if (empty($this->joiningRoom)) {
            $this->joiningRoom['pkLogic'] = PkLogic::new();
            $this->joiningRoom['createTime'] = time();
            $this->joiningRoom['pkLogic']->joinUser($userid);
            $this->joiningRoom['id'] = $this->joinedUser[$userid] = $this->joiningRoom['pkLogic']->getActorId();
            $this->joiningRoom['users'][$userid] = ['userid' => $userid, 'nickname' => $nickname, 'avatar_url' => $avatar_url];

            $this->waitJoin();
            return $this->joiningRoom;
        } else {
            foreach($this->joiningRoom['users'] as $user) {
                Userfd::getInstance()->send($user['userid'], ['c' => 'tt', 'm' => 'join', 'id' => $this->joiningRoom['id'], 'userid' => $userid, 'nickname' => $nickname, 'avatar_url' => $avatar_url]);
            }

            $this->joinedUser[$userid] = $this->joiningRoom['id'];

            $this->joiningRoom['users'][$userid] = ['userid' => $userid, 'nickname' => $nickname, 'avatar_url' => $avatar_url];
            $joinedCount = $this->joiningRoom['pkLogic']->joinUser($userid);
            if ($joinedCount >= 4) {
                return $this->startGame();
            } else {
                return $this->joiningRoom;
            }
        }
    }

    public function startGame() {
        $pkid = $this->joiningRoom['id'];
        $this->playingRooms[$pkid] = $this->joiningRoom;
        $this->joiningRoom = null;
        $this->playingRooms[$pkid]['starttime'] = time();
        $this->playingRooms[$pkid]['pkLogic']->startGame();
        return $this->playingRooms[$pkid];
    }

    public function proxyGame($userid) {
        if (!isset($this->joinedUser[$userid])) {
            return;
        }

        if (isset($this->joiningRoom['users'][$userid])) {
            $this->joiningRoom['pkLogic']->setGameProxy($userid, true);
        } else {
            $pkid = $this->joinedUser[$userid];
            $gameLogic = $this->playingRooms[$pkid]['pkLogic']->getGameLogicByUid($userid);
            if (empty($gameLogic)) {
                return;
            }

            $gameLogic->setProxy(true);

            go(function () use ($gameLogic) {
                $gameLogic->tickUpdate();
            }
              );
        }
    }

    public function getJoinedCount() {
        return count($this->joiningRoom['users']);
    }

    public function waitJoin() {
        go(function () {
            while (1) {
                co::sleep(1);

                if (empty($this->joiningRoom)) {
                    return;
                }

                $uid = rand(111, 9999);
                $nicknames = ['芒果', 'Smile、格调', '坚果pro', '凉之渡', '搞趣阿达', 'CH.smallhow', '洋洋洋', '刘会存', '靖', '开心果', '梦绕云山', '生命如水', 'json', 'Kevin', 'Henry', 'Heisey', '给自由以名义', 'MR.LIU', 'life', 'charlie'];
                $nickname_rand = rand(0, count($nicknames) - 1);
                $avatars = ['https://wx.qlogo.cn/mmopen/vi_32/DYAIOgq83epqg7FwyBUGd5xMXxLQXgW2TDEBhnNjPVla8GmKiccP0pFiaLK1BGpAJDMiaoyGHR9Nib2icIX9Na4Or0g/132',
                            'https://wx.qlogo.cn/mmopen/vi_32/Z5C3QFwrKmhvL2CibnBt2BwibrHY3AiaVcayvQqJLDuS9FtL4ha5ibG77hOonCvvgTwgnDV4ce1aoJvmkehvxjUWcA/132',
                            'https://wx.qlogo.cn/mmopen/vi_32/oiba5C0TtnibrTibwKQe2iclgykmXKibL4MhJmkDaibKr36wcce0QMm8vh8XvH2C2DXc96QRWQ8Mib3l233WToOyeUhpQ/132',
                            'https://wx.qlogo.cn/mmopen/vi_32/Q0j4TwGTfTL6SwcnPGFP5vlYPT4EHG2nPicf0dUqicuvuTk31HSbwIsyGPiasABl8VYCfJHYa4n3WfM0jRzHNnMOg/132',
                            'https://wx.qlogo.cn/mmopen/vi_32/PiajxSqBRaEIk9ZLV5zQ3DVVoqxibZr6ZG5AlIQHxy2T0Cc2veBwicMEbk5e1fPv16bRH4To7WE1yDn5IsJVcKfiaQ/132',
                            'https://wx.qlogo.cn/mmopen/vi_32/ic9v1meib9tArlwNFpJiarxlhqSMRYXbuMlibRtYUySTuHdERzrg6nQuvWdWu5RUput0dNS6lAusiam4x9M1M1TqCSg/132',
                            'https://wx.qlogo.cn/mmopen/vi_32/DYAIOgq83erLIsGAMmbGjMkQx9AWUxN1MP6rB73KcVILbmJbQRX0yhne8PEdRgMKRDpcvmvPonwlSeL5Bt4Qdg/132',
                            'https://wx.qlogo.cn/mmopen/vi_32/DYAIOgq83eruZodkjQRFOOdL32T9rAgZy9VT5lCo3CbFoKzUKLnmvNfNNuWxSSXutjKRs7t1cKqjuIAcJRFyfg/132',
                            'https://wx.qlogo.cn/mmopen/vi_32/y9KuUt8q4Rst51G9xUz7Vew2N8sqKE07OouiaesE54WzbWHeZVVX1LrP9IDFzxefhictc0csrMicQuGTtr8ibeHCBg/132',
                            'https://wx.qlogo.cn/mmopen/vi_32/DYAIOgq83eoUkTwAacDQEywicZic5P7wG3bj6cSPdogMBGEFIdTAyYBTeewKJIuiaIO41vG8RFESNibS4RcCf5zzYQ/132',
                            'https://wx.qlogo.cn/mmopen/vi_32/Iudw0p2utzibFibAMDwZ4iaXNloWNA3J3n6brSFM8NRI263hurKluDpXFlyT01TZZ0kkGTRl1BTaoXmETibu0YY0bw/132',
                            'https://wx.qlogo.cn/mmopen/vi_32/0iaWEjweUtZqGuZ2ClicpcGZ1xO047oDYhyWpQO5iblibsaylGjDGtWU1YwoYGGGicVKPWr8qic3ZHiaGHiclgtcFMbrnw/132',
                            'https://wx.qlogo.cn/mmopen/vi_32/Q0j4TwGTfTKvnsAuD4OMaokFTFP8Lo6RBwTCOmwQWgibPJV87I4dhM6Gpqu1w3hreFgX6BE6NCzGqJZQ2v1mr2w/132',
                            'https://wx.qlogo.cn/mmopen/vi_32/Q0j4TwGTfTLqqHPKKz5EFoys0UcicHPibyT6P9zmN87os2JmiaTCGJ0CoWNMrkwlfgHNIEnvIlsGTpPTGXHVJOnoA/132',
                            'https://wx.qlogo.cn/mmopen/vi_32/DYAIOgq83epFaDXf8PVNCCgGaUHia09ULD8ribz3axB7jgK4Np6rniaNt7XVZ1DeZadlEBmfS0kdrMwINWmKh5PCw/132',
                            'https://wx.qlogo.cn/mmopen/vi_32/Q0j4TwGTfTKzX88Ox2CD0MksRaMTw2fdTdNUdATWEZeEKuXiam0X0kd8nhXQEPrpQicxN6CFPfVoO2910BWDhOicA/132',
                            'https://wx.qlogo.cn/mmopen/vi_32/DEiaFzfWQxaZsKWo5a7VxdPhx5ia8wysQ0lXHh8XCR5tEibQNrtK3jXJKQGpgn8jFnmAZBJPuibcN5EgUxl0D3npVw/132',
                            'https://wx.qlogo.cn/mmopen/vi_32/1uG5Rd1cjfr9bYOQ4mqtyBMib70pAM7AWRUcYKDLTMmJPnyBic3RkUvJvdqQNMUXtSSxHjkjD69DQpONsPIyQDiag/132',
                            'https://wx.qlogo.cn/mmopen/vi_32/Q0j4TwGTfTIEqPDJn7OIaJOpa0RjdW4zibeN2CMJd6kbcKBGBWWyGWGnz4SFeiaaBouJDo5ydibWibsDb7slKIib1iag/132'];
                $avatar_rand = rand(0, count($avatars) - 1);
                if (RoomLogic::getInstance()->getJoinedCount() < 4 && time() - $this->joiningRoom['createTime'] > 2) {
                    $room = RoomLogic::getInstance()->joinRoom($uid, $nicknames[$nickname_rand], $avatars[$avatar_rand]);
                    AiLogic::createAi($room['id'], $uid);
                }
            }
        }
          );
    }
}
