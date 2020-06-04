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

type TriggerNode struct {
	base.BaseNode
	ctx                 *model.Context
	transport           *fimpgo.MqttTransport
	activeSubscriptions []string
	msgInStream         fimpgo.MessageCh
	msgInStreamName     string
	config              TriggerConfig
	thingRegistry       storage.RegistryStorage
}

type TriggerConfig struct {
	Timeout                      int64 // in seconds
	ValueFilter                  model.Variable
	PropFilterName               string // fimp props filter . Property name
	PropFilterValue              string // Property value
	ValueJPath                   string // JPath path which is used to extract value from trigger message
	ValueJPathResultType         string // Type of extracted variable
	InputVariableType            string
	IsValueFilterEnabled         bool
	RegisterAsVirtualService     bool // if true - the node will be exposed as service in inclusion report
	LookupServiceNameAndLocation bool
	VirtualServiceGroup          string                 // is used as service group in inclusion report
	VirtualServiceProps          map[string]interface{} // mostly used to announce supported features of the service , for instance supported modes , states , setpoints , etc...

}

func NewTriggerNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *model.Context) model.Node {
	node := TriggerNode{ctx: ctx}
	node.SetStartNode(true)
	node.SetMsgReactorNode(true)
	node.SetFlowOpCtx(flowOpCtx)
	node.SetMeta(meta)
	node.config = TriggerConfig{}
	node.msgInStreamName = node.FlowOpCtx().FlowId + "_" + string(node.GetMetaNode().Id)
	node.SetupBaseNode()
	node.activeSubscriptions = []string{}
	return &node
}

func (node *TriggerNode) Init() error {
	node.activeSubscriptions = []string{}
	node.initSubscriptions()
	return nil
}

func (node *TriggerNode) Cleanup() error {
	node.transport.UnregisterChannel(node.msgInStreamName)
	return nil
}

func (node *TriggerNode) initSubscriptions() {
	if node.GetMetaNode().Type == "trigger" {
		node.GetLog().Info("TriggerNode is listening for events . Name = ", node.Meta().Label)
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
				node.GetLog().Error(" Can't subscribe to service with empty address")
			}
		}
	}
	node.msgInStream = make(fimpgo.MessageCh, 10)
	node.transport.RegisterChannelWithFilter(node.msgInStreamName, node.msgInStream,fimpgo.FimpFilter{
		Topic:     node.Meta().Address,
		Service:   node.Meta().Service,
		Interface: node.Meta().ServiceInterface,
	})

}

func (node *TriggerNode) LoadNodeConfig() error {
	err := mapstructure.Decode(node.Meta().Config, &node.config)
	if err != nil {
		node.GetLog().Error("Error while decoding node configs.Err:", err)
	}

	connInstance := node.ConnectorRegistry().GetInstance("thing_registry")
	var ok bool
	if connInstance != nil {
		node.thingRegistry, ok = connInstance.Connection.(storage.RegistryStorage)
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

func (node *TriggerNode) LookupAddressToAlias(address string) {

	if node.thingRegistry == nil {
		return
	}
	service, err := node.thingRegistry.GetServiceByFullAddress(address)
	if err == nil {
		node.ctx.SetVariable("flow_service_alias", "string", service.Alias, "", node.FlowOpCtx().FlowId, true)
		node.ctx.SetVariable("flow_location_alias", "string", service.LocationAlias, "", node.FlowOpCtx().FlowId, true)
		node.ctx.SetVariable("flow_location_type", "string", service.LocationType, "", node.FlowOpCtx().FlowId, true)
		node.ctx.SetVariable("flow_location_sub_type", "string", service.LocationSubType, "", node.FlowOpCtx().FlowId, true)
	}
}
// WaitForEvent is started during flow initialization  or from another flow .
// Method acts as event listener and creates flow on new event .

func (node *TriggerNode) WaitForEvent(nodeEventStream chan model.ReactorEvent) {
	node.SetReactorRunning(true)
	defer func() {
		node.SetReactorRunning(false)
		node.GetLog().Debug("Msg processed by the node ")
	}()
	timeout := node.config.Timeout
	if timeout == 0 {
		timeout = 86400 // 24 hours
	}
	for {
		//start := time.Now()
		//node.GetLog().Debug("Waiting for msg")
		select {
		case newMsg := <-node.msgInStream:
			//node.GetLog().Debug("--New message--")
			node.GetLog().Debug("--New message--")
			if node.config.ValueJPath != "" {
				rMsg := model.Message{RawPayload:newMsg.Payload.ValueObj }
				newVal , err := model.GetValueByPath(&rMsg,"jpath", node.config.ValueJPath, node.config.ValueJPathResultType)
				if err == nil {
					node.GetLog().Debug("JPath extracted value =",newVal)
					newMsg.Payload = fimpgo.NewMessage(newMsg.Payload.Type,newMsg.Payload.Service,node.config.ValueJPathResultType,newVal,nil,nil,nil)
				}else {
					node.GetLog().Debug("Can't extract value using JPath.Err:",err)
					continue
				}
			}
			if !node.config.IsValueFilterEnabled || ((newMsg.Payload.Value == node.config.ValueFilter.Value) && node.config.IsValueFilterEnabled) {

					if node.config.PropFilterName != "" {
						// Properties filter is enabled
						if newMsg.Payload.Properties[node.config.PropFilterName] != node.config.PropFilterValue {
							// Props value didn't match
							continue
						}
					}
					rMsg := model.Message{AddressStr: newMsg.Topic, Address: *newMsg.Addr, Payload: *newMsg.Payload}
					newEvent := model.ReactorEvent{Msg: rMsg, TransitionNodeId: node.Meta().SuccessTransition}
					if node.config.LookupServiceNameAndLocation {
						node.LookupAddressToAlias(newEvent.Msg.AddressStr)
					}
					// Flow is executed within flow runner goroutine
					node.FlowRunner()(newEvent)
			}

			//if node.config.Timeout > 0 {
			//	elapsed := time.Since(start)
			//	timeout = timeout - int64(elapsed.Seconds())
			//}

		case <-time.After(time.Second * time.Duration(timeout)):
			node.GetLog().Debug("Timeout ")
			newEvent := model.ReactorEvent{TransitionNodeId: node.Meta().TimeoutTransition}
			node.GetLog().Debug("Starting new flow (timeout)")
			node.FlowRunner()(newEvent)
			node.GetLog().Debug("Flow started (timeout) ")
		case signal := <-node.FlowOpCtx().TriggerControlSignalChannel:
			node.GetLog().Debug("Control signal ")
			if signal == model.SIGNAL_STOP {
				node.GetLog().Info("Trigger stopped by SIGNAL_STOP ")
				return
			}
		}
	}
}

func (node *TriggerNode) OnInput(msg *model.Message) ([]model.NodeID, error) {
	return nil, nil
}
