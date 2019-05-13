<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');

use EasySwoole\Actor\ActorBean;
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


    public function isInRoom($userid) {
        $result = array();
        if (isset($this->joiningRoom['users'][$userid])) {
            $this->joiningRoom['pkLogic']->setGameProxy($userid, false);

            $result['id'] = $this->joiningRoom['id'];
            $result['createtime'] = $this->joiningRoom['createtime'];
            $result['state'] = GameLogic::STATE_JOIN;
            $result['users'] = array_values($this->joiningRoom['users']);
            return $result;
        }

        if (isset($this->joinedUser[$userid])) {
            $result['id'] = $pkid = $this->joinedUser[$userid];
            $this->playingRooms[$pkid]['pkLogic']->setGameProxy($userid, false);

            $result['createtime'] = $this->playingRooms[$pkid]['createtime'];
            $result['starttime'] = $this->playingRooms[$pkid]['starttime'];
            $result['state'] = $state = $this->playingRooms[$pkid]['pkLogic']->getState($userid);
            $result['users'] = array_values($this->playingRooms[$pkid]['users']);

            if ($state > GameLogic::STATE_INIT) {
                $result['tetris'] = $this->playingRooms[$pkid]['pkLogic']->getAllTetris();
            }

            return $result;
        }
    }

    public function joinRoom($userid) {
        $ret = array();
        if (empty($this->joiningRoom)) {
            $this->joiningRoom['pkLogic'] = PkLogic::new();
            $ret['id'] = $this->joiningRoom['id'] = $this->joinedUser[$userid] = $this->joiningRoom['pkLogic']->getActorId();
            $ret['createtime'] = $this->joiningRoom['createtime'] = time();

            $this->joiningRoom['pkLogic']->joinUser($userid);
            $this->joiningRoom['users'][$userid] = ['userid' => $userid];

            $this->waitJoin();

            $ret['state'] = GameLogic::STATE_JOIN;
            $ret['users'] = array_values($this->joiningRoom['users']);
        } else {
            $ret['id'] = $pkid = $this->joiningRoom['id'];
            $ret['createtime'] = $this->joiningRoom['createtime'];

            $this->joinedUser[$userid] = $pkid;
            $this->joiningRoom['users'][$userid] = ['userid' => $userid];

            $joinedCount = $this->joiningRoom['pkLogic']->joinUser($userid);

            if ($joinedCount >= 4) {
                $this->playingRooms[$pkid] = $this->joiningRoom;
                $this->joiningRoom = null;
                $this->playingRooms[$pkid]['starttime'] = time();
                $this->playingRooms[$pkid]['pkLogic']->startGame();

                $ret['starttime'] = $this->playingRooms[$pkid]['starttime'];
                $ret['state'] = GameLogic::STATE_INIT;
                $ret['users'] = array_values($this->playingRooms[$pkid]['users']);

                foreach($this->playingRooms[$pkid]['users'] as $user) {
                    if ($user['userid'] != $userid) {
                        Userfd::getInstance()->send($user['userid'], ['c' => 'tt', 'm' => 'join', 'data' => $ret]);
                    }
                }
            } else {
                $ret['state'] = GameLogic::STATE_JOIN;
                $ret['users'] = array_values($this->joiningRoom['users']);

                foreach($this->joiningRoom['users'] as $user) {
                    if ($user['userid'] != $userid) {
                        Userfd::getInstance()->send($user['userid'], ['c' => 'tt', 'm' => 'join', 'data' => $ret]);
                    }
                }
            }
        }

        return $ret;
    }

    public function exitRoom($pkid) {
        $this->playingRooms[$pkid]['pkLogic']->destroy();
        unset($this->playingRooms[$pkid]);
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
            });
        }
    }

    public function waitJoin() {
        go(function () {
            while (1) {
                co::sleep(1);

                if (empty($this->joiningRoom)) {
                    return;
                }

                $uid = rand(111, 9999);

                if (count($this->joiningRoom['users']) < 4 && time() - $this->joiningRoom['createtime'] > 10) {
                    $room = RoomLogic::getInstance()->joinRoom($uid);
                    AiLogic::createAi($room['id'], $uid);
                }
            }
        });
    }
}
