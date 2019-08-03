package loop

import (
	"github.com/thingsplex/tpflow/model"
	"github.com/thingsplex/tpflow/node/base"
	"github.com/mitchellh/mapstructure"
)

//Node
type Node struct {
	base.BaseNode
	ctx     *model.Context
	config  NodeConfig
	counter int64
	countUp bool
}

type NodeConfig struct {
	StartValue     int64
	EndValue       int64
	Step           int64
	SaveToVariable bool
}

func NewNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *model.Context) model.Node {
	node := Node{ctx: ctx}
	node.SetMeta(meta)
	node.SetFlowOpCtx(flowOpCtx)
	node.SetupBaseNode()
	return &node
}

func (node *Node) LoadNodeConfig() error {
	defValue := NodeConfig{}
	err := mapstructure.Decode(node.Meta().Config, &defValue)
	if err != nil {
		node.GetLog().Error("Can't decode configuration", err)
	} else {
		node.config = defValue
		if node.config.Step == 0 {
			node.config.Step = 1
		}
		if defValue.EndValue > defValue.StartValue {
			node.countUp = true
			if node.config.Step > defValue.EndValue {
				node.config.Step = defValue.EndValue
			}
		} else {
			if node.config.Step < defValue.EndValue {
				node.config.Step = defValue.EndValue
			}
		}
	}
	return nil
}

func (node *Node) WaitForEvent(responseChannel chan model.ReactorEvent) {

}

func (node *Node) OnInput(msg *model.Message) ([]model.NodeID, error) {
	node.GetLog().Debug("Executing Node . Name = ", node.Meta().Label)
	if node.countUp {
		node.counter = node.counter + node.config.Step
	} else {
		node.counter = node.counter - node.config.Step
	}
	node.ctx.SetVariable("loop_counter", "int", node.counter, "auto", node.FlowOpCtx().FlowId, true)
	node.GetLog().Debug("Counter = ", node.counter)
	if (node.countUp && node.counter >= node.config.EndValue) || (!node.countUp && node.counter <= node.config.EndValue) {
		node.counter = node.config.StartValue
		return []model.NodeID{node.Meta().ErrorTransition}, nil
	}
	return []model.NodeID{node.Meta().SuccessTransition}, nil
}
