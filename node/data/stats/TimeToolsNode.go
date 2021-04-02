package stats

import (
	"github.com/mitchellh/mapstructure"
	"github.com/thingsplex/tpflow/model"
	"github.com/thingsplex/tpflow/node/base"
	"time"
)

// Measures time between invocations
type TimeToolsNode struct {
	base.BaseNode
	ctx        *model.Context
	nodeConfig TimeToolsNodeConfig
	timeStamp  time.Time
}

// Save elapsed time into variable
type TimeToolsNodeConfig struct {
	Operation              string // stopwatch - time between invocations , now - current time in seconds
	TargetVariableName     string // Variable to save result
	IsTargetVariableGlobal bool
}

func NewTimeToolsNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *model.Context) model.Node {
	node := TimeToolsNode{ctx: ctx}
	node.SetMeta(meta)
	node.SetFlowOpCtx(flowOpCtx)
	node.SetupBaseNode()
	return &node
}

func (node *TimeToolsNode) LoadNodeConfig() error {
	defValue := TimeToolsNodeConfig{}
	node.timeStamp = time.Now()
	err := mapstructure.Decode(node.Meta().Config, &defValue)
	if err != nil {
		node.GetLog().Error(" Can't decode configuration", err)
	} else {
		node.nodeConfig = defValue
	}
	return nil
}

func (node *TimeToolsNode) OnInput(msg *model.Message) ([]model.NodeID, error) {
	node.GetLog().Debugf(" Executing TimeToolsNode . Name = %s", node.Meta().Label)

	var result int64
	var resultStr string

	switch node.nodeConfig.Operation {
	case "stopwatch":
		result = int64(time.Now().Sub(node.timeStamp) / time.Second) // time in seconds
	case "stopwatch_monotonic":
		//TODO : implement
	case "now_full":
		result = time.Now().Unix() // time in seconds
	case "now_full_RFC3339":
		resultStr = time.Now().Format(time.RFC3339) // time in seconds
	case "now_year":
		result = int64(time.Now().Year())
	case "now_month":
		result = int64(time.Now().Month())
	case "now_day":
		result = int64(time.Now().Day())
	case "now_hour":
		result = int64(time.Now().Hour())
	case "now_min":
		result = int64(time.Now().Minute())
	case "now_sec":
		result = int64(time.Now().Second())
	}

	// Save input value to variable
	var err error

	if node.nodeConfig.Operation == "now_full_RFC3339" {
		if node.nodeConfig.IsTargetVariableGlobal {
			err = node.ctx.SetVariable(node.nodeConfig.TargetVariableName, "string", resultStr, "", "global", true)
		} else {
			err = node.ctx.SetVariable(node.nodeConfig.TargetVariableName, "string", resultStr, "", node.FlowOpCtx().FlowId, true)
		}
	}else {
		if node.nodeConfig.IsTargetVariableGlobal {
			err = node.ctx.SetVariable(node.nodeConfig.TargetVariableName, "int", result, "", "global", true)
		} else {
			err = node.ctx.SetVariable(node.nodeConfig.TargetVariableName, "int", result, "", node.FlowOpCtx().FlowId, true)
		}
	}


	node.timeStamp = time.Now()

	return []model.NodeID{node.Meta().SuccessTransition}, err
}

func (node *TimeToolsNode) WaitForEvent(responseChannel chan model.ReactorEvent) {

}
