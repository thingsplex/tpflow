package fimp

import (
	"bytes"
	"errors"
	"github.com/alivinco/fimpgo"
	"github.com/alivinco/tpflow/model"
	"github.com/alivinco/tpflow/node/base"
	"github.com/mitchellh/mapstructure"
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
	}
	node.addressTemplate, err = template.New("address").Funcs(funcMap).Parse(node.Meta().Address)
	if err != nil {
		node.GetLog().Error(" Failed while parsing url template.Error:", err)
	}

	fimpTransportInstance := node.ConnectorRegistry().GetInstance("fimpmqtt")
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
	fimpMsg := fimpgo.FimpMessage{Type: node.Meta().ServiceInterface, Service: node.Meta().Service, Properties: node.config.Props}
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

	msgBa, err := fimpMsg.SerializeToJson()
	if err != nil {
		return nil, err
	}
	var addrTemplateBuffer bytes.Buffer
	node.addressTemplate.Execute(&addrTemplateBuffer, nil)
	address := addrTemplateBuffer.String()
	node.GetLog().Debug(" Address: ", address)
	node.GetLog().Debug(" Action message : ", fimpMsg)
	node.transport.PublishRaw(address, msgBa)
	return []model.NodeID{node.Meta().SuccessTransition}, nil
}
