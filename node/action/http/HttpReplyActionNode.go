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
	ctx    *model.Context
	config NodeConfig
	httpServerConn *http.Connector
}

type NodeConfig struct {
	ConnectorID      string
	VariableName     string
	VariableType     string
	IsVariableGlobal bool
	ResponsePayloadFormat string // json,xml,html
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
	return err
}

func (node *Node) WaitForEvent(responseChannel chan model.ReactorEvent) {
}

func (node *Node) OnInput(msg *model.Message) ([]model.NodeID, error) {
	node.GetLog().Info(" Executing Node . Name = ", node.Meta().Label)

	variable, err := node.ctx.GetVariable(node.config.VariableName, node.FlowOpCtx().FlowId)
	if err != nil {
		node.GetLog().Error("Can't get variable . Error:", err)
		return nil, err
	}
	var body []byte
	if variable.ValueType == "string" {
		body = []byte(variable.Value.(string))
	}else {
		body,err = json.Marshal(variable.Value)
		if err != nil {
			node.GetLog().Error("Can't marshal json . Error:", err)
			return []model.NodeID{node.Meta().ErrorTransition}, err
		}
	}

	contentType := "application/json"

	if node.config.ResponsePayloadFormat == "html" {
		contentType = "text/html"
	}

	node.httpServerConn.ReplyToRequest(msg.RequestId,body,contentType)

	return []model.NodeID{node.Meta().SuccessTransition}, nil
}
