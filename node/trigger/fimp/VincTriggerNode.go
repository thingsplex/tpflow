package fimp

import (
	"errors"
	"fmt"
	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/fimptype/primefimp"
	"github.com/mitchellh/mapstructure"
	"github.com/thingsplex/tpflow/model"
	"github.com/thingsplex/tpflow/node/base"
	"time"
)

type VincTriggerNode struct {
	base.BaseNode
	ctx                 *model.Context
	transport           *fimpgo.MqttTransport
	msgInStream         fimpgo.MessageCh
	msgInStreamName     string
	config              VincTriggerConfig
}

type VincTriggerConfig struct {
	Timeout                      int64 // in seconds
	ValueFilter                  string
	InputVariableType            string
	IsValueFilterEnabled         bool
	EventType                    string // mode/shortcut
}

func NewVincTriggerNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *model.Context) model.Node {
	node := VincTriggerNode{ctx: ctx}
	node.SetStartNode(true)
	node.SetMsgReactorNode(true)
	node.SetFlowOpCtx(flowOpCtx)
	node.SetMeta(meta)
	node.config = VincTriggerConfig{}
	node.msgInStreamName = node.FlowOpCtx().FlowId + "_" + string(node.GetMetaNode().Id)
	node.SetupBaseNode()
	return &node
}

func (node *VincTriggerNode) Init() error {
	node.initSubscriptions()
	return nil
}

func (node *VincTriggerNode) Cleanup() error {
	node.transport.UnregisterChannel(node.msgInStreamName)
	return nil
}

func (node *VincTriggerNode) initSubscriptions() {
	node.GetLog().Info("TriggerNode is listening for events . Name = ", node.Meta().Label)

	node.transport.Subscribe("pt:j1/mt:cmd/rt:app/rn:vinculum/ad:1")
	node.transport.Subscribe("pt:j1/mt:evt/rt:app/rn:vinculum/ad:1")

	node.msgInStream = make(fimpgo.MessageCh, 10)
	node.transport.RegisterChannelWithFilter(node.msgInStreamName, node.msgInStream,fimpgo.FimpFilter{
		Topic:     "pt:j1/+/rt:app/rn:vinculum/ad:1",
		Service:   "vinculum",
		Interface: "*",
	})
}

func (node *VincTriggerNode) LoadNodeConfig() error {
	err := mapstructure.Decode(node.Meta().Config, &node.config)
	if err != nil {
		node.GetLog().Error("Error while decoding node configs.Err:", err)
	}
	var ok bool
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

// WaitForEvent is started during flow initialization  or from another flow .
// Method acts as event listener and creates flow on new event .

func (node *VincTriggerNode) WaitForEvent(nodeEventStream chan model.ReactorEvent) {
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
		select {
		case newMsg := <-node.msgInStream:
			var eventValue string
			if newMsg.Payload.Type == "cmd.pd7.request" {
				request := primefimp.Request{}
				err := newMsg.Payload.GetObjectValue(&request)
				if err != nil {
					continue
				}
				if request.Component == "shortcut" && request.Cmd == "set"  {
					node.GetLog().Info("shortcut")
					if node.config.EventType == "shortcut" {
						eventValue = fmt.Sprintf("%.0f",request.Id)
					}
				}
			}else if newMsg.Payload.Type == "evt.pd7.notify" {
				notify := primefimp.Notify{}
				err := newMsg.Payload.GetObjectValue(&notify)
				if err != nil {
					continue
				}
				if notify.Component == "hub" && notify.Cmd == "set" {
					if node.config.EventType == "mode" {
						hub := notify.GetModeChange()
						if hub != nil {
							eventValue = hub.Current
						}else {
							node.GetLog().Info("ERROR 2")
						}
					}
				}
			}
			node.GetLog().Infof("Home event = %s",eventValue)
			if eventValue != "" {
				if !node.config.IsValueFilterEnabled || ((eventValue == node.config.ValueFilter) && node.config.IsValueFilterEnabled) {
					node.GetLog().Debug("Starting flow")
					rMsg := model.Message{Payload: fimpgo.FimpMessage{Value: eventValue, ValueType: fimpgo.VTypeString}}
					newEvent := model.ReactorEvent{Msg: rMsg, TransitionNodeId: node.Meta().SuccessTransition}
					// Flow is executed within flow runner goroutine
					go node.FlowRunner()(newEvent)
				}
			}

		case <-time.After(time.Second * time.Duration(timeout)):
			node.GetLog().Debug("Timeout ")
			newEvent := model.ReactorEvent{TransitionNodeId: node.Meta().TimeoutTransition}
			node.GetLog().Debug("Starting new flow (timeout)")
			go node.FlowRunner()(newEvent)
			node.GetLog().Debug("Flow started (timeout) ")
		case signal := <-node.FlowOpCtx().NodeControlSignalChannel:
			node.GetLog().Debug("Control signal ")
			if signal == model.SIGNAL_STOP {
				return
			}
		}
	}
}

func (node *VincTriggerNode) OnInput(msg *model.Message) ([]model.NodeID, error) {
	return nil, nil
}



/*

pt:j1/mt:cmd/rt:app/rn:vinculum/ad:1
{
  "corid": "",
  "ctime": "2020-03-04T17:05:13.836283",
  "props": null,
  "serv": "vinculum",
  "tags": null,
  "type": "cmd.pd7.request",
  "uid": "eea412c0-5e31-11ea-f2e6-9547cfed7ea1",
  "val_t": "object",
  "ver": "1",
  "val": {
    "cmd": "set",
    "component": "shortcut",
    "id": 9,
    "client": null,
    "param": null,
    "requestId": null
  },
  "resp_to": "pt:j1/mt:rsp/rt:cloud/rn:remote-client/ad:smarthome-app",
  "src": "app"
}
 */

/*
{
  "corid": "",
  "ctime": "2020-03-04T17:05:48+0100",
  "props": {},
  "serv": "vinculum",
  "tags": [],
  "type": "evt.pd7.notify",
  "uid": "8a59b1d2-5425-45c0-b177-aa30fc5d8299",
  "val": {
    "cmd": "set",
    "component": "hub",
    "id": "mode",
    "param": {
      "current": "sleep",
      "prev": "away"
    }
  },
  "val_t": "object",
  "ver": "1"
}

pt:j1/mt:evt/rt:app/rn:vinculum/ad:1

{
  "corid": "",
  "ctime": "2020-03-04T17:05:48+0100",
  "props": {},
  "serv": "vinculum",
  "tags": [],
  "type": "evt.pd7.notify",
  "uid": "815d394e-7e25-4a9f-af18-20fca36a08f0",
  "val": {
    "cmd": "set",
    "component": "house",
    "id": null,
    "param": {
      "fimp": true,
      "learning": null,
      "mode": "sleep",
      "time": "2020-03-04T16:05:48Z",
      "uptime": 718587649
    }
  },
  "val_t": "object",
  "ver": "1"
}
 */
