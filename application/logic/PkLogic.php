<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');

/**
 * ExampleModel Class
 *
 * @package			Ycsocket
 * @subpackage		Model
 * @category		Example Model
 * @author			caohao
 */
class PkLogic extends ActorBean {
    private $gameLogics = array();

    public function __construct() {
    }

    public function joinUser($uid) {
        $this->gameLogics[$uid] = GameLogic::new($this->actorId, $uid);
        $this->gameLogics[$uid]->createGame();

        return count($this->gameLogics);
    }

    public function isGameOver() {
        $aliveCount = 0;
        $states = array();
        $winner = 0;
        foreach($this->gameLogics as $uid => $gameLogic) {
            $states[$uid] = $gameLogic->getState();
            if ($states[$uid] <= GameLogic::STATE_RUNNING) {
                $aliveCount++;
                $winner = $uid;
            }

            if ($aliveCount > 1) {
                return false;
            }
        }

        return $winner;
    }

    public function startGame() {
        foreach($this->gameLogics as $gameLogics) {
            go(function () use ($gameLogics) {
                $gameLogics->startGame();
            });
        }
    }

    public function getState($uid) {
        return $this->gameLogics[$uid]->getState();
    }

    public function getGameLogicByUid($uid) {
        return $this->gameLogics[$uid];
    }

    public function getAllTetris() {
        $ret = array();
        foreach($this->gameLogics as $uid => $gameLogics) {
            $ret[$uid] = $gameLogics->getTetris();
        }
        return $ret;
    }

    public function setGameProxy($uid, $proxy) {
        $this->gameLogics[$uid]->setProxy($proxy);
    }

    function onDestroy() {
        foreach($this->gameLogics as $gameLogics) {
            $gameLogics->destroy();
        }
    }

    public function printBoards() {
        echo "======== printBoards ======\n";
    }
}
