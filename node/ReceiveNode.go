package node

import (
	"github.com/alivinco/fimpgo"
	"github.com/alivinco/tpflow/model"
	"github.com/alivinco/tpflow/utils"
	"github.com/mitchellh/mapstructure"
	"time"
)

type ReceiveNode struct {
	BaseNode
	ctx *model.Context
	transport *fimpgo.MqttTransport
	activeSubscriptions *[]string
	msgInStream model.MsgPipeline
	config ReceiveConfig

}

type ReceiveConfig struct {
	Timeout int64 // in seconds
	ValueFilter model.Variable
	InputVariableType string
	IsValueFilterEnabled bool
	RegisterAsVirtualService bool
	VirtualServiceGroup string
}

func NewReceiveNode(flowOpCtx *model.FlowOperationalContext ,meta model.MetaNode,ctx *model.Context,transport *fimpgo.MqttTransport) model.Node {
	node := ReceiveNode{ctx:ctx,transport:transport}
	node.isStartNode = false
	node.isMsgReactor = true
	node.flowOpCtx = flowOpCtx
	node.meta = meta
	node.config = ReceiveConfig{}
	node.SetupBaseNode()
	return &node
}

func (node *ReceiveNode) ConfigureInStream(activeSubscriptions *[]string,msgInStream model.MsgPipeline) {
	node.activeSubscriptions = activeSubscriptions
	node.msgInStream = msgInStream
	node.initSubscriptions()
}

func (node *ReceiveNode) initSubscriptions() {
	node.getLog().Info("ReceiveNode is listening for events . Name = ", node.meta.Label)
	needToSubscribe := true
	for i := range *node.activeSubscriptions {
			if (*node.activeSubscriptions)[i] == node.meta.Address {
				needToSubscribe = false
				break
			}
	}
	if needToSubscribe {
		if node.meta.Address != "" {
			node.getLog().Info("Subscribing for service by address :", node.meta.Address)
			node.transport.Subscribe(node.meta.Address)
			*node.activeSubscriptions = append(*node.activeSubscriptions, node.meta.Address)
		}else {
			node.getLog().Error("Can't subscribe to service with empty address")
		}
	}
}


func (node *ReceiveNode) LoadNodeConfig() error {
	err := mapstructure.Decode(node.meta.Config,&node.config)
	if err != nil{
		node.getLog().Error("Failed to load node configs.Err:",err)
	}
	return err
}

func (node *ReceiveNode) WaitForEvent(nodeEventStream chan model.ReactorEvent) {
	node.isReactorRunning = true
	defer func() {
		node.isReactorRunning = false
		node.getLog().Debug("Reactor-WaitForEvent is stopped ")
	}()
	node.getLog().Debug("Reactor-Waiting for event .chan size = ",len(node.msgInStream))
	start := time.Now()
	timeout := node.config.Timeout
	if timeout == 0 {
		timeout = 86400 // 24 hours
	}

	for {
		select {
		case newMsg := <-node.msgInStream:
			node.getLog().Info("New message :")
			if newMsg.CancelOp {
				return
			}
			if utils.RouteIncludesTopic(node.meta.Address,newMsg.AddressStr) &&
				(newMsg.Payload.Service == node.meta.Service || node.meta.Service == "*") &&
				(newMsg.Payload.Type == node.meta.ServiceInterface || node.meta.ServiceInterface == "*") {
				if !node.config.IsValueFilterEnabled {
					newEvent := model.ReactorEvent{Msg:newMsg,TransitionNodeId:node.meta.SuccessTransition}
					select {
					case nodeEventStream <- newEvent:
						return
					default:
						node.getLog().Debug("Message is dropped (no listeners) ")
					}

				} else if newMsg.Payload.Value == node.config.ValueFilter.Value {
					newEvent := model.ReactorEvent{Msg:newMsg,TransitionNodeId:node.meta.SuccessTransition}
					select {
					case nodeEventStream <- newEvent:
						return
					default:
						node.getLog().Debug("Message is dropped (no listeners) ")
					}
				}
			}
			if node.config.Timeout > 0 {
				elapsed := time.Since(start)
				timeout = timeout - int64(elapsed.Seconds())
			}
			node.getLog().Debug("Not interested .")

		case <-time.After(time.Second * time.Duration(timeout)):
			node.getLog().Debug(" Timeout ")
			newEvent := model.ReactorEvent{}
			newEvent.TransitionNodeId = node.meta.TimeoutTransition
			select {
			case nodeEventStream <- newEvent:
				return
			default:
				node.getLog().Debug("Message is dropped (no listeners) ")
			}
		case signal := <-node.flowOpCtx.NodeControlSignalChannel:
			node.getLog().Debug("Control signal ")
			if signal == model.SIGNAL_STOP {
				return
			}
		}
	}
}

func (node *ReceiveNode) OnInput( msg *model.Message) ([]model.NodeID,error) {
	return nil,nil
}
