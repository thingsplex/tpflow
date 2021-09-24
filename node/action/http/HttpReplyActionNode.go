package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/thingsplex/tpflow/connector/plugins/http"
	"github.com/thingsplex/tpflow/flow/context"
	"github.com/thingsplex/tpflow/model"
	"github.com/thingsplex/tpflow/node/base"
	"text/template"
)

// Node is Http reply node that sends response to request received from HTTP trigger node
type Node struct {
	base.BaseNode
	ctx            *context.Context
	config         NodeConfig
	httpServerConn *http.Connector
	nodeGlobalId   string
	responseTemplate   *template.Template
}

type NodeConfig struct {
	ConnectorID           string
	InputVar              model.NodeVariableDef
	ResponsePayloadFormat string // json,fimp,xml,html
	IsWs                  bool   // message is sent over active WS connection
	IsPublishOnly         bool   // true - means there is no trigger node , only publish
	Alias                 string // alias that will be used in url instead of flowId
	ResponseTemplate      string // content of template will be used if input variable is not set .
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
		name := fmt.Sprintf("%s %s", node.FlowOpCtx().FlowMeta.Name,node.BaseNode.GetMetaNode().Label)
		node.httpServerConn.RegisterFlow(node.nodeGlobalId, true, true, true, nil,node.config.Alias,name,http.AuthConfig{})
	}

	if node.config.ResponseTemplate != "" {
		funcMap := template.FuncMap{
			"variable": func(varName string, isGlobal bool) (interface{}, error) {
				var vari context.Variable
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
		node.responseTemplate, err = template.New("transform").Funcs(funcMap).Parse(node.config.ResponseTemplate)
		if err != nil {
			node.GetLog().Error(" Failed while parsing request template.Error:", err)
			return err
		}
	}

	return err
}

func (node *Node) WaitForEvent(responseChannel chan model.ReactorEvent) {
}

func (node *Node) OnInput(msg *model.Message) ([]model.NodeID, error) {
	node.GetLog().Debug(" Executing Node . Name = ", node.Meta().Label)
	var body []byte

	if node.config.ResponseTemplate == "" {
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
			msg.Payload.Topic = msg.AddressStr
			if msg.ValPayload == nil {
				body, _ = msg.Payload.SerializeToJson()
			}else {
				body, _ = json.Marshal(msg.ValPayload)
			}
		}
	} else {
		var templateBuffer bytes.Buffer
		var template = struct {
			Variable interface{} // Available from template
			FlowId string
			NodeId model.NodeID
			NodeLabel string
			TunId string
			TunToken string
		}{Variable: msg.Payload.Value,FlowId: node.FlowOpCtx().FlowId,NodeId: node.GetMetaNode().Id,
			NodeLabel: node.GetMetaNode().Label,
			TunId: node.httpServerConn.Config().TunAddress,
			TunToken: node.httpServerConn.Config().TunEdgeToken}
		node.responseTemplate.Execute(&templateBuffer, template)
		body = templateBuffer.Bytes()
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
