package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/futurehomeno/fimpgo"
	"github.com/mitchellh/mapstructure"
	"github.com/thingsplex/tpflow/connector/plugins/http"
	"github.com/thingsplex/tpflow/flow/context"
	"github.com/thingsplex/tpflow/model"
	"github.com/thingsplex/tpflow/node/base"
	"io/ioutil"
	"time"
)

const (
	PayloadFormatFimp = "fimp"
	PayloadFormatJson = "json"
	PayloadFormatNone = "none"
	PayloadFormatForm = "form"
)

type Node struct {
	base.BaseNode
	ctx            *context.Context
	httpServerConn *http.Connector
	config         Config
	msgInStream    chan http.RequestEvent
	nodeGlobalId   string
}

type Config struct {
	Timeout             int64 // in seconds
	ConnectorID         string // If set node will use non-default connector , for instance it can be used to listen events on remote hub
	PayloadFormat       string // fimp,json, none , form
	Method              string // GET,POST,PUT,DELETE
	IsSync              bool   // sync - the flow will wait for reply action , async - response is returned to caller right away
	IsWs                bool   // A flow must have only one websocket trigger node.
	MapFormParamsToVars bool
	Alias               string // alias to be used instead of flow id
	OutputVar           model.NodeVariableDef // Defines target storage for payload. If not set , payload will be set to input variable.
	AuthConfig          http.AuthConfig // Node level auth config
}

func NewHttpTriggerNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *context.Context) model.Node {
	node := Node{ctx: ctx}
	node.SetStartNode(true)
	node.SetMsgReactorNode(true)
	node.SetFlowOpCtx(flowOpCtx)
	node.SetMeta(meta)
	node.config = Config{}
	node.SetupBaseNode()
	return &node
}

func (node *Node) Init() error {
	node.msgInStream = make(chan http.RequestEvent,10)
	name := fmt.Sprintf("%s %s", node.FlowOpCtx().FlowMeta.Name,node.BaseNode.GetMetaNode().Label)
	node.httpServerConn.RegisterFlow(node.nodeGlobalId,node.config.IsSync,node.config.IsWs,false,node.msgInStream,node.config.Alias,name,node.config.AuthConfig)
	return nil
}

func (node *Node) Cleanup() error {
	if node.httpServerConn != nil {
		node.httpServerConn.UnregisterFlow(node.FlowOpCtx().FlowId)
	}

	return nil
}

func (node *Node) LoadNodeConfig() error {
	err := mapstructure.Decode(node.Meta().Config, &node.config)
	if err != nil {
		node.GetLog().Error("Error while decoding node configs.Err:", err)
	}

	var ok bool

	if node.config.ConnectorID == "" {
		node.config.ConnectorID = "httpserver"
	}

	httpServInstance := node.ConnectorRegistry().GetInstance(node.config.ConnectorID)
	if httpServInstance != nil {
		node.httpServerConn, ok = httpServInstance.Connection.GetConnection().(*http.Connector)
		if !ok {
			node.GetLog().Error("can't cast connection to httpserv ")
			return errors.New("can't cast connection to httpserv ")
		}
	} else {
		node.GetLog().Error("Connector registry doesn't have httpserv instance")
		return errors.New("can't find httpserv connector")
	}
	if node.config.IsWs {
		// Flow shares single WS connection.
		node.nodeGlobalId = node.FlowOpCtx().FlowId
	}else {
		// Flow can have multiple rest endpoints.
		node.nodeGlobalId = node.FlowOpCtx().FlowId + "_" + string(node.GetMetaNode().Id)
	}

	return err
}

// WaitForEvent is started during flow initialization  or from another flow .
// Method acts as event listener and creates flow on new event .
func (node *Node) WaitForEvent(nodeEventStream chan model.ReactorEvent) {
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
		node.GetLog().Debug("WaitForEvent has quit. ")
		timer.Stop()
	}()

	for {
		if timeout > 0 {
			timer.Reset(timeout)
		}
		var fimpMsg *fimpgo.FimpMessage
		select {
		case newMsg := <-node.msgInStream:
			node.GetLog().Debug("--New http message--")
			if newMsg.HttpRequest.Method != node.config.Method {
				node.httpServerConn.ReplyToRequest(newMsg.RequestId, nil, "")
				continue
			}
			var err error
			switch node.config.PayloadFormat {
			case PayloadFormatFimp:
				var body []byte
				if newMsg.IsWsMsg || newMsg.IsFromCloud {
					body = newMsg.Payload
				}else {
					body, err = ioutil.ReadAll(newMsg.HttpRequest.Body)
					if err != nil {
						node.GetLog().Info("Can't read http payload . Err ",err.Error())
						node.httpServerConn.ReplyToRequest(newMsg.RequestId, nil, "")
						continue
					}
				}

				fimpMsg,err = fimpgo.NewMessageFromBytes(body)
				if err != nil {
					node.GetLog().Info("Can't decode fimp message from http payload . Err ",err.Error())
					node.httpServerConn.ReplyToRequest(newMsg.RequestId, nil, "")
					continue
				}

				if node.config.OutputVar.Name != "" {
					err = node.ctx.SetVariable(node.config.OutputVar.Name, fimpgo.VTypeObject, fimpMsg,"", node.FlowOpCtx().FlowId, node.config.OutputVar.InMemory)
					if err != nil {
						node.GetLog().Error("Can't save input data to variable . Err :",err.Error())
					}
					fimpMsg = nil
				}

			case PayloadFormatJson:
				var body []byte
				if newMsg.IsWsMsg || newMsg.IsFromCloud {
					body = newMsg.Payload
				}else {
					body, err = ioutil.ReadAll(newMsg.HttpRequest.Body)
					if err != nil  {
						node.GetLog().Info("Can't read http payload or payload is empty . Err:",err.Error())
						node.httpServerConn.ReplyToRequest(newMsg.RequestId, nil, "")
						continue
					}
				}
				node.GetLog().Debug("New json message: ",string(body))
				var jVal interface{}
				err = json.Unmarshal(body,&jVal)
				if err != nil {
					node.GetLog().Info("Can't unmarshal JSON from HTTP payload . Err ",err.Error())
					node.httpServerConn.ReplyToRequest(newMsg.RequestId, nil, "")
					continue
				}
				if node.config.OutputVar.Name != "" {
					err = node.ctx.SetVariable(node.config.OutputVar.Name, fimpgo.VTypeObject, jVal,"", node.FlowOpCtx().FlowId, node.config.OutputVar.InMemory)
					if err != nil {
						node.GetLog().Error("Can't save input data to variable . Err :",err.Error())
					}
					fimpMsg = nil
				}else {
					fimpMsg = fimpgo.NewObjectMessage("evt.httpserv.json_request","tpflow",jVal,nil,nil,nil)
				}

			case PayloadFormatNone:
				//fimpMsg = fimpgo.NewNullMessage("evt.httpserv.empty_request","tpflow",nil,nil,nil)

			case PayloadFormatForm:
				err := newMsg.HttpRequest.ParseForm()
				if err != nil {
					node.GetLog().Info("Can't parse form params . Err:",err.Error())
				}
				r := map[string]string {}
				node.GetLog().Info("Mapping param ",newMsg.HttpRequest.Form)
				for k,v :=range newMsg.HttpRequest.Form {
					if len(v)>0 {
						r[k] = v[0]
						if node.config.MapFormParamsToVars {
							node.ctx.SetVariable(k, "string", v[0], "", node.FlowOpCtx().FlowId, true)
						}
					}
				}
				if node.config.OutputVar.Name != "" {
					err = node.ctx.SetVariable(node.config.OutputVar.Name, fimpgo.VTypeObject, fimpMsg,"", node.FlowOpCtx().FlowId, node.config.OutputVar.InMemory)
					if err != nil {
						node.GetLog().Error("Can't save input data to variable . Err :",err.Error())
					}
					fimpMsg = nil
				}else {
					fimpMsg = fimpgo.NewObjectMessage("evt.httpserv.form_request","tpflow",r,nil,nil,nil)
				}
			}

			if !node.config.IsSync {
				node.httpServerConn.ReplyToRequest(newMsg.RequestId, nil, "")
			}

			rMsg := model.Message{RequestId: newMsg.RequestId}
			if fimpMsg != nil {
				rMsg.Payload = *fimpMsg
			}
			newEvent := model.ReactorEvent{Msg: rMsg, TransitionNodeId: node.Meta().SuccessTransition}
			// Flow is executed within flow runner goroutine
			node.FlowRunner()(newEvent)

		case <-timer.C:
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

func (node *Node) OnInput(msg *model.Message) ([]model.NodeID, error) {
	return nil, nil
}
