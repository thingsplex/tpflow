package base

import (
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tpflow/connector"
	"github.com/thingsplex/tpflow/model"
)

type BaseNode struct {
	meta              model.MetaNode
	ctx               *model.Context
	flowOpCtx         *model.FlowOperationalContext
	isStartNode       bool // true - if node is first in a flow
	isMsgReactor      bool // true - node reacts on messages and requires input stream .
	isReactorRunning  bool
	logFields         log.Fields
	connectorRegistry *connector.Registry
	flowRunner        model.FlowRunner
}

func (node *BaseNode) FlowRunner() model.FlowRunner {
	return node.flowRunner
}

func (node *BaseNode) SetReactorRunning(isReactorRunning bool) {
	node.isReactorRunning = isReactorRunning
}

func (node *BaseNode) ConnectorRegistry() *connector.Registry {
	return node.connectorRegistry
}

func (node *BaseNode) Meta() model.MetaNode {
	return node.meta
}

func (node *BaseNode) SetMeta(meta model.MetaNode) {
	node.meta = meta
}

func (node *BaseNode) Ctx() *model.Context {
	return node.ctx
}

func (node *BaseNode) SetCtx(ctx *model.Context) {
	node.ctx = ctx
}

func (node *BaseNode) FlowOpCtx() *model.FlowOperationalContext {
	return node.flowOpCtx
}

func (node *BaseNode) SetFlowOpCtx(flowOpCtx *model.FlowOperationalContext) {
	node.flowOpCtx = flowOpCtx
}

func (node *BaseNode) SetFlowRunner(flowRunner model.FlowRunner) {
	node.flowRunner = flowRunner
}

func (node *BaseNode) SetupBaseNode() {
	node.logFields = log.Fields{"comp": "fnode", "ntype": node.meta.Type, "fid": node.flowOpCtx.FlowId, "nid": node.meta.Id}
}

func (node *BaseNode) GetLog() *log.Entry {
	return log.WithFields(node.logFields)
}

func (node *BaseNode) GetMetaNode() *model.MetaNode {
	return &node.meta
}
func (node *BaseNode) GetNextSuccessNodes() []model.NodeID {
	return []model.NodeID{node.meta.SuccessTransition}
}

func (node *BaseNode) GetNextErrorNode() model.NodeID {
	return node.meta.ErrorTransition
}

func (node *BaseNode) GetNextTimeoutNode() model.NodeID {
	return node.meta.TimeoutTransition
}

func (node *BaseNode) SetStartNode(flag bool) {
	node.isStartNode = flag
}

func (node *BaseNode) IsStartNode() bool {
	return node.isStartNode
}

func (node *BaseNode) SetMsgReactorNode(flag bool) {
	node.isMsgReactor = flag
}

func (node *BaseNode) IsMsgReactorNode() bool {
	return node.isMsgReactor
}

func (node *BaseNode) IsReactorRunning() bool {
	return node.isReactorRunning
}

// is invoked right before is started but after node was configured
func (node *BaseNode) Init() error {
	return nil
}

// is invoked when node flow is stopped
func (node *BaseNode) Cleanup() error {
	return nil
}

//func (node *BaseNode) ConfigureInStream(activeSubscriptions *[]string,msgInStream model.MsgPipeline) {
//}

func (node *BaseNode) SetConnectorRegistry(connectorRegistry *connector.Registry) {
	node.connectorRegistry = connectorRegistry
}
