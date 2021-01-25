package fimp

import (
	"bytes"
	"errors"
	"github.com/futurehomeno/fimpgo"
	"github.com/mitchellh/mapstructure"
	"github.com/thingsplex/tpflow/model"
	"github.com/thingsplex/tpflow/node/base"
	"github.com/thingsplex/tpflow/registry/storage"
	"text/template"
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
	addressTemplate 	*template.Template
	subAddress          string
}
// TODO : Copy Json path value extraction
type ReceiveConfig struct {
	Timeout                  int64 // in seconds
	ConnectorID              string // If set node will use non-default connector , for instance it can be used to listen events on remote hub
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
		if (node.activeSubscriptions)[i] == node.subAddress {
			needToSubscribe = false
			break
		}
	}
	if needToSubscribe {
		if node.subAddress != "" {
			node.GetLog().Info("Subscribing for service by address :", node.subAddress)
			node.transport.Subscribe(node.subAddress)
			node.activeSubscriptions = append(node.activeSubscriptions, node.subAddress)
		} else {
			node.GetLog().Error("Can't subscribe to service with empty address")
		}
	}
	node.msgInStream = make(fimpgo.MessageCh, 10)
	node.transport.RegisterChannelWithFilter(node.msgInStreamName, node.msgInStream,fimpgo.FimpFilter{
		Topic:     node.subAddress,
		Service:   node.Meta().Service,
		Interface: node.Meta().ServiceInterface,
	})
}

func (node *ReceiveNode) LoadNodeConfig() error {
	err := mapstructure.Decode(node.Meta().Config, &node.config)
	if err != nil {
		node.GetLog().Error("Failed to load node configs.Err:", err)
	}

	funcMap := template.FuncMap{
		"variable": func(varName string, isGlobal bool) (interface{}, error) {
			//node.GetLog().Debug("Getting variable by name ",varName)
			var vari model.Variable
			var err error
			if isGlobal {
				vari, err = node.ctx.GetVariable(varName, "global")
			} else {
				vari, err = node.ctx.GetVariable(varName, node.FlowOpCtx().FlowId)
			}

			if vari.IsNumber() {
				return vari.ToNumber()
			}
			vstr, ok := vari.Value.(string)
			if ok {
				return vstr, err
			} else {
				node.GetLog().Debug("Only simple types are supported ")
				return "", errors.New("Only simple types are supported ")
			}
		},
		"setting": func(name string) (interface{}, error) {
			if node.FlowOpCtx().FlowMeta.Settings != nil {
				s := node.FlowOpCtx().FlowMeta.Settings[name]
				return s.String(),nil
			}else {
				return "",nil
			}
		},
	}
	node.addressTemplate, err = template.New("address").Funcs(funcMap).Parse(node.Meta().Address)
	var addrTemplateBuffer bytes.Buffer
	node.addressTemplate.Execute(&addrTemplateBuffer, nil)
	node.subAddress = addrTemplateBuffer.String()

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

	if node.config.ConnectorID == "" {
		node.config.ConnectorID = "fimpmqtt"
	}

	fimpTransportInstance := node.ConnectorRegistry().GetInstance(node.config.ConnectorID)
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
	timeout := time.Second * time.Duration(node.config.Timeout)
	var timer  *time.Timer
	if timeout == 0 {
		timer = time.NewTimer(time.Hour * 24)
		timer.Stop()
	}else {
		timer = time.NewTimer(timeout)
	}
	defer func() {
		node.SetReactorRunning(false)
		node.GetLog().Debug("Reactor-WaitForEvent is stopped ")
		timer.Stop()
	}()
	node.GetLog().Debug("Reactor-Waiting for event .chan size = ", len(node.msgInStream))
	for {
		if timeout > 0 {
			timer.Reset(timeout)
		}
		select {
		case newMsg := <-node.msgInStream:
			node.GetLog().Debug("New message :")
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

		case <-timer.C:
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
