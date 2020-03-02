package wait

import (
	"github.com/thingsplex/tpflow/node/base"
	"time"
)
import (
	"github.com/futurehomeno/fimpgo"
	"github.com/thingsplex/tpflow/model"
)

type WaitNode struct {
	base.BaseNode
	delay     int
	ctx       *model.Context
	transport *fimpgo.MqttTransport
}

func NewWaitNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *model.Context) model.Node {
	node := WaitNode{ctx: ctx}
	node.SetMeta(meta)
	node.SetFlowOpCtx(flowOpCtx)
	return &node
}

func (node *WaitNode) LoadNodeConfig() error {
	delay, ok := node.Meta().Config.(float64)
	if ok {
		node.delay = int(delay)
	} else {
		node.GetLog().Error(" Can't cast Wait node delay value")
	}
	return nil
}

func (node *WaitNode) WaitForEvent(nodeEventStream chan model.ReactorEvent) {

}

func (node *WaitNode) OnInput(msg *model.Message) ([]model.NodeID, error) {
	node.GetLog().Info(" Waiting  for = ", node.delay)
	timer := time.NewTimer(time.Millisecond * time.Duration(node.delay))
	select {
	case  <- timer.C:
		return []model.NodeID{node.Meta().SuccessTransition}, nil
	case signal := <-node.FlowOpCtx().NodeControlSignalChannel:
		timer.Stop()
		node.GetLog().Debug("Control signal ")
		if signal == model.SIGNAL_STOP {
			return nil,nil
		}
	}
	return []model.NodeID{node.Meta().SuccessTransition}, nil
}
