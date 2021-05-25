package stats

import (
	"github.com/mitchellh/mapstructure"
	"github.com/thingsplex/tpflow/flow/context"
	"github.com/thingsplex/tpflow/model"
	"github.com/thingsplex/tpflow/node/base"
)

type MetricsNode struct {
	base.BaseNode
	ctx        *context.Context
	nodeConfig MetricsNodeConfig
}

type MetricsNodeConfig struct {
	Operation   string                //  inc,dec, add(float64),sub(float64)
	Step        float64               // step for inc and dec operations. Default value = 1
	InputVar    model.NodeVariableDef // argument for add and sub will be taken from the variable , if empty , value is taken from input message
	OutputVar   model.NodeVariableDef // Variable to save result
	MaxValue    float64               // mac
	MinValue    float64               //
	ResetValue  float64
	LimitAction string // reset - reset to ResetValue , keep_limit - keep border value , no_limits - value is not limited
}

func NewMetricsNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *context.Context) model.Node {
	node := MetricsNode{ctx: ctx}
	node.SetMeta(meta)
	node.SetFlowOpCtx(flowOpCtx)
	node.SetupBaseNode()
	return &node
}

func (node *MetricsNode) LoadNodeConfig() error {
	defValue := MetricsNodeConfig{}
	err := mapstructure.Decode(node.Meta().Config, &defValue)
	if err != nil {
		node.GetLog().Error(" Can't decode configuration", err)
	} else {
		node.nodeConfig = defValue
	}
	return nil
}

func (node *MetricsNode) OnInput(msg *model.Message) ([]model.NodeID, error) {
	var value float64
	var err error
	node.GetLog().Debugf(" Executing MetricsNode . Name = %s", node.Meta().Label)


	inputVarFlowId := "global"
	if !node.nodeConfig.InputVar.IsGlobal {
		inputVarFlowId = node.FlowOpCtx().FlowId
	}

	targetVarFlowId := "global"
	if !node.nodeConfig.OutputVar.IsGlobal {
		targetVarFlowId = node.FlowOpCtx().FlowId
	}

	valVar, err := node.ctx.GetVariable(node.nodeConfig.OutputVar.Name, targetVarFlowId)
	if err != nil {
		node.GetLog().Debug("Value variable doesn't exist , configuring default limit. Err:", err.Error())
		value = node.nodeConfig.ResetValue
		err = nil
	}else {
		value,err = valVar.ToNumber()
	}


	switch node.nodeConfig.Operation {
	case "inc":
		value += node.nodeConfig.Step
	case "dec":
		value -= node.nodeConfig.Step
	case "add":
		fallthrough
	case "sub":
		fallthrough
	case "set":
		var inputVar context.Variable
		if node.nodeConfig.InputVar.Name != "" {
			inputVar, err = node.ctx.GetVariable(node.nodeConfig.InputVar.Name, inputVarFlowId)
			if err != nil {
				node.GetLog().Error("Variable doesn't exist . Err:", err.Error())
				return []model.NodeID{node.Meta().ErrorTransition}, err
			}
		}else {
			inputVar.Value = msg.Payload.Value
			inputVar.ValueType = msg.Payload.ValueType
		}

		if inputVar.IsNumber() {
			nVar, err := inputVar.ToNumber()
			if err != nil {
				node.GetLog().Error("Variable is not a number . Err:", err.Error())
				return []model.NodeID{node.Meta().ErrorTransition}, err
			}
			switch node.nodeConfig.Operation {
			case "add":
				value += nVar
			case "sub":
				value -= nVar
			case "set":
				value = nVar
			}
		}
	}

	if value > node.nodeConfig.MaxValue {
		if node.nodeConfig.LimitAction == "reset" {
			//TODO : maybe transition to red/fail node
			value = node.nodeConfig.ResetValue
		} else if node.nodeConfig.LimitAction == "keep_limit" {
			value = node.nodeConfig.MaxValue
		}
	}

	if value < node.nodeConfig.MinValue {
		if node.nodeConfig.LimitAction == "reset" {
			value = node.nodeConfig.ResetValue
		} else if node.nodeConfig.LimitAction == "keep_limit" {
			value = node.nodeConfig.MinValue
		}
	}
	//node.GetLog().Debugf("Output var name = %s , flow = %s",node.nodeConfig.OutputVar.Name,targetVarFlowId)
	err = node.ctx.SetVariable(node.nodeConfig.OutputVar.Name, "float", value, "", targetVarFlowId, node.nodeConfig.OutputVar.InMemory)
	if err != nil {
		node.GetLog().Error("Can't save output variable . Err:", err.Error())
		return []model.NodeID{node.Meta().ErrorTransition}, err
	}
	return []model.NodeID{node.Meta().SuccessTransition}, err
}

func (node *MetricsNode) WaitForEvent(responseChannel chan model.ReactorEvent) {

}
