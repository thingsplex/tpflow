package websocket

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	ws "github.com/thingsplex/tpflow/connector/plugins/websocket"
	"github.com/thingsplex/tpflow/flow/context"
	"github.com/thingsplex/tpflow/model"
	"github.com/thingsplex/tpflow/node/base"
	"runtime/debug"
	"time"
)

//Node is trigger node which listening and gets triggered by websocket messages
type Node struct {
	base.BaseNode
	ctx          *context.Context
	connector    *ws.Connector
	config       Config
	nodeGlobalId string
	msgInStream  ws.MsgStream
}

type Config struct {
	Timeout       int64                 // timeout in seconds
	ConnectorID   string                // If set, the node will use non-default connector , for instance it can be used to listen events on remote hub
	PayloadFormat string                // fimp,json, none , form
	OutputVar     model.NodeVariableDef // Defines target storage for payload. If not set , payload will be set to input variable.
}

func NewWebsocketTriggerNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *context.Context) model.Node {
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
	node.msgInStream = node.connector.RegisterTrigger(string(node.Meta().Id))
	return nil
}

func (node *Node) Cleanup() error {
	node.connector.UnregisterTrigger(string(node.Meta().Id))
	return nil
}

func (node *Node) LoadNodeConfig() error {
	err := mapstructure.Decode(node.Meta().Config, &node.config)
	if err != nil {
		node.GetLog().Error("Error while decoding node configs.Err:", err)
	}

	var ok bool

	if node.config.ConnectorID == "" {
		node.config.ConnectorID = "websocket"
	}

	wsConnectionInstance := node.ConnectorRegistry().GetInstance(node.config.ConnectorID)
	if wsConnectionInstance != nil {
		node.connector, ok = wsConnectionInstance.Connection.GetConnection().(*ws.Connector)
		if !ok {
			node.GetLog().Error("can't cast connection to ws.Connector ")
			return errors.New("can't cast connection to ws.Connector ")
		}
	} else {
		node.GetLog().Error("Connector registry doesn't have ws.Connector instance")
		return errors.New("can't find ws.Connector connector")
	}
	return err
}

// WaitForEvent is started during flow initialization  or from another flow .
// Method acts as event listener and creates flow on new event .
func (node *Node) WaitForEvent(nodeEventStream chan model.ReactorEvent) {
	node.SetReactorRunning(true)
	timeout := time.Second * time.Duration(node.config.Timeout)
	var timer *time.Timer
	if timeout == 0 {
		timer = time.NewTimer(time.Hour * 24)
		timer.Stop()
	} else {
		timer = time.NewTimer(timeout)
	}
	defer func() {
		if r := recover(); r != nil {
			node.GetLog().Error("<WsConn> WS connection failed with PANIC")
			node.GetLog().Error(string(debug.Stack()))
		}
		node.SetReactorRunning(false)
		node.GetLog().Debug("<WsConn>WaitForEvent has quit. ")
		timer.Stop()
	}()

	for {
		if timeout > 0 {
			timer.Reset(timeout)
		}
		rMsg := model.Message{PayloadType: model.MsgPayloadNone}
		select {
		case newMsg := <-node.msgInStream:

			payloadFormat := model.MsgPayloadBinary
			if newMsg.PayloadEncoding == websocket.TextMessage {
				payloadFormat = model.MsgPayloadBinaryString
			}

			if node.config.OutputVar.Name != "" {
				var err error
				if payloadFormat == model.MsgPayloadBinary {
					err = node.ctx.SetVariable(node.config.OutputVar.Name, context.VarTypeBinBlob, newMsg.Payload, "", node.FlowOpCtx().FlowId, node.config.OutputVar.InMemory)
				} else {
					err = node.ctx.SetVariable(node.config.OutputVar.Name, context.VarTypeString, string(newMsg.Payload), "", node.FlowOpCtx().FlowId, node.config.OutputVar.InMemory)
				}

				if err != nil {
					node.GetLog().Error("Can't save input data to variable . Err :", err.Error())
					break
				}

			} else {
				rMsg.PayloadType = byte(payloadFormat)
				rMsg.RawPayload = newMsg.Payload
			}

			newEvent := model.ReactorEvent{Msg: rMsg, TransitionNodeId: node.Meta().SuccessTransition}
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
