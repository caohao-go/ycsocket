<?php
/**
 * Created by PhpStorm.
 * User: yf
 * Date: 2018-12-27
 * Time: 12:10
 */

class Actor
{
    use Singleton;

    protected $actorList = [];
    private $tempDir;
    private $serverName = 'TT';
    private $run = false;

    function __construct()
	{
        $this->tempDir = BASEPATH . "/actor/data";
    }
	
	public function actorCreate(string $actorClass, $args) 
	{
		if(!isset($this->actorList[$actorClass])){
			return null;
		}
		
		$actorClient = new ActorClient($this->actorList[$actorClass], $this->tempDir, $this->serverName);
		$actorId = $actorClient->new(-1, $args);
		
		if(empty($actorId)) {
			return null;
		}
		
		return $actorClient;
	}
	
	public function actorExist(string $actorClass, string $actorId, $timeout) 
	{
		if(empty($actorId)) {
			return false;
		}
		
		$actorClient = new ActorClient($this->actorList[$actorClass], $this->tempDir, $this->serverName);
		return $actorClient->exist($actorId, $timeout);
	}
	
    public function setServerName(string $serverName): Actor
    {
        $this->modifyCheck();
        $this->serverName = $serverName;
        return $this;
    }

    function setTempDir(string $dir):Actor
    {
        $this->modifyCheck();
        $this->tempDir = $dir;
        return $this;
    }

    function register(string $actorClass):ActorConfig
    {
		$config = new ActorConfig();
		ActorFactory::configure($config, $actorClass);
		if(empty($config->getActorName())){
			throw new Exception("actor name for class:{$actorClass} is required");
		}
		
		if(in_array($config->getActorName(),$this->actorList)){
			throw new Exception("actor name for class:{$actorClass} is duplicate");
		}
		
		$this->actorList[$actorClass] = $config;
		return $config;
    }
    
    function attachToServer(\swoole_server $server)
    {
        $list = $this->initProcess();
        foreach ($list as $process){
            /** @var $proces ActorProcess */
            $server->addProcess($process->getProcess());
        }
    }

    function initProcess():array
    {
        $this->run = true;
        $processList = [];
        foreach ($this->actorList as $actorClass => $config){
            /** @var $config ActorConfig */
            $subName = "ActorProcess.{$this->serverName}.{$config->getActorName()}";
            for($i = 1;$i <= $config->getActorProcessNum();$i++){
                $processConfig = new ProcessConfig();
                $processConfig->setActorClass($actorClass);
                $processConfig->setTempDir($this->tempDir);
                $finaleName = "{$subName}.{$i}";
                $processConfig->setIndex($i);
                $processConfig->setProcessName($finaleName);
                $processConfig->setBacklog($config->getBacklog());
                $processConfig->setOnStart($config->getOnStart());
                $processConfig->setOnShutdown($config->getOnShutdown());
                $processConfig->setBlock($config->isBlock());
                $processConfig->setOnTick($config->getOnTick());
                $processConfig->setTick($config->getTick());
                $processConfig->setProcessOnException($config->getProcessOnException());
                $process = new ActorProcess($finaleName,$processConfig);
                $processList[] = $process;
            }
        }
        return $processList;
    }

    private function modifyCheck()
    {
        if($this->run){
            throw new Exception('you can not modify configure after init process check');
        }
    }
}