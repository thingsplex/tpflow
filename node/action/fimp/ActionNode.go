package fimp

import (
	"bytes"
	"errors"
	"github.com/futurehomeno/fimpgo"
	"github.com/mitchellh/mapstructure"
	"github.com/thingsplex/tpflow/model"
	"github.com/thingsplex/tpflow/node/base"
	"text/template"
)

type Node struct {
	base.BaseNode
	ctx             *model.Context
	transport       *fimpgo.MqttTransport
	config          NodeConfig
	addressTemplate *template.Template
}

type NodeConfig struct {
	DefaultValue             model.Variable
	VariableName             string
	VariableType             string
	IsVariableGlobal         bool
	Props                    fimpgo.Props
	RegisterAsVirtualService bool
	VirtualServiceGroup      string
	VirtualServiceProps      map[string]interface{} // mostly used to announce supported features of the service , for instance supported modes , states , setpoints , etc...
	ResponseToTopic          string // in request-response communication requester can set topic to which server will send response
	IsResponseToReq          bool // Indicates that the action is response to request and corid field should be set to the same value as request uuid
	ConnectorID              string
	ReplyToReq               bool
	ForwardInputMessage      bool
}

func NewNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *model.Context) model.Node {
	node := Node{ctx: ctx}
	node.SetMeta(meta)
	node.SetFlowOpCtx(flowOpCtx)
	node.config = NodeConfig{DefaultValue: model.Variable{}}
	node.SetupBaseNode()
	return &node
}

func (node *Node) LoadNodeConfig() error {
	err := mapstructure.Decode(node.Meta().Config, &node.config)
	if err != nil {
		node.GetLog().Error("Can't decode config.Err:", err)

	}

	funcMap := template.FuncMap{
		"variable": func(varName string, isGlobal bool) (interface{}, error) {
			//node.GetLog().Debug("Getting variable by name ",varName)
			var vari model.Variable
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
				node.GetLog().Debug("Only simple types are supported ")
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
	node.addressTemplate, err = template.New("address").Funcs(funcMap).Parse(node.Meta().Address)
	if err != nil {
		node.GetLog().Error(" Failed while parsing url template.Error:", err)
	}
	if node.config.ConnectorID == "" {
		node.config.ConnectorID = "fimpmqtt"
	}
	fimpTransportInstance := node.ConnectorRegistry().GetInstance(node.config.ConnectorID)
	var ok bool
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

func (node *Node) WaitForEvent(responseChannel chan model.ReactorEvent) {

}

func (node *Node) OnInput(msg *model.Message) ([]model.NodeID, error) {
	node.GetLog().Info("Executing Node . Name = ", node.Meta().Label)
	//node.GetLog().Infof("Request uuid %s",msg.Payload.UID)
	//fimpMsg := fimpgo.FimpMessage{Type: node.Meta().ServiceInterface, Service: node.Meta().Service, Properties: node.config.Props}
	fimpMsg := fimpgo.NewMessage(node.Meta().ServiceInterface,node.Meta().Service,"",nil,node.config.Props,nil,nil)
	fimpMsg.Source = "flow_"+node.FlowOpCtx().FlowId
	fimpMsg.ResponseToTopic = node.config.ResponseToTopic
	if node.config.IsResponseToReq {
		fimpMsg.CorrelationID = msg.Payload.UID
	}

	if fimpMsg.Version == "" {
		fimpMsg.Version = "1"
	}
	if node.config.VariableName != "" {
		node.GetLog().Debug("Using variable as input :", node.config.VariableName)
		flowId := node.FlowOpCtx().FlowId
		if node.config.IsVariableGlobal {
			flowId = "global"
		}
		variable, err := node.ctx.GetVariable(node.config.VariableName, flowId)
		if err != nil {
			node.GetLog().Error("Can't get variable . Error:", err)
			return nil, err
		}
		fimpMsg.ValueType = variable.ValueType
		fimpMsg.Value = variable.Value
	} else {
		if node.config.DefaultValue.Value == "" || node.config.DefaultValue.ValueType == "" {
			fimpMsg.Value = msg.Payload.Value
			fimpMsg.ValueType = msg.Payload.ValueType
		} else {
			fimpMsg.Value = node.config.DefaultValue.Value
			fimpMsg.ValueType = node.config.DefaultValue.ValueType
		}
	}
	var address string
	if node.config.ReplyToReq && msg.Payload.ResponseToTopic != "" {
		address = msg.Payload.ResponseToTopic
		fimpMsg.CorrelationID = msg.Payload.UID
	}else {
		var addrTemplateBuffer bytes.Buffer
		node.addressTemplate.Execute(&addrTemplateBuffer, nil)
		address = addrTemplateBuffer.String()
	}

	node.GetLog().Debug(" Publishing to: ", address)
	node.transport.PublishToTopic(address,fimpMsg)
	return []model.NodeID{node.Meta().SuccessTransition}, nil
}
