package fimp

import (
	"github.com/futurehomeno/fimpgo"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tpflow/model"
	"github.com/thingsplex/tpflow/node/base"
)

type LogNode struct {
	base.BaseNode
	ctx             *model.Context
	transport       *fimpgo.MqttTransport
	config          LogNodeConfig
}

type LogNodeConfig struct {
	VariableName             string
	IsVariableGlobal         bool
	LogLevel                 string
	Text                     string
}

func NewNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *model.Context) model.Node {
	node := LogNode{ctx: ctx}
	node.SetMeta(meta)
	node.SetFlowOpCtx(flowOpCtx)
	node.config = LogNodeConfig{}
	node.SetupBaseNode()
	return &node
}

func (node *LogNode) LoadNodeConfig() error {
	err := mapstructure.Decode(node.Meta().Config, &node.config)
	if err != nil {
		node.GetLog().Error("Can't decode config.Err:", err)
	}
	return err
}

func (node *LogNode) WaitForEvent(responseChannel chan model.ReactorEvent) {

}

func (node *LogNode) OnInput(msg *model.Message) ([]model.NodeID, error) {
	level , err := log.ParseLevel(node.config.LogLevel)
	if err != nil {
		level,_ = log.ParseLevel("info")
	}
	if node.config.Text != "" {
		node.GetLog().Logf(level,node.config.Text)
	}else if node.config.VariableName != "" {

		flowId := node.FlowOpCtx().FlowId
		if node.config.IsVariableGlobal {
			flowId = "global"
		}
		variable, err := node.ctx.GetVariable(node.config.VariableName, flowId)
		if err != nil {
			node.GetLog().Error("Can't get variable . Error:", err)
			return nil, err
		}
		node.GetLog().Logf(level,"%+v",variable)
	} else {
		node.GetLog().Logf(level,"%+v",msg.Payload)
	}

	return []model.NodeID{node.Meta().SuccessTransition}, nil
}
