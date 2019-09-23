package setvar

import (
	"github.com/thingsplex/tpflow/model"
	"github.com/thingsplex/tpflow/node/base"
	"github.com/mitchellh/mapstructure"
)

type SetVariableNode struct {
	base.BaseNode
	ctx        *model.Context
	nodeConfig SetVariableNodeConfig
}

type SetVariableNodeConfig struct {
	Name               string
	Description        string
	UpdateGlobal       bool // true - update global variable ; false - update local variable
	UpdateInputMsg     bool // true - update input message  ; false - update context variable
	IsVariableInMemory bool // true - is saved on disk ; false - in memory only
	DefaultValue       model.Variable
}

func NewSetVariableNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *model.Context) model.Node {
	node := SetVariableNode{ctx: ctx}
	node.SetMeta(meta)
	node.SetFlowOpCtx(flowOpCtx)
	node.SetupBaseNode()
	return &node
}

func (node *SetVariableNode) LoadNodeConfig() error {
	defValue := SetVariableNodeConfig{}
	err := mapstructure.Decode(node.Meta().Config, &defValue)
	if err != nil {
		node.GetLog().Error(" Can't decode configuration", err)
	} else {
		node.nodeConfig = defValue
	}
	return nil
}

func (node *SetVariableNode) OnInput(msg *model.Message) ([]model.NodeID, error) {
	node.GetLog().Info(" Executing SetVariableNode . Name = ", node.Meta().Label)

	if node.nodeConfig.UpdateInputMsg {
		// Update input value with value from node config .
		msg.Payload.Value = node.nodeConfig.DefaultValue.Value
		msg.Payload.ValueType = node.nodeConfig.DefaultValue.ValueType
	} else {
		// Save input value to variable
		if node.nodeConfig.DefaultValue.ValueType == "" {
			if node.nodeConfig.UpdateGlobal {
				node.ctx.SetVariable(node.nodeConfig.Name, msg.Payload.ValueType, msg.Payload.Value, node.nodeConfig.Description, "global", node.nodeConfig.IsVariableInMemory)
			} else {
				node.ctx.SetVariable(node.nodeConfig.Name, msg.Payload.ValueType, msg.Payload.Value, node.nodeConfig.Description, node.FlowOpCtx().FlowId, node.nodeConfig.IsVariableInMemory)
			}
		} else {
			// Save default value from node config to variable
			if node.nodeConfig.UpdateGlobal {
				node.ctx.SetVariable(node.nodeConfig.Name, node.nodeConfig.DefaultValue.ValueType, node.nodeConfig.DefaultValue.Value, node.nodeConfig.Description, "global", node.nodeConfig.IsVariableInMemory)
			} else {
				node.ctx.SetVariable(node.nodeConfig.Name, node.nodeConfig.DefaultValue.ValueType, node.nodeConfig.DefaultValue.Value, node.nodeConfig.Description, node.FlowOpCtx().FlowId, node.nodeConfig.IsVariableInMemory)
			}
		}
	}
	return []model.NodeID{node.Meta().SuccessTransition}, nil
}

func (node *SetVariableNode) WaitForEvent(responseChannel chan model.ReactorEvent) {

}
