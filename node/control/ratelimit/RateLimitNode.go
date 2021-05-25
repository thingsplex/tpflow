package ratelimit

import (
	"context"
	"github.com/mitchellh/mapstructure"
	context2 "github.com/thingsplex/tpflow/flow/context"
	"github.com/thingsplex/tpflow/model"
	"github.com/thingsplex/tpflow/node/base"
    "golang.org/x/time/rate"
	"time"
)

type NodeConfig struct {
	TimeInterval    int64   // Time interval in seconds
	Limit           int     // Limit is expressed in events / time interval
	Action          string  // skip,wait
}

// IF node
type Node struct {
	base.BaseNode
	limiter *rate.Limiter
	config NodeConfig
	ctx    *context2.Context
}

func NewNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *context2.Context) model.Node {
	node := Node{ctx: ctx}
	node.SetMeta(meta)
	node.SetFlowOpCtx(flowOpCtx)
	node.SetupBaseNode()
	return &node
}

func (node *Node) LoadNodeConfig() error {
	exp := NodeConfig{}
	err := mapstructure.Decode(node.Meta().Config, &exp)
	if err != nil {
		node.GetLog().Error("Failed to load node config", err)
	} else {
		node.config = exp
	}
	limit := rate.Every(time.Second*time.Duration(node.config.TimeInterval))
	node.limiter = rate.NewLimiter(limit,node.config.Limit)
	return nil
}

func (node *Node) WaitForEvent(responseChannel chan model.ReactorEvent) {

}

func (node *Node) OnInput(msg *model.Message) ([]model.NodeID, error) {
	if node.config.Action == "skip" {
		if node.limiter.Allow() {
			return []model.NodeID{node.Meta().SuccessTransition}, nil
		}
	}else if node.config.Action == "wait" {
		node.GetLog().Debug("Waiting")
		ctx := context.Background()
		node.limiter.Wait(ctx)
		return []model.NodeID{node.Meta().SuccessTransition}, nil
	}
	node.GetLog().Debug("Request skipped by rate limiter.")
	return []model.NodeID{node.Meta().ErrorTransition}, nil
}
