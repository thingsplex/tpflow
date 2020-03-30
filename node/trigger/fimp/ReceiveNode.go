package fimp

import (
	"errors"
	"github.com/futurehomeno/fimpgo"
	"github.com/mitchellh/mapstructure"
	"github.com/thingsplex/tpflow/model"
	"github.com/thingsplex/tpflow/node/base"
	"github.com/thingsplex/tpflow/registry/storage"
	"time"
)

// ReceiveNode
type ReceiveNode struct {
	base.BaseNode
	ctx                 *model.Context
	transport           *fimpgo.MqttTransport
	activeSubscriptions []string
	msgInStream         fimpgo.MessageCh
	msgInStreamName     string
	config              ReceiveConfig
	thingRegistry       *storage.LocalRegistryStore
}

type ReceiveConfig struct {
	Timeout                  int64 // in seconds
	ValueFilter              model.Variable
	InputVariableType        string
	IsValueFilterEnabled     bool
	RegisterAsVirtualService bool
	VirtualServiceGroup      string
}

func NewReceiveNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *model.Context) model.Node {
	node := ReceiveNode{ctx: ctx}
	node.SetStartNode(false)
	node.SetMsgReactorNode(true)
	node.SetFlowOpCtx(flowOpCtx)
	node.SetMeta(meta)
	node.config = ReceiveConfig{}
	node.msgInStreamName = node.FlowOpCtx().FlowId + "_" + string(node.GetMetaNode().Id)
	node.SetupBaseNode()
	node.activeSubscriptions = []string{}
	return &node
}

func (node *ReceiveNode) Init() error {
	node.activeSubscriptions = []string{}
	node.initSubscriptions()
	return nil
}

func (node *ReceiveNode) Cleanup() error {
	node.transport.UnregisterChannel(node.msgInStreamName)
	return nil
}

func (node *ReceiveNode) initSubscriptions() {
	node.GetLog().Info("ReceiveNode is listening for events . Name = ", node.Meta().Label)
	needToSubscribe := true
	for i := range node.activeSubscriptions {
		if (node.activeSubscriptions)[i] == node.Meta().Address {
			needToSubscribe = false
			break
		}
	}
	if needToSubscribe {
		if node.Meta().Address != "" {
			node.GetLog().Info("Subscribing for service by address :", node.Meta().Address)
			node.transport.Subscribe(node.Meta().Address)
			node.activeSubscriptions = append(node.activeSubscriptions, node.Meta().Address)
		} else {
			node.GetLog().Error("Can't subscribe to service with empty address")
		}
	}
	node.msgInStream = make(fimpgo.MessageCh, 10)
	node.transport.RegisterChannelWithFilter(node.msgInStreamName, node.msgInStream,fimpgo.FimpFilter{
		Topic:     node.Meta().Address,
		Service:   node.Meta().Service,
		Interface: node.Meta().ServiceInterface,
	})
}

func (node *ReceiveNode) LoadNodeConfig() error {
	err := mapstructure.Decode(node.Meta().Config, &node.config)
	if err != nil {
		node.GetLog().Error("Failed to load node configs.Err:", err)
	}

	connInstance := node.ConnectorRegistry().GetInstance("thing_registry")
	var ok bool
	if connInstance != nil {
		node.thingRegistry, ok = connInstance.Connection.(*storage.LocalRegistryStore)
		if !ok {
			node.thingRegistry = nil
			node.GetLog().Error("Can't get things connection to things registry . Cast to LocalRegistryStore failed")
		}
	} else {
		node.GetLog().Error("Connector registry doesn't have thing_registry instance")
	}

	fimpTransportInstance := node.ConnectorRegistry().GetInstance("fimpmqtt")
	if fimpTransportInstance != nil {
		node.transport, ok = fimpTransportInstance.Connection.GetConnection().(*fimpgo.MqttTransport)
		if !ok {
			node.GetLog().Error("can't cast connection to mqttfimpgo ")
			return errors.New("can't cast connection to mqttfimpgo ")
		}
	} else {
		node.GetLog().Error("Connector registry doesn't have fimp instance")
		return errors.New("can't find fimp connector")
	}


	return err
}

func (node *ReceiveNode) WaitForEvent(nodeEventStream chan model.ReactorEvent) {
	node.SetReactorRunning(true)
	defer func() {
		node.SetReactorRunning(false)
		node.GetLog().Debug("Reactor-WaitForEvent is stopped ")
	}()
	node.GetLog().Debug("Reactor-Waiting for event .chan size = ", len(node.msgInStream))
	//start := time.Now()
	timeout := node.config.Timeout
	if timeout == 0 {
		timeout = 86400 // 24 hours
	}

	for {
		select {
		case newMsg := <-node.msgInStream:
			node.GetLog().Info("New message :")
			if !node.config.IsValueFilterEnabled {
					rMsg := model.Message{AddressStr: newMsg.Topic, Address: *newMsg.Addr, Payload: *newMsg.Payload}
					newEvent := model.ReactorEvent{Msg: rMsg, TransitionNodeId: node.Meta().SuccessTransition}
					select {
					case nodeEventStream <- newEvent:
						return
					default:
						node.GetLog().Debug("Message is dropped (no listeners) ")
					}

			} else if newMsg.Payload.Value == node.config.ValueFilter.Value {
					rMsg := model.Message{AddressStr: newMsg.Topic, Address: *newMsg.Addr, Payload: *newMsg.Payload}
					newEvent := model.ReactorEvent{Msg: rMsg, TransitionNodeId: node.Meta().SuccessTransition}
					select {
					case nodeEventStream <- newEvent:
						return
					default:
						node.GetLog().Debug("Message is dropped (no listeners) ")
					}
			}
			//if node.config.Timeout > 0 {
			//	elapsed := time.Since(start)
			//	timeout = timeout - int64(elapsed.Seconds())
			//}

		case <-time.After(time.Second * time.Duration(timeout)):
			node.GetLog().Debug(" Timeout ")
			newEvent := model.ReactorEvent{}
			newEvent.TransitionNodeId = node.Meta().TimeoutTransition
			select {
			case nodeEventStream <- newEvent:
				return
			default:
				node.GetLog().Debug("Message is dropped (no listeners) ")
			}
		case signal := <-node.FlowOpCtx().TriggerControlSignalChannel:
			node.GetLog().Debug("Control signal ")
			if signal == model.SIGNAL_STOP {
				node.GetLog().Info("Received node stopped by SIGNAL_STOP ")
				return
			}
		}
	}
}

func (node *ReceiveNode) OnInput(msg *model.Message) ([]model.NodeID, error) {
	return nil, nil
}
