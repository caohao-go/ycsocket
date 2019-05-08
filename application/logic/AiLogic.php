<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');

/**
 * ExampleModel Class
 *
 * @package			Ycsocket
 * @subpackage		Model
 * @category		Example Model
 * @author			caohao
 */
class AiLogic {
    public static function createAi($pkId, $uid) {
        go(function () use ($pkId, $uid) {
            co::sleep(3);
            AiLogic::tickMakeAction($pkId, $uid);
        }
          );
    }

    public static function tickMakeAction($pkId, $uid) {
        $strategy = new Tetris\Ai();
        $gameLogic = PkLogic::getBean($pkId)->getGameLogicByUid($uid);

        while (1) {
            $state = $gameLogic->getState();
            if ($state == GameLogic::STATE_DEATH || $state == GameLogic::STATE_OVER) {
                return;
            } else if ($state != GameLogic::STATE_RUNNING) {
                co::sleep(0.1);
                continue;
            }

            $moveAns = $strategy->make_best_decision($gameLogic->getTetrisUnit(), $gameLogic->getCurrentShape());
            foreach($moveAns['action_moves'] as $move) {
                if ($move['cmd'] == -1) {
                    continue;
                }

                $finish = $gameLogic->action($move['cmd']);
                if ($finish) {
                    return;
                }

                if ($move['cmd'] != Tetris\Base::ACTION_DOWN) {
                    co::sleep(0.3);
                } else {
                    co::sleep(0.01);
                }
            }
        }
    }
}
