package influx

import (
	"errors"
	"github.com/alivinco/tpflow/model"
	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/mitchellh/mapstructure"
	"text/template"
)

type InfluxdbReadNode struct {
	BaseNode
	ctx *model.Context
	//transport *fimpgo.MqttTransport
	config        InfluxdbReadConfig
	queryTemplate *template.Template
	connection    influx.Client
}

type InfluxdbReadConfig struct {
	ConnectorId        string
	QueryExpression    string
	ResultVariableName string
	IsLVariableGlobal  bool
}

func NewInfluxdbReadNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *model.Context) model.Node {
	node := InfluxdbReadNode{ctx: ctx}
	node.meta = meta
	node.flowOpCtx = flowOpCtx
	node.config = InfluxdbReadConfig{}
	node.SetupBaseNode()
	return &node
}

func (node *InfluxdbReadNode) LoadNodeConfig() error {
	err := mapstructure.Decode(node.meta.Config, &node.config)
	if err != nil {
		node.GetLog().Error("Can't decode config.Err:", err)
		return err

	}

	connInstance := node.connectorRegistry.GetInstance(node.config.ConnectorId)
	var ok bool
	if connInstance != nil {
		node.connection, ok = connInstance.Connection.(influx.Client)
		if !ok {
			node.connection = nil
			node.GetLog().Error("Can't get things connection to things registry . Cast to influx client failed")
		}
	} else {
		node.GetLog().Error("Connector registry doesn't have influxdb instance")
	}

	funcMap := template.FuncMap{
		"variable": func(varName string, isGlobal bool) (interface{}, error) {
			//node.GetLog().Debug("Getting variable by name ",varName)
			var vari model.Variable
			var err error
			if isGlobal {
				vari, err = node.ctx.GetVariable(varName, "global")
			} else {
				vari, err = node.ctx.GetVariable(varName, node.flowOpCtx.FlowId)
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
	node.queryTemplate, err = template.New("query").Funcs(funcMap).Parse(node.meta.Address)
	if err != nil {
		node.GetLog().Error(" Failed while parsing url template.Error:", err)
	}
	return err
}

func (node *InfluxdbReadNode) WaitForEvent(responseChannel chan model.ReactorEvent) {

}

func (node *InfluxdbReadNode) OnInput(msg *model.Message) ([]model.NodeID, error) {
	node.GetLog().Info("Executing InfluxdbReadNode . Name = ", node.meta.Label)

	//msgBa, err := fimpMsg.SerializeToJson()
	//if err != nil {
	//	return nil,err
	//}
	//var addrTemplateBuffer bytes.Buffer
	//node.queryTemplate.Execute(&addrTemplateBuffer,nil)
	//address:= addrTemplateBuffer.String()
	//
	//
	//node.GetLog().Debug(" Address: ",address)
	//node.GetLog().Debug(" Action message : ", fimpMsg)

	//node.transport.PublishRaw(address, msgBa)
	return []model.NodeID{node.meta.SuccessTransition}, nil
}
