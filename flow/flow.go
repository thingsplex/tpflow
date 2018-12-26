package flow

import (
	log "github.com/Sirupsen/logrus"
	"github.com/alivinco/tpflow/connector"
	"github.com/alivinco/tpflow/model"
	"github.com/alivinco/tpflow/node"
	"github.com/pkg/errors"
	"runtime/debug"
	"sync"
	"time"
)

type Flow struct {
	Id             string
	Name           string
	Description    string
	FlowMeta       *model.FlowMeta
	globalContext  *model.Context
	opContext      model.FlowOperationalContext
	currentNodeIds []model.NodeID
	nodes          []model.Node
	//nodeOutboundStream  chan model.ReactorEvent
	//activeSubscriptions []string
	TriggerCounter    int64
	ErrorCounter      int64
	StartedAt         time.Time
	WaitingSince      time.Time
	LastExecutionTime time.Duration
	logFields         log.Fields
	connectorRegistry *connector.Registry
	subflowsCounter   int
	mtx               sync.Mutex
}

func NewFlow(metaFlow model.FlowMeta, globalContext *model.Context) *Flow {
	flow := Flow{globalContext: globalContext}
	flow.nodes = make([]model.Node, 0)
	flow.currentNodeIds = make([]model.NodeID, 1)
	flow.globalContext = globalContext
	flow.opContext = model.FlowOperationalContext{NodeIsReady: make(chan bool), NodeControlSignalChannel: make(chan int), State: "LOADED"}
	flow.initFromMetaFlow(&metaFlow)
	flow.subflowsCounter = 0
	flow.mtx = sync.Mutex{}

	return &flow
}

func (fl *Flow) getLog() *log.Entry {
	return log.WithFields(fl.logFields)
}

func (fl *Flow) SetStoragePath(path string) {
	fl.opContext.StoragePath = path
}

func (fl *Flow) SetExternalLibsDir(path string) {
	fl.opContext.ExtLibsDir = path
}

func (fl *Flow) initFromMetaFlow(meta *model.FlowMeta) {
	fl.Id = meta.Id
	fl.Name = meta.Name
	fl.Description = meta.Description
	fl.FlowMeta = meta
	fl.opContext.FlowId = meta.Id
	//fl.localMsgInStream = make(map[model.NodeID]model.MsgPipeline,10)
	fl.logFields = log.Fields{"fid": fl.Id, "comp": "flow"}
	fl.globalContext.RegisterFlow(fl.Id)
}

// LoadAndConfigureAllNodes creates all nodes objects from FlowMeta definitions and configures node inbound streams .
func (fl *Flow) LoadAndConfigureAllNodes() {
	defer func() {
		if r := recover(); r != nil {
			fl.getLog().Error(" Flow process CRASHED with error while doing node configuration : ", r)
			debug.PrintStack()
			fl.opContext.State = "CONFIG_ERROR"
		}
	}()
	fl.getLog().Infof(" ---------Initializing Flow Id = %s , Name = %s -----------", fl.Id, fl.Name)
	var err error
	for _, metaNode := range fl.FlowMeta.Nodes {
		if !fl.IsNodeValid(&metaNode) {
			fl.getLog().Errorf(" Node %s contains invalid configuration parameters ",metaNode.Label)
			fl.opContext.State = "CONFIG_ERROR"
			return
		}
		newNode := fl.GetNodeById(metaNode.Id)
		if newNode == nil {
			fl.getLog().Infof(" Loading node NEW . Type = %s , Label = %s", metaNode.Type, metaNode.Label)
			constructor, ok := node.Registry[metaNode.Type]
			if ok {
				newNode = constructor(&fl.opContext, metaNode, fl.globalContext)
				newNode.SetConnectorRegistry(fl.connectorRegistry)
				err = newNode.LoadNodeConfig()
				if err != nil {
					fl.getLog().Errorf(" Node type %s can't be loaded . Error : %s", metaNode.Type, err)
					fl.opContext.State = "CONFIG_ERROR"
					return
				}
				fl.AddNode(newNode)
			} else {
				fl.getLog().Errorf(" Node type = %s isn't supported. Node is skipped", metaNode.Type)
				continue
			}
		} else {
			fl.getLog().Infof(" Reusing existing node ")
		}

		fl.getLog().Info(" Running Init() function of the node")
		newNode.Init()
		fl.getLog().Info(" Done")
		if newNode.IsStartNode() {
			newNode.SetFlowRunner(fl.Run)
			go newNode.WaitForEvent(nil)
		}
		fl.getLog().Info(" Node is loaded and added.")

	}
	fl.opContext.State = "CONFIGURED"
}

func (fl *Flow) GetContext() *model.Context {
	return fl.globalContext
}

//func (fl*Flow) GetCurrentMessage()*model.Message {
//	//return &fl.currentMsg
//}

func (fl *Flow) SetNodes(nodes []model.Node) {
	fl.nodes = nodes
}

func (fl *Flow) ReloadNodes(nodes []model.Node) {
	fl.Stop()
	fl.nodes = nodes
	fl.Start()
}

func (fl *Flow) GetNodeById(id model.NodeID) model.Node {
	for i := range fl.nodes {
		if fl.nodes[i].GetMetaNode().Id == id {
			return fl.nodes[i]
		}
	}
	return nil
}

func (fl *Flow) GetFlowStats() *model.FlowStatsReport {
	stats := model.FlowStatsReport{}
	currentNode := fl.GetNodeById(fl.currentNodeIds[0])
	if currentNode != nil {
		stats.CurrentNodeId = currentNode.GetMetaNode().Id
		stats.CurrentNodeLabel = currentNode.GetMetaNode().Label
	}
	var numberOfActiveTriggers = 0
	var numberOfTrigger = 0
	for i := range fl.nodes {
		if fl.nodes[i].IsStartNode() {
			numberOfTrigger++
			if fl.nodes[i].IsReactorRunning() {
				numberOfActiveTriggers++
			}
		}
	}
	stats.NumberOfNodes = len(fl.nodes)
	stats.NumberOfActiveSubflows = fl.subflowsCounter
	stats.NumberOfTriggers = numberOfTrigger
	stats.NumberOfActiveTriggers = numberOfActiveTriggers
	stats.StartedAt = fl.StartedAt
	stats.WaitingSince = fl.WaitingSince
	stats.LastExecutionTime = int64(fl.LastExecutionTime / time.Millisecond)
	return &stats
}

func (fl *Flow) AddNode(node model.Node) {
	fl.nodes = append(fl.nodes, node)
}

func (fl *Flow) IsNodeIdValid(currentNodeId model.NodeID, transitionNodeId model.NodeID) bool {
	if transitionNodeId == "" {
		return true
	}

	if currentNodeId == transitionNodeId {
		fl.getLog().Error(" Transition node can't be the same as current")
		return false
	}
	for i := range fl.nodes {
		if fl.nodes[i].GetMetaNode().Id == transitionNodeId {
			return true
		}
	}
	fl.getLog().Error(" Transition node doesn't exist")
	return false
}

func (fl *Flow) IsNodeValid(node *model.MetaNode) bool {
	//var flowHasStartNode bool
	//for i := range fl.nodes {
	//	node := fl.nodes[i].GetMetaNode()
		if node.Type == "trigger" || node.Type == "action" || node.Type == "receive" {
			if node.Address == "" || node.ServiceInterface == "" || node.Service == "" {
				fl.getLog().Error(" Flow is not valid , node is not configured . Node ", node.Label)
				return false
			}
		}
		//if fl.nodes[i].IsStartNode() {
		//	flowHasStartNode = true
		//}
	//}
	//if !flowHasStartNode {
	//	fl.getLog().Error(" Flow is not valid, start node not found")
	//	return false
	//}
	return true
}

// Invoked by trigger node in it's own goroutine
func (fl *Flow) Run(reactorEvent model.ReactorEvent) {
	fl.mtx.Lock()
	fl.subflowsCounter++
	fl.mtx.Unlock()
	var transitionNodeId model.NodeID
	defer func() {
		fl.mtx.Lock()
		fl.subflowsCounter--
		fl.mtx.Unlock()
		if r := recover(); r != nil {
			fl.getLog().Error(" Flow process CRASHED with error : ", r)
			fl.getLog().Errorf(" Crashed while processing message from Current Node = %v Next Node = %v ", fl.currentNodeIds[0], transitionNodeId)
			transitionNodeId = ""
		}
		fl.LastExecutionTime = time.Since(fl.StartedAt)
		fl.getLog().Infof(" ------Flow %s completed ----------- ", fl.Name)

	}()

	fl.getLog().Infof(" ------Flow %s started ----------- ", fl.Name)
	fl.StartedAt = time.Now()
	//fl.getLog().Debug(" msg.payload : ",fl.currentMsg)
	if reactorEvent.Err != nil {
		fl.getLog().Error(" TriggerNode failed with error :", reactorEvent.Err)
		fl.currentNodeIds[0] = ""
	}

	fl.TriggerCounter++
	currentMsg := reactorEvent.Msg

	//fl.currentNodeId = fl.nodes[i].GetMetaNode().Id
	transitionNodeId = reactorEvent.TransitionNodeId
	fl.getLog().Debug(" Next node id = ", transitionNodeId)
	//fl.getLog().Debug(" Current nodes = ",fl.currentNodeIds)
	if !fl.IsNodeIdValid(fl.currentNodeIds[0], transitionNodeId) {
		fl.getLog().Errorf(" Unknown transition node %s from first node.Switching back to first node", transitionNodeId)
		transitionNodeId = ""
		return
	}
	var nodeOutboundStream chan model.ReactorEvent
	for {
		if !fl.opContext.IsFlowRunning {
			break
		}
		for i := range fl.nodes {
			if !fl.opContext.IsFlowRunning {
				break
			}
			if fl.nodes[i].GetMetaNode().Id == transitionNodeId {
				var err error
				var nextNodes []model.NodeID
				fl.currentNodeIds[0] = fl.nodes[i].GetMetaNode().Id
				if fl.nodes[i].IsMsgReactorNode() {
					// lazy channel init
					if nodeOutboundStream == nil {
						nodeOutboundStream = make(chan model.ReactorEvent)
					}
					if !fl.nodes[i].IsReactorRunning() {
						go fl.nodes[i].WaitForEvent(nodeOutboundStream)
					}
					// Blocking wait
					select {
					case reactorEvent := <-nodeOutboundStream:
						fl.getLog().Debug(" New event from reactor node.")
						currentMsg = reactorEvent.Msg
						transitionNodeId = reactorEvent.TransitionNodeId
						err = reactorEvent.Err
					case signal := <-fl.opContext.NodeControlSignalChannel:
						fl.getLog().Debug("Control signal ")
						if signal == model.SIGNAL_STOP {
							return
						}
					}

				} else {
					nextNodes, err = fl.nodes[i].OnInput(&currentMsg)

					if len(nextNodes) > 0 {
						transitionNodeId = nextNodes[0]
					} else {
						transitionNodeId = ""
					}
				}

				if err != nil {
					fl.ErrorCounter++
					fl.getLog().Errorf(" Node executed with error . Doing error transition to %s. Error : %s", transitionNodeId, err)
				}

				if !fl.IsNodeIdValid(fl.currentNodeIds[0], transitionNodeId) {
					fl.getLog().Errorf(" Unknown transition node %s.Switching back to first node", transitionNodeId)
					transitionNodeId = ""
				}
				fl.getLog().Debug(" Next node id = ", transitionNodeId)

			} else if transitionNodeId == "" {
				// Flow is finished . Returning to first step.
				fl.currentNodeIds[0] = ""
				return
			}
		}

	}
	//fl.opContext.State = "STOPPED"
	fl.getLog().Infof(" Runner for flow %s stopped.", fl.Name)

}

// Starts Flow loop in its own goroutine and sets isFlowRunning flag to true
// Init sequence : STARTING -> RUNNING , STATING -> NOT_CONFIGURED ,
func (fl *Flow) Start() error {
	if fl.GetFlowState() == "RUNNING" {
		log.Info("Flow is already running")
		return nil
	}
	fl.getLog().Info(" Starting flow : ", fl.Name)
	fl.opContext.State = "STARTING"
	fl.opContext.IsFlowRunning = true
	fl.LoadAndConfigureAllNodes()
	if fl.opContext.State == "CONFIGURED" {
		// Init all nodes
		//for i := range fl.nodes {
		//	fl.nodes[i].Init()
		//}
		fl.opContext.State = "RUNNING"
		fl.opContext.IsFlowRunning = true
		fl.getLog().Infof(" Flow %s is running", fl.Name)
	} else {
		fl.opContext.State = "NOT_CONFIGURED"
		fl.getLog().Errorf(" Flow %s is not valid and will not be started.Flow should have at least one trigger or wait node ", fl.Name)
		return errors.New("Flow should have at least one trigger or wait node")
	}
	return nil
}

// Terminates flow loop , stops goroutine .
func (fl *Flow) Stop() error {
	if fl.GetFlowState() == "STOPPING" {
		log.Info("Flow is already stopping")
		return nil
	}
	fl.getLog().Info(" Stopping flow  ", fl.Name)
	fl.opContext.IsFlowRunning = false
	fl.opContext.State = "STOPPING"
	var breakLoop = false
	for {
		// Sending STOP signal to all active triggers
		fl.getLog().Debug(" Sending STOP signals to all reactors")
		select {
		case fl.opContext.NodeControlSignalChannel <- model.SIGNAL_STOP:
		default:
			breakLoop = true
		}
		if breakLoop {
			break
		}

	}
	var isSomeNodeRunning bool
	//Check if all triggers has stopped
	for {
		isSomeNodeRunning = false
		for i := range fl.nodes {
			if fl.nodes[i].IsMsgReactorNode() {
				if fl.nodes[i].IsReactorRunning() {
					isSomeNodeRunning = true
					break
				}

			}
		}
		if isSomeNodeRunning {
			fl.getLog().Debug(" Some reactors are still running . Waiting.....")
			time.Sleep(time.Second * 2)
		} else {
			break
		}
	}

	// Wait until all subflows are stopped

	for {
		if fl.subflowsCounter == 0 {
			break
		} else {
			fl.getLog().Debug(" Some subflows are still running . Waiting.....")
			time.Sleep(time.Second * 2)
		}
	}

	fl.getLog().Debug(" Starting node cleanup")
	for i := range fl.nodes {
		fl.nodes[i].Cleanup()
	}
	fl.getLog().Debug(" nodes cleanup completed")
	fl.getLog().Info(" Stopped .  ", fl.Name)
	fl.opContext.State = "STOPPED"
	return nil
}

func (fl *Flow) CleanupBeforeDelete() {
	if fl.GetFlowState() == "LOADED" {
		fl.getLog().Info(" Nothing to cleanup ")
		return
	}
	fl.getLog().Info(" All streams and running goroutins were closed  ")
	fl.globalContext.UnregisterFlow(fl.Id)
}

func (fl *Flow) GetFlowState() string {
	return fl.opContext.State
}

func (fl *Flow) IsNodeCurrentNode(nodeId model.NodeID) bool {
	for i := range fl.currentNodeIds {
		if fl.currentNodeIds[i] == nodeId {
			return true
		}
	}
	return false
}

func (fl *Flow) SetConnectorRegistry(resources *connector.Registry) {
	fl.connectorRegistry = resources
}
