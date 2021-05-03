package http

import (
	"encoding/json"
	"errors"
	"github.com/futurehomeno/fimpgo"
	"github.com/mitchellh/mapstructure"
	"github.com/thingsplex/tpflow/connector/plugins/http"
	"github.com/thingsplex/tpflow/model"
	"github.com/thingsplex/tpflow/node/base"
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
	ctx            *model.Context
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
}

func NewHttpTriggerNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *model.Context) model.Node {
	node := Node{ctx: ctx}
	node.SetStartNode(true)
	node.SetMsgReactorNode(true)
	node.SetFlowOpCtx(flowOpCtx)
	node.SetMeta(meta)
	node.config = Config{}
	node.SetupBaseNode()
	node.nodeGlobalId = node.FlowOpCtx().FlowId + "_" + string(node.GetMetaNode().Id)
	return &node
}

func (node *Node) Init() error {
	node.msgInStream = make(chan http.RequestEvent,10)
	node.httpServerConn.RegisterFlow(node.nodeGlobalId,node.config.IsSync,node.msgInStream)
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
				continue
			}

			switch node.config.PayloadFormat {
			case PayloadFormatFimp:
				var body []byte
				_ , err := newMsg.HttpRequest.Body.Read(body)
				if err != nil {
					node.GetLog().Info("Can't read http payload . Err ",err.Error())
					continue
				}
				fimpMsg,err = fimpgo.NewMessageFromBytes(body)
				if err != nil {
					node.GetLog().Info("Can't decode fimp message from http payload . Err ",err.Error())
					continue
				}
			case PayloadFormatJson:
				var body []byte
				_ , err := newMsg.HttpRequest.Body.Read(body)
				if err != nil {
					node.GetLog().Info("Can't read http payload . Err ",err.Error())
					continue
				}
				var jVal interface{}
				err = json.Unmarshal(body,jVal)
				if err != nil {
					node.GetLog().Info("Can't unmarshal JSON from HTTP payload . Err ",err.Error())
					continue
				}
				fimpMsg = fimpgo.NewObjectMessage("evt.httpserv.json_request","tpflow",jVal,nil,nil,nil)

			case PayloadFormatNone:
				fimpMsg = fimpgo.NewNullMessage("evt.httpserv.empty_request","tpflow",nil,nil,nil)

			case PayloadFormatForm:
				newMsg.HttpRequest.ParseForm()
				fimpMsg = fimpgo.NewObjectMessage("evt.httpserv.form_request","tpflow",	newMsg.HttpRequest.Form,nil,nil,nil)
			}

			rMsg := model.Message{RequestId: newMsg.RequestId,Payload: *fimpMsg}
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
