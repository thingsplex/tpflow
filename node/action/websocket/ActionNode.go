package websocket

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	ws "github.com/thingsplex/tpflow/connector/plugins/websocket"
	"github.com/thingsplex/tpflow/flow/context"
	"github.com/thingsplex/tpflow/model"
	"github.com/thingsplex/tpflow/node/base"
	"text/template"
)

type Node struct {
	base.BaseNode
	ctx             *context.Context
	connector       *ws.Connector
	config          NodeConfig
	addressTemplate *template.Template
}

type NodeConfig struct {
	ConnectorID     string
	InputVar        model.NodeVariableDef
	PayloadFormat   string
	WsPayloadType   string
	RequestTemplate string // content of template will be used if input variable is not set .
}

func NewNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *context.Context) model.Node {
	node := Node{ctx: ctx}
	node.SetMeta(meta)
	node.SetFlowOpCtx(flowOpCtx)
	node.config = NodeConfig{}
	node.SetupBaseNode()
	return &node
}

func (node *Node) LoadNodeConfig() error {
	err := mapstructure.Decode(node.Meta().Config, &node.config)
	if err != nil {
		node.GetLog().Error("Can't decode config.Err:", err)

	}

	if node.config.ConnectorID == "" {
		node.config.ConnectorID = "websocket-client"
	}
	var ok bool
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

func (node *Node) WaitForEvent(responseChannel chan model.ReactorEvent) {

}

func (node *Node) OnInput(msg *model.Message) ([]model.NodeID, error) {
	node.GetLog().Info("Executing Node . Name = ", node.Meta().Label)
	var body []byte
	var err error
	if node.config.InputVar.Name != "" {
		// variable of arbitrary type
		variable, err := node.ctx.GetVariable(node.config.InputVar.Name, node.FlowOpCtx().FlowId)
		if err != nil {
			node.GetLog().Error("Can't get variable . Error:", err)
			return nil, err
		}
		body, err = variable.ToBinary()
	} else {
		// fimp handling , serializing fimp into binary payload
		if msg.PayloadType == model.MsgPayloadFimp {
			msg.Payload.Topic = msg.AddressStr
			body, err = msg.Payload.SerializeToJson()
		} else {
			body = msg.RawPayload
		}
	}
	if err != nil {
		return []model.NodeID{node.Meta().ErrorTransition}, err
	}
	node.GetLog().Debug("Sending message to websocket server. Payload:", string(body))
	err = node.connector.Publish(websocket.TextMessage, body)
	if err != nil {
		return []model.NodeID{node.Meta().ErrorTransition}, err
	}
	return []model.NodeID{node.Meta().SuccessTransition}, nil
}
