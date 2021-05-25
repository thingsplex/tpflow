package stats

import (
	"github.com/mitchellh/mapstructure"
	"github.com/thingsplex/tpflow/flow/context"
	"github.com/thingsplex/tpflow/model"
	"github.com/thingsplex/tpflow/node/base"
	"time"
)

type MeterNode struct {
	base.BaseNode
	ctx              *context.Context
	nodeConfig       MeterNodeConfig
	lastInputEventTs time.Time
	lastReportTs     time.Time
	accumulatedValue float64
	lastInputValue   float64
}

type MeterNodeConfig struct {
	Type                string // pulse , accumulator
	PulseValue          float64
	ReportTriggerType   string  // value - value increases by x% , time - max time passed since last report
	TrigMaxTimeInterval int64   // Trigger Max time since last report ,
	TrigMaxAccumValue   float64 // Trigger Max accumulated value
	ReportType          string  // monotonic , sampled
	Description         string  // documentation
	InputVar            model.NodeVariableDef // variable holds changes in flow rate (instant power , water speed of flow )
	OutputVar           model.NodeVariableDef // variable that is used to store accumulated result
	ResetValue          float64 // Max value after which meter resets to 0
}

// Input can be either impulse signal which will work along with fixed rate ( constant power usage 100W) or value (W) . Node emits accumulated value either by interval or by once
// goes over certain limit (for instance 10%) . Every time input value changes , the node calculates usage and adds value to internal accumulator.

func NewMeterNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *context.Context) model.Node {
	node := MeterNode{ctx: ctx}
	node.SetMeta(meta)
	node.SetFlowOpCtx(flowOpCtx)
	node.SetupBaseNode()
	return &node
}

func (node *MeterNode) LoadNodeConfig() error {
	defValue := MeterNodeConfig{}
	err := mapstructure.Decode(node.Meta().Config, &defValue)
	if err != nil {
		node.GetLog().Error(" Can't decode configuration", err)
	} else {
		node.nodeConfig = defValue
	}
	return nil
}

func (node *MeterNode) OnInput(msg *model.Message) ([]model.NodeID, error) {
	node.GetLog().Debugf(" Executing MeterNode . Name = %s", node.Meta().Label)

	return []model.NodeID{node.Meta().SuccessTransition}, nil
}

func (node *MeterNode) WaitForEvent(responseChannel chan model.ReactorEvent) {

}
