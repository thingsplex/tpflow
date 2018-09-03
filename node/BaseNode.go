package node

import (
	log "github.com/Sirupsen/logrus"
	"github.com/alivinco/fimpgo"
	"github.com/alivinco/tpflow/model"
	"github.com/alivinco/tpflow/shared"
)

type BaseNode struct {

	meta model.MetaNode
	ctx *model.Context
	flowOpCtx *model.FlowOperationalContext
	isStartNode bool   // true - if node is first in a flow
	isMsgReactor bool  // true - node reacts on messages and requires input stream .
	isReactorRunning bool
	transport *fimpgo.MqttTransport
	logFields log.Fields
	sharedResources *shared.GlobalSharedResources
}

func (node *BaseNode) SetupBaseNode() {
	node.logFields = log.Fields{"comp":"fnode","ntype":node.meta.Type,"fid":node.flowOpCtx.FlowId,"nid":node.meta.Id}
}

func (node *BaseNode) getLog() *log.Entry {
	return log.WithFields(node.logFields)
}

func (node *BaseNode) GetMetaNode()*model.MetaNode {
	return &node.meta
}
func (node *BaseNode) GetNextSuccessNodes()[]model.NodeID {
	return []model.NodeID{node.meta.SuccessTransition}
}

func (node *BaseNode) GetNextErrorNode()model.NodeID {
	return node.meta.ErrorTransition
}

func (node *BaseNode) GetNextTimeoutNode()model.NodeID{
	return node.meta.TimeoutTransition
}

func (node *BaseNode) IsStartNode() bool {
	return node.isStartNode
}

func (node *BaseNode) IsMsgReactorNode() bool {
	return node.isMsgReactor
}

func (node *BaseNode) IsReactorRunning() bool {
	return node.isReactorRunning
}

// is invoked when node is started
func (node *BaseNode) Init() error {
	return nil
}

// is invoked when node flow is stopped
func (node *BaseNode) Cleanup() error {
	return nil
}

func (node *BaseNode) ConfigureInStream(activeSubscriptions *[]string,msgInStream model.MsgPipeline) {
}

func (node *BaseNode) SetSharedResources(sharedResource *shared.GlobalSharedResources) {
	node.sharedResources = sharedResource
}