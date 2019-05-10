<?php if ( ! defined('BASEPATH')) exit('No direct script access allowed');

/**
 * ExampleModel Class
 *
 * @package			Ycsocket
 * @subpackage		Model
 * @category		Example Model
 * @author			caohao
 */
class GameLogic extends ActorBean {
    const STATE_JOIN = 1;     //加入中
    const STATE_INIT = 2;     //初始化
    const STATE_RUNNING = 3;  //游戏中
    const STATE_DEATH = 4;    //玩家死亡，游戏还未结束
    const STATE_OVER = 5;     //游戏结束

    public $pkId;
    public $state;
    public $uid;
    public $currentShape;
    public $nextShape;
    public $timestamp;
    public $score;
    public $speed = 500;
    public $proxy = false;

    public function __construct($pkId, $uid) {
        $this->pkId = $pkId;
        $this->uid = $uid;
    }

    public function getPkLogic() {
        return PkLogic::getBean($this->pkId);
    }

    public function getTetris() {
        $data = array();
        $data['state'] = $this->state;
        $data['score'] = $this->score;
        $data['speed'] = $this->speed;
        $data['currentShape'] = $this->currentShape;
        $data['nextShape'] = $this->nextShape;
        unset($data['currentShape']['shapes']);
        unset($data['nextShape']['shapes']);
        return $data;
    }

    public function createGame() {
        $this->state = GameLogic::STATE_JOIN;

        $this->currentShape = null;
        $this->nextShape = null;
        $this->score = 0;

        $this->timestamp = self::msectime();
    }

    public function startGame() {
        $this->state = GameLogic::STATE_INIT;

        co::sleep(3);

        echo "------running-----\n";

        $this->state = GameLogic::STATE_RUNNING;

        if ($this->uid < 10000) {
            $this->proxy = true;
        }

        $this->currentShape = $this->createShape();
        $this->nextShape = $this->createShape();

        $ret = ['c' => 'tt', 'm' => 'start', 'pkid' => $this->pkId, 'currentShape' => $this->currentShape, 'nextShape' => $this->nextShape];
        unset($ret['currentShape']['shapes']);
        unset($ret['nextShape']['shapes']);
        Userfd::getInstance()->send($this->uid, $ret);

        $this->score = 0;

        $this->timestamp = self::msectime();


        go(function () {
            $this->tickUpdate();
        });
    }

    public function tickUpdate() {
        while (1) {
            if (!$this->proxy && $this->state != GameLogic::STATE_RUNNING) {
                break;
            }

            $now = self::msectime();
            if ( $now - $this->timestamp > $this->speed ) {
                $finish = $this->bean()->action("down");
                $this->timestamp = $now;
                if ($finish) {
                    break;
                }
            }

            co::sleep(0.1);
        }
    }

    public function action($cmd) {
        $this->getPkLogic()->printBoards();

        echo time() . " | $cmd | do action\n";

        $this->currentShape = $this->nextShape;
        $this->nextShape = $this->createShape();

        return false;
    }

    private function createShape() {
        $shapeTypes = ["LShape", "JShape", "IShape", "OShape", "TShape", "SShape", "ZShape"];

        $shapeIdx = rand(0, 100) % count($shapeTypes);
        $shapePos = rand(0, 100) % 4;
        $shape = $shapeTypes[$shapeIdx];

        return ["x" => 4, "y" => 0, "shapeIdx" => $shapeIdx, "idx" => $shapePos, "shapes" => $shape];
    }

    public function getCurrentShape() {
        return $this->currentShape;
    }


    public function setProxy($proxy) {
        $this->proxy = $proxy;
    }

    public function getState() {
        return $this->state;
    }

    public function setState($state) {
        $this->state = $state;
    }

    public static function msectime() {
        list($msec, $sec) = explode(' ', microtime());
        $msectime = (float)sprintf('%.0f', (floatval($msec) + floatval($sec)) * 1000);
        return $msectime;
    }
}
