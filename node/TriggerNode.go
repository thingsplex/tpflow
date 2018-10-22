package node

import (
	"errors"
	"github.com/alivinco/fimpgo"
	"github.com/alivinco/tpflow/model"
	"github.com/alivinco/tpflow/registry"
	"github.com/alivinco/tpflow/utils"
	"github.com/mitchellh/mapstructure"
	"time"
)

type TriggerNode struct {
	BaseNode
	ctx                 *model.Context
	transport           *fimpgo.MqttTransport
	activeSubscriptions []string
	msgInStream         fimpgo.MessageCh
	config              TriggerConfig
	thingRegistry       *registry.ThingRegistryStore
}

type TriggerConfig struct {
	Timeout int64 // in seconds
	ValueFilter model.Variable
	InputVariableType string
	IsValueFilterEnabled bool
	RegisterAsVirtualService bool // if true - the node will be exposed as service in inclusion report
	LookupServiceNameAndLocation bool
	VirtualServiceGroup string   // is used as service group in inclusion report
	VirtualServiceProps map[string]interface{} // mostly used to announce supported features of the service , for instance supported modes , states , setpoints , etc...
}

func NewTriggerNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *model.Context) model.Node {
	node := TriggerNode{ctx: ctx}
	node.isStartNode = true
	node.isMsgReactor = true
	node.flowOpCtx = flowOpCtx
	node.meta = meta
	node.config = TriggerConfig{}
	node.SetupBaseNode()
	node.activeSubscriptions = []string{}
	return &node
}

func (node *TriggerNode) Init() error {
	node.initSubscriptions()
	return nil
}

func (node *TriggerNode) ConfigureInStream(activeSubscriptions *[]string, msgInStream model.MsgPipeline) {
	//node.getLog().Info("Configuring Stream")
	//node.activeSubscriptions = activeSubscriptions
	//node.msgInStream = msgInStream
	//node.initSubscriptions()
}

func (node *TriggerNode) initSubscriptions() {
	if node.meta.Type == "trigger" {
		node.getLog().Info("TriggerNode is listening for events . Name = ", node.meta.Label)
		needToSubscribe := true
		for i := range node.activeSubscriptions {
			if (node.activeSubscriptions)[i] == node.meta.Address {
				needToSubscribe = false
				break
			}
		}
		if needToSubscribe {
			if node.meta.Address != ""{
				node.getLog().Info("Subscribing for service by address :", node.meta.Address)
				node.transport.Subscribe(node.meta.Address)
				node.activeSubscriptions = append(node.activeSubscriptions, node.meta.Address)
			}else {
				node.getLog().Error(" Can't subscribe to service with empty address")
			}

		}
	}
}

func (node *TriggerNode) LoadNodeConfig() error {
	err := mapstructure.Decode(node.meta.Config,&node.config)
	if err != nil{
		node.getLog().Error("Error while decoding node configs.Err:",err)
	}

	connInstance := node.connectorRegistry.GetInstance("thing_registry")
	var ok bool;
	if connInstance != nil {
		node.thingRegistry,ok = connInstance.Connection.(*registry.ThingRegistryStore)
		if !ok {
			node.thingRegistry = nil
			node.getLog().Error("Can't get things connection to things registry . Cast to ThingRegistryStore failed")
		}
	}else {
		node.getLog().Error("Connector registry doesn't have thing_registry instance")
	}

	fimpTransportInstance := node.connectorRegistry.GetInstance("fimpmqtt")
	if fimpTransportInstance != nil {
		node.transport,ok = fimpTransportInstance.Connection.GetConnection().(*fimpgo.MqttTransport)
		if !ok {
			node.getLog().Error("can't cast connection to mqttfimpgo ")
			return errors.New("can't cast connection to mqttfimpgo ")
		}
	}else {
		node.getLog().Error("Connector registry doesn't have fimp instance")
		return errors.New("can't find fimp connector")
	}
	node.msgInStream = make(fimpgo.MessageCh,10)
	node.transport.RegisterChannel(node.flowOpCtx.FlowId+"_"+string(node.GetMetaNode().Id),node.msgInStream)
	return err
}

func (node *TriggerNode) LookupAddressToAlias(address string) {

	if node.thingRegistry == nil {
		return
	}
	service,err := node.thingRegistry.GetServiceByFullAddress(address)
	if err == nil {
		node.ctx.SetVariable("flow_service_alias","string",service.Alias,"",node.flowOpCtx.FlowId,true)
		node.ctx.SetVariable("flow_location_alias","string",service.LocationAlias,"",node.flowOpCtx.FlowId,true)
	}
}

func (node *TriggerNode) WaitForEvent(nodeEventStream chan model.ReactorEvent) {
	node.isReactorRunning = true
	defer func() {
		node.isReactorRunning = false
		node.getLog().Debug("Msg processed by the node ")
	}()
	timeout := node.config.Timeout
	if timeout == 0 {
		timeout = 86400 // 24 hours
	}
	for {
		start := time.Now()
		node.getLog().Debug( "Waiting for msg")
		select {
		case newMsg := <-node.msgInStream:
			node.getLog().Debug("--New message--")
			if utils.RouteIncludesTopic(node.meta.Address,newMsg.Topic) &&
				(newMsg.Payload.Service == node.meta.Service || node.meta.Service == "*") &&
				(newMsg.Payload.Type == node.meta.ServiceInterface || node.meta.ServiceInterface == "*") {

				if !node.config.IsValueFilterEnabled || ( (newMsg.Payload.Value == node.config.ValueFilter.Value) && node.config.IsValueFilterEnabled)  {
					rMsg := model.Message{AddressStr:newMsg.Topic,Address:*newMsg.Addr,Payload:*newMsg.Payload}
					newEvent := model.ReactorEvent{Msg:rMsg,TransitionNodeId:node.meta.SuccessTransition}
					if node.config.LookupServiceNameAndLocation {
						node.LookupAddressToAlias(newEvent.Msg.AddressStr)
					}
					go node.flowRunner(newEvent)
				}
			}else {
				node.getLog().Debug("Not interested .")
			}

			if node.config.Timeout > 0 {
				elapsed := time.Since(start)
				timeout =  timeout - int64(elapsed.Seconds())
			}


		case <-time.After(time.Second * time.Duration(timeout)):
			node.getLog().Debug("Timeout ")
			newEvent := model.ReactorEvent{TransitionNodeId:node.meta.TimeoutTransition}
			node.getLog().Debug("Starting new flow (timeout)")
			go node.flowRunner(newEvent)
			node.getLog().Debug("Flow started (timeout) ")
		case signal := <-node.flowOpCtx.NodeControlSignalChannel:
			node.getLog().Debug("Control signal ")
			if signal == model.SIGNAL_STOP {
				return
			}
		}
	}
}

func (node *TriggerNode) OnInput(msg *model.Message) ([]model.NodeID, error) {
	return nil,nil
}
