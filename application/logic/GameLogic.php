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
    const STATE_PROXY = 4;    //托管中
    const STATE_DEATH = 5;    //玩家死亡，游戏还未结束
    const STATE_OVER = 6;     //游戏结束

    public $pkId;
    public $state;
    public $uid;
    public $tetrisUnit;
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
        unset($data['currentShape']['shapes']);
        $data['tetrisUnit'] = $this->tetrisUnit->get_vector_boards();
        return $data;
    }

    public function createGame() {
        $this->state = GameLogic::STATE_JOIN;

        $this->tetrisUnit = new Tetris\Base();

        $this->currentShape = null;
        $this->nextShape = null;
        $this->score = 0;

        $this->timestamp = self::msectime();
    }

    public function startGame() {
        $this->state = GameLogic::STATE_INIT;

        co::sleep(3);

        $this->state = GameLogic::STATE_RUNNING;

        if ($this->uid < 10000) {
            $this->proxy = true;
        }

        $this->currentShape = $this->createShape();
        $this->nextShape = $this->createShape();

        $this->score = 0;

        $this->timestamp = self::msectime();

        go(function () {
            $this->tickUpdate();
        }
          );
    }

    public function tickUpdate() {
        while (1) {
            if (!$this->proxy) {
                break;
            }

            $finish = $this->bean()->updateGame();
            $this->getPkLogic()->printBoards();
            if ($finish) {
                break;
            }

            co::sleep(0.1);
        }
    }

    public function updateGame() {
        $finish = false;
        $now = self::msectime();
        if ( $now - $this->timestamp > $this->speed ) {
            if ( $this->currentShape != null ) {
                $finish = $this->bean()->doAction(Tetris\Base::ACTION_DOWN);
            }
            $this->timestamp = $now;
        }

        return $finish;
    }

    public function action($cmd) {
        $finish = $this->bean()->doAction($cmd);
        $this->getPkLogic()->printBoards();
        return $finish;
    }

    public function doAction($cmd) {
        $touch_down = false;

        $action_ret = $this->tetrisUnit->detect_action($cmd, $this->currentShape);
        if ($cmd != Tetris\Base::ACTION_UP) {
            if ($action_ret === false) {
                if ($cmd == Tetris\Base::ACTION_DOWN) {
                    $touch_down = true;
                }
            } else {
                $this->currentShape = $action_ret;
            }
        } else {
            $this->currentShape["y"] = $action_ret;
            $touch_down = true;
        }

        if ($touch_down) {
            if ($this->touchDown()) {
                $this->state = GameLogic::STATE_DEATH;
                return true;
            }
        }

        return false;
    }

    public function shapeDown($x, $y, $shapeIdx, $idx) {
        if ($this->currentShape["shapeIdx"] != $shapeIdx) {
            return false;
        }

        $this->currentShape["x"] = $x;
        $this->currentShape["y"] = $y;
        $this->currentShape["idx"] = $idx;

        $shapeArr = $this->currentShape["shapes"][$idx];
        if (!$this->tetrisUnit->check_available($x, $y, $shapeArr)) {
            return false;
        }

        $finish = $this->touchDown();
        if ($finish) {
            $this->state = GameLogic::STATE_DEATH;
            $winner = $this->getPkLogic()->isGameOver();
            if (!empty($winner)) {
                return ['state' => GameLogic::STATE_OVER, 'winner' => $winner];
            } else {
                return ['state' => GameLogic::STATE_DEATH];
            }
        } else {
            return ['state' => $this->state, 'currentShape' => $this->currentShape, 'nextShape' => $this->nextShape];
        }
    }

    private function touchDown() {
        $tx = $this->currentShape["x"];
        $ty = $this->currentShape["y"];
        $shapeArr = $this->currentShape["shapes"][$this->currentShape["idx"]];
        $eliminatedLines = $this->tetrisUnit->touch_down($tx, $ty, $shapeArr);

        $this->updateScore($eliminatedLines);

        $this->currentShape = $this->nextShape;
        $this->nextShape = $this->createShape();

        if ($this->ifDeath()) {
            return true;
        }
        return false;
    }

    private function ifDeath() {
        $shapeArr = $this->currentShape["shapes"][$this->currentShape["idx"]];
        return $this->tetrisUnit->is_overlay($this->currentShape["x"], $this->currentShape["y"], $shapeArr);
    }

    public function createShape() {
        $shapeTypes = [Tetris\Shape::$LShape, Tetris\Shape::$JShape, Tetris\Shape::$IShape, Tetris\Shape::$OShape, Tetris\Shape::$TShape, Tetris\Shape::$SShape, Tetris\Shape::$ZShape];

        $shapeIdx = rand(0, 100) % count($shapeTypes);
        $shapePos = rand(0, 100) % 4;
        $shape = $shapeTypes[$shapeIdx];

        return ["x" => 4, "y" => 0, "shapeIdx" => $shapeIdx, "idx" => $shapePos, "shapes" => $shape];
    }

    public function getCurrentShape() {
        return $this->currentShape;
    }

    public function getTetrisUnit() {
        return $this->tetrisUnit;
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

    public function getBoardsWithShape() {
        $boards = $this->tetrisUnit->boards;

        $tx = $this->currentShape["x"];
        $ty = $this->currentShape["y"];
        $idx = $this->currentShape["idx"];
        $shapeArr = $this->currentShape["shapes"][$idx];

        for ($i = 0; $i < 4; $i++) {
            for ($j = 0; $j < 4; $j++) {
                if ($shapeArr[$i][$j] == 1) {
                    $boards[$ty+$i][$tx+$j] = 1;
                }
            }
        }

        return $boards;
    }

    private function updateScore($line) {
        switch ($line) {
        case 1:
            $this->score += 100;
            break;
        case 2:
            $this->score += 300;
            break;
        case 3:
            $this->score += 500;
            break;
        case 4:
            $this->score += 800;
            break;
        default:
            ;
        }
    }

    public static function msectime() {
        list($msec, $sec) = explode(' ', microtime());
        $msectime = (float)sprintf('%.0f', (floatval($msec) + floatval($sec)) * 1000);
        return $msectime;
    }
}
