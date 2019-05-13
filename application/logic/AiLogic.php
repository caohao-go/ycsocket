<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');

use EasySwoole\Actor\ActorBean;

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
        });
    }

    public static function tickMakeAction($pkId, $uid) {
        $gameLogic = PkLogic::getBean($pkId)->getGameLogicByUid($uid);

        while (1) {
            $state = $gameLogic->getState();
            if ($state == GameLogic::STATE_DEATH || $state == GameLogic::STATE_OVER) {
                break;
            } else if ($state != GameLogic::STATE_RUNNING) {
                co::sleep(0.1);
                continue;
            }

            $cmds = ["left", "right", "down", "up"];
            $finish = $gameLogic->action($cmds[rand(0, 3)]);
            co::sleep(0.301);
            if($finish) {
                break;
            }
        }
    }
}
