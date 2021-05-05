package http

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/thingsplex/tpflow/connector/plugins/http"
	"github.com/thingsplex/tpflow/model"
	"github.com/thingsplex/tpflow/node/base"
)

// Node is Http reply node that sends response to request received from HTTP trigger node
type Node struct {
	base.BaseNode
	ctx            *model.Context
	config         NodeConfig
	httpServerConn *http.Connector
	nodeGlobalId   string
}

type NodeConfig struct {
	ConnectorID           string
	InputVar              model.NodeVariableDef
	ResponsePayloadFormat string // json,fimp,xml,html
	IsWs                  bool   // message is sent over active WS connection
	IsPublishOnly         bool   // true - means there is no trigger node , only publish
}

func NewNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *model.Context) model.Node {
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
		node.GetLog().Error(" Failed while loading configurations.Error:", err)
		return err
	}
	if node.config.ConnectorID == "" {
		node.config.ConnectorID = "httpserver"
	}
	var ok bool
	httpServInstance := node.ConnectorRegistry().GetInstance(node.config.ConnectorID)
	if httpServInstance != nil {
		node.httpServerConn, ok = httpServInstance.Connection.GetConnection().(*http.Connector)
		if !ok {
			node.GetLog().Error("can't cast connection to httpserv ")
			return fmt.Errorf("can't cast connection to httpserv ")
		}
	} else {
		node.GetLog().Error("Connector registry doesn't have httpserv instance")
		return fmt.Errorf("can't find httpserv connector")
	}

	if node.config.IsWs {
		// Flow shares single WS connection.
		node.nodeGlobalId = node.FlowOpCtx().FlowId
	} else {
		// Flow can have multiple rest endpoints.
		node.nodeGlobalId = node.FlowOpCtx().FlowId + "_" + string(node.GetMetaNode().Id)
	}

	if node.config.IsWs && node.config.IsPublishOnly {
		node.httpServerConn.RegisterFlow(node.nodeGlobalId, true, true, true, nil)
	}

	return err
}

func (node *Node) WaitForEvent(responseChannel chan model.ReactorEvent) {
}

func (node *Node) OnInput(msg *model.Message) ([]model.NodeID, error) {
	node.GetLog().Info(" Executing Node . Name = ", node.Meta().Label)
	var body []byte
	if node.config.InputVar.Name != "" {
		variable, err := node.ctx.GetVariable(node.config.InputVar.Name, node.FlowOpCtx().FlowId)
		if err != nil {
			node.GetLog().Error("Can't get variable . Error:", err)
			return nil, err
		}

		if variable.ValueType == "string" {
			body = []byte(variable.Value.(string))
		} else {
			body, err = json.Marshal(variable.Value)
			if err != nil {
				node.GetLog().Error("Can't marshal json . Error:", err)
				return []model.NodeID{node.Meta().ErrorTransition}, err
			}
		}
	} else {
		body, _ = msg.Payload.SerializeToJson()
	}

	contentType := "application/json"

	if node.config.ResponsePayloadFormat == "html" {
		contentType = "text/html"
	}
	if node.config.IsWs {
		node.httpServerConn.PublishWs(node.nodeGlobalId, body)
	} else {
		node.httpServerConn.ReplyToRequest(msg.RequestId, body, contentType)
	}

	return []model.NodeID{node.Meta().SuccessTransition}, nil
}
