package flow

import (
	log "github.com/Sirupsen/logrus"
	"github.com/alivinco/fimpgo"
	"github.com/alivinco/tpflow/model"
	"github.com/alivinco/tpflow/node"
	"github.com/alivinco/tpflow/shared"
	"github.com/alivinco/tpflow/utils"
	"github.com/pkg/errors"
	"time"
)

type Flow struct {
	Id                  string
	Name                string
	Description         string
	FlowMeta            *model.FlowMeta
	globalContext       *model.Context
	opContext           model.FlowOperationalContext
	currentNodeIds      [] model.NodeID
	currentMsg          model.Message
	nodes               []model.Node
	nodeInboundStreams  map[model.NodeID]model.MsgPipeline
	nodeOutboundStream  chan model.ReactorEvent
	msgTransport        *fimpgo.MqttTransport
	activeSubscriptions []string
	msgInStream         model.MsgPipeline
	TriggerCounter      int64
	ErrorCounter        int64
	StartedAt           time.Time
	WaitingSince        time.Time
	LastExecutionTime   time.Duration
	logFields           log.Fields
 	sharedResources     * shared.GlobalSharedResources
}

func NewFlow(metaFlow model.FlowMeta, globalContext *model.Context, msgTransport *fimpgo.MqttTransport) *Flow {
	flow := Flow{globalContext: globalContext}

	flow.nodes = make([]model.Node, 0)
	flow.currentNodeIds = make([]model.NodeID,1)
	flow.nodeInboundStreams = make(map[model.NodeID]model.MsgPipeline)
	flow.msgTransport = msgTransport
	flow.globalContext = globalContext
	flow.opContext = model.FlowOperationalContext{NodeIsReady:make(chan bool),NodeControlSignalChannel:make(chan int),State:"LOADED"}
	flow.initFromMetaFlow(&metaFlow)

	return &flow
}

func (fl *Flow) getLog() *log.Entry {
	return log.WithFields(fl.logFields)
}

func (fl *Flow) SetStoragePath(path string) {
	fl.opContext.StoragePath = path
}

func (fl *Flow) initFromMetaFlow(meta *model.FlowMeta) {
	fl.Id = meta.Id
	fl.Name = meta.Name
	fl.Description = meta.Description
	fl.FlowMeta = meta
	fl.opContext.FlowId = meta.Id
	//fl.localMsgInStream = make(map[model.NodeID]model.MsgPipeline,10)
	fl.logFields = log.Fields{"fid":fl.Id,"comp":"flow"}
	fl.globalContext.RegisterFlow(fl.Id)
}

// LoadAndConfigureAllNodes creates all nodes objects from FlowMeta definitions and configures node inbound streams .
func (fl *Flow) LoadAndConfigureAllNodes() {
	defer func() {
		if r := recover(); r != nil {
			fl.getLog().Error(" Flow process CRASHED with error while doing node configuration : ",r)
			fl.opContext.State = "INIT_FAIL"
		}
	}()
	fl.getLog().Infof(" ---------Initializing Flow Id = %s , Name = %s -----------",fl.Id,fl.Name)
	fl.nodeOutboundStream = make(chan model.ReactorEvent)
	var isNewNode = false
	for _,metaNode := range fl.FlowMeta.Nodes {
		newNode := fl.GetNodeById(metaNode.Id)
		if newNode == nil {
			isNewNode = true
			fl.getLog().Infof(" Loading node NEW . Type = %s , Label = %s",metaNode.Type,metaNode.Label)
			constructor ,ok := node.Registry[metaNode.Type]
			if ok {
				newNode = constructor(&fl.opContext,metaNode,fl.globalContext,fl.msgTransport)
			}else {
				fl.getLog().Errorf(" Node type = %s isn't supported",metaNode.Type)
				continue
			}
		}else {
			fl.getLog().Infof(" Reusing existing node ")
		}
		if newNode.IsMsgReactorNode() {
			fl.getLog().Debug(" Creating a channel for Reactor-Node id = ",metaNode.Id)
			// Each reactor node gets its own channel.
			fl.nodeInboundStreams[metaNode.Id] = make(model.MsgPipeline)
			newNode.ConfigureInStream(&fl.activeSubscriptions,fl.nodeInboundStreams[metaNode.Id])
		}
		newNode.SetSharedResources(fl.sharedResources)
		err := newNode.LoadNodeConfig()
		if err == nil && isNewNode {
			fl.AddNode(newNode)
			fl.getLog().Info(" Node is loaded and added.")
		}else {
			fl.getLog().Errorf(" Node type %s can't be loaded . Error : %s",metaNode.Type,err)
		}
	}
}

func (fl*Flow) GetContext()*model.Context {
	return fl.globalContext
}

func (fl*Flow) GetCurrentMessage()*model.Message {
	return &fl.currentMsg
}

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

func (fl *Flow)GetFlowStats() (*model.FlowStatsReport) {
	stats := model.FlowStatsReport{}
	currentNode := fl.GetNodeById(fl.currentNodeIds[0])
	if currentNode != nil {
		stats.CurrentNodeId = currentNode.GetMetaNode().Id
		stats.CurrentNodeLabel = currentNode.GetMetaNode().Label
		stats.IsAtStartingPoint = currentNode.IsStartNode()
	}
	stats.StartedAt = fl.StartedAt
	stats.WaitingSince = fl.WaitingSince
	stats.LastExecutionTime = int64(fl.LastExecutionTime/time.Millisecond)
	return &stats
}


func (fl *Flow) AddNode(node model.Node) {
	fl.nodes = append(fl.nodes, node)
}


func (fl *Flow) IsNodeIdValid(currentNodeId model.NodeID, transitionNodeId model.NodeID) bool {
	if transitionNodeId == ""{
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

func (fl *Flow) IsFlowValid() bool {
	var flowHasStartNode bool
	for i := range fl.nodes {
		node := fl.nodes[i].GetMetaNode()
		if node.Type == "trigger" || node.Type == "action" || node.Type == "receive" {
			if node.Address == "" ||  node.ServiceInterface == "" || node.Service == ""	{
				fl.getLog().Error(" Flow is not valid , node is not configured . Node ",node.Label)
				return false
			}
		}
		if fl.nodes[i].IsStartNode() {
			flowHasStartNode = true
		}
	}
	if !flowHasStartNode {
		fl.getLog().Error(" Flow is not valid, start node not found")
		return false
	}
	return true
}

func (fl *Flow) Run() {
	var transitionNodeId model.NodeID
	defer func() {
		if r := recover(); r != nil {
			fl.getLog().Error(" Flow process CRASHED with error : ",r)
			fl.getLog().Errorf(" Crashed while processing message from Current Node = %v Next Node = %v ",fl.currentNodeIds[0], transitionNodeId)
			transitionNodeId = ""
		}
	}()

	for {
		if !fl.opContext.IsFlowRunning {
			break
		}
		for i := range fl.nodes {
			if !fl.opContext.IsFlowRunning {
				break
			}
			if fl.currentNodeIds[0] == "" && fl.nodes[i].IsStartNode() {
				fl.LastExecutionTime = time.Since(fl.StartedAt)
				fl.getLog().Infof(" ------Flow %s is waiting for event ----------- ",fl.Name)
				// Initial message received by trigger , which is passed further throughout the flow.
				fl.WaitingSince = time.Now()
				// Starting all Start nodes and waiting for event from one of them
				fl.currentNodeIds = fl.currentNodeIds[:0]
				for si := range fl.nodes {
					if fl.nodes[si].IsStartNode() {
						if ! fl.nodes[si].IsReactorRunning(){
							go fl.nodes[si].WaitForEvent(fl.nodeOutboundStream)
						}
						fl.currentNodeIds = append(fl.currentNodeIds,fl.nodes[si].GetMetaNode().Id )
					}

				}
				// Blocking wait
				reactorEvent :=<- fl.nodeOutboundStream
				fl.getLog().Debug(" New event from reactor node.")
				fl.StartedAt = time.Now()
				fl.currentMsg = reactorEvent.Msg
				//fl.getLog().Debug(" msg.payload : ",fl.currentMsg)
				if reactorEvent.Err != nil {
					fl.getLog().Error(" TriggerNode failed with error :", reactorEvent.Err)
					fl.currentNodeIds[0] = ""
				}
				if !fl.opContext.IsFlowRunning {
					break
				}
				fl.TriggerCounter++
				//fl.currentNodeId = fl.nodes[i].GetMetaNode().Id
				transitionNodeId = reactorEvent.TransitionNodeId
				fl.getLog().Debug(" Next node id = ",transitionNodeId)
				//fl.getLog().Debug(" Current nodes = ",fl.currentNodeIds)
				if !fl.IsNodeIdValid(fl.currentNodeIds[0], transitionNodeId) {
					fl.getLog().Errorf(" Unknown transition node %s from first node.Switching back to first node", transitionNodeId)
					transitionNodeId = ""
				}
				//fl.getLog().Debug(" Transition from Trigger to node = ", transitionNodeId)
			} else if fl.nodes[i].GetMetaNode().Id == transitionNodeId {
				var err error
				var nextNodes []model.NodeID
				fl.currentNodeIds[0] = fl.nodes[i].GetMetaNode().Id
				if fl.nodes[i].IsMsgReactorNode() {
					if ! fl.nodes[i].IsReactorRunning(){
						go fl.nodes[i].WaitForEvent(fl.nodeOutboundStream)
					}
					// Blocking wait
					reactorEvent :=<- fl.nodeOutboundStream
					fl.currentMsg = reactorEvent.Msg
					transitionNodeId = reactorEvent.TransitionNodeId
					err = reactorEvent.Err
					fl.getLog().Debug(" New event from reactor node.")
					//fl.getLog().Debug(" msg.payload : ",fl.currentMsg)

				}else {
					nextNodes, err = fl.nodes[i].OnInput(&fl.currentMsg)
					if len(nextNodes)>0 {
						transitionNodeId = nextNodes[0]
					}else {
						transitionNodeId = ""
					}
				}

				if err != nil {
					fl.ErrorCounter++
					fl.getLog().Errorf(" Node executed with error . Doing error transition to %s. Error : %s", transitionNodeId,err)
				}

				if !fl.IsNodeIdValid(fl.currentNodeIds[0], transitionNodeId) {
					fl.getLog().Errorf(" Unknown transition node %s.Switching back to first node", transitionNodeId)
					transitionNodeId = ""
				}
				fl.getLog().Debug(" Next node id = ",transitionNodeId)

			} else if transitionNodeId == "" {
				// Flow is finished . Returning to first step.
				fl.currentNodeIds[0] = ""
			}
		}

	}
	fl.opContext.State = "STOPPED"
	fl.getLog().Infof(" Runner for flow %s stopped.", fl.Name)

}

func (fl *Flow) IsFlowInterestedInMessage(topic string ) bool {
	for i :=range fl.activeSubscriptions {
		if utils.RouteIncludesTopic(fl.activeSubscriptions[i],topic) {
			return true
		}else {
			//fl.getLog().Debug(" Not interested in topic : ",topic)
		}
	}
	return false
}

func (fl *Flow) CloseAllInboundStreams() {
	for _,stream := range fl.nodeInboundStreams {
		cancelMsg := model.Message{CancelOp:true}
		select {
		case stream <- cancelMsg:
			continue
		default:
			continue
		}
		close(stream)
	}
	fl.getLog().Debug(" All node inbound streams were closed")
}

// Starts Flow loop in its own goroutine and sets isFlowRunning flag to true
func (fl *Flow) Start() error {
	if fl.GetFlowState() == "RUNNING" {
		log.Info("Flow is already running")
		return nil
	}
	fl.getLog().Info(" Starting flow : ", fl.Name)
	fl.opContext.State = "STARTING"
	fl.opContext.IsFlowRunning = true
	fl.LoadAndConfigureAllNodes()
	isFlowValid := fl.IsFlowValid()
	if isFlowValid{
		// Init all nodes
		for i := range fl.nodes {
			fl.nodes[i].Init()
		}
		fl.opContext.State = "RUNNING"
		fl.StartMsgStreamRouter()
		fl.getLog().Infof(" Flow %s is running", fl.Name)
		go fl.Run()
	}else {
		fl.opContext.State = "NOT_CONFIGURED"
		fl.getLog().Errorf(" Flow %s is not valid and will not be started.Flow should have at least one trigger or wait node ",fl.Name)
		return errors.New("Flow should have at least one trigger or wait node")
	}
	return nil
}

// Terminates flow loop , stops goroutine .
func (fl *Flow) Stop() error {
	fl.getLog().Info(" Stopping flow  ", fl.Name)
	// is invoked when node flow is stopped
	for _,topic := range fl.activeSubscriptions {
		fl.getLog().Info(" Unsubscribing from topic : ",topic)
		fl.msgTransport.Unsubscribe(topic)
	}
	fl.activeSubscriptions = nil
	fl.opContext.IsFlowRunning = false
	select {
	case fl.opContext.NodeControlSignalChannel <- model.SIGNAL_STOP:
	default:
		//fl.getLog().Debug(" No signal listener.")
	}

	select {
	case fl.msgInStream <- model.Message{CancelOp:true}:
	default:
		//fl.getLog().Debug(" No msgInStream.")
	}

	select {
	case fl.nodeOutboundStream <- model.ReactorEvent{}:
	default:
		//fl.getLog().Debug(" No signal listener.")
	}
	fl.getLog().Debug(" Starting node cleanup")
	for i := range fl.nodes {
		fl.nodes[i].Cleanup()
	}
	fl.getLog().Debug(" nodes cleanup completed")
	fl.getLog().Info(" Stopped .  ", fl.Name)
	return nil
}

func (fl *Flow) CleanupBeforeDelete() {
	if fl.GetFlowState() == "LOADED" {
		fl.getLog().Info(" Nothing to cleanup ")
		return
	}
	fl.CloseAllInboundStreams()
	close(fl.nodeOutboundStream)
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

func (fl *Flow) StartMsgStreamRouter() {
	// Message broadcast from flow incomming stream to reactor nodes
	go func() {
		fl.getLog().Info(" !!! Starting flow msg router !!! ")
		defer func() {
			fl.getLog().Info(" !!! Router is stopped !!! ")
		}()
		for msg := range fl.msgInStream {
			for nodeId,stream := range fl.nodeInboundStreams {
				if fl.IsNodeCurrentNode(nodeId){
					fl.getLog().Debug("Router got new msg , forwarding it to node = ", nodeId)
					select {
					case stream <- msg:
					default:
						fl.getLog().Debug(" Message is dropped (no listeners) nodeId = ", nodeId)
					}
				}
			}
			if msg.CancelOp {
				return
			}

		}
	}()

}

func (fl *Flow) SetMessageStream(msgInStream model.MsgPipeline) {
	fl.msgInStream = msgInStream
}

func (fl *Flow) SetSharedResources(resources *shared.GlobalSharedResources) {
	fl.sharedResources = resources
}
