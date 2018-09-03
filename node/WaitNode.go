package node

import "time"
import (
	"github.com/alivinco/fimpgo"
	"github.com/alivinco/tpflow/model"
)

type WaitNode struct {
	BaseNode
	ctx *model.Context
	transport *fimpgo.MqttTransport
}

func NewWaitNode(flowOpCtx *model.FlowOperationalContext,meta model.MetaNode,ctx *model.Context,transport *fimpgo.MqttTransport) model.Node {
	node := WaitNode{ctx:ctx,transport:transport}
	node.meta = meta
	node.flowOpCtx  = flowOpCtx
	return &node
}

func (node *WaitNode) LoadNodeConfig() error {
	delay ,ok := node.meta.Config.(float64)
	if ok {
		node.meta.Config = int(delay)
	}else {
		node.getLog().Error(" Can't cast Wait node delay value")
	}

	return nil
}

func (node *WaitNode) WaitForEvent(nodeEventStream chan model.ReactorEvent) {

}

func (node *WaitNode) OnInput( msg *model.Message) ([]model.NodeID,error) {
	delayMilisec, ok := node.meta.Config.(int)
	if ok {
		node.getLog().Info(" Waiting  for = ", delayMilisec)
		time.Sleep(time.Millisecond * time.Duration(delayMilisec))
	} else {
		node.getLog().Error(" Wrong time format")
	}
	return []model.NodeID{node.meta.SuccessTransition},nil
}


