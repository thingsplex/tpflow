package node

import (
	"github.com/alivinco/tpflow/model"
	"github.com/mitchellh/mapstructure"
)

type LoopNode struct {
	BaseNode
	ctx       *model.Context
	config    LoopNodeConfig
	counter   int64
	countUp   bool
}

type LoopNodeConfig struct {
	StartValue     int64
	EndValue       int64
	Step           int64
	SaveToVariable bool
}

func NewLoopNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *model.Context) model.Node {
	node := LoopNode{ctx: ctx}
	node.meta = meta
	node.flowOpCtx = flowOpCtx
	node.SetupBaseNode()
	return &node
}

func (node *LoopNode) LoadNodeConfig() error {
	defValue := LoopNodeConfig{}
	err := mapstructure.Decode(node.meta.Config, &defValue)
	if err != nil {
		node.getLog().Error("Can't decode configuration", err)
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

func (node *LoopNode) WaitForEvent(responseChannel chan model.ReactorEvent) {

}

func (node *LoopNode) OnInput(msg *model.Message) ([]model.NodeID, error) {
	node.getLog().Debug("Executing LoopNode . Name = ", node.meta.Label)
	if node.countUp {
		node.counter = node.counter + node.config.Step
	} else {
		node.counter = node.counter - node.config.Step
	}
	node.ctx.SetVariable("loop_counter","int",node.counter,"auto",node.flowOpCtx.FlowId,true)
	node.getLog().Debug("Counter = ", node.counter)
	if (node.countUp && node.counter >= node.config.EndValue) || (!node.countUp && node.counter <= node.config.EndValue) {
		node.counter = node.config.StartValue
		return []model.NodeID{node.meta.ErrorTransition}, nil
	}
	return []model.NodeID{node.meta.SuccessTransition}, nil
}
