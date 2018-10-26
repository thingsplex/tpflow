package node

import (
	"github.com/alivinco/tpflow/model"
	"github.com/alivinco/tpflow/node/action/exec"
	actfimp "github.com/alivinco/tpflow/node/action/fimp"
	"github.com/alivinco/tpflow/node/action/rest"
	"github.com/alivinco/tpflow/node/control/ifn"
	"github.com/alivinco/tpflow/node/control/loop"
	"github.com/alivinco/tpflow/node/control/wait"
	"github.com/alivinco/tpflow/node/data/setvar"
	"github.com/alivinco/tpflow/node/data/transform"
	trigfimp "github.com/alivinco/tpflow/node/trigger/fimp"
	"github.com/alivinco/tpflow/node/trigger/time"
)

type Constructor func(context *model.FlowOperationalContext, meta model.MetaNode, ctx *model.Context) model.Node

var Registry = map[string]Constructor{
	"trigger":      trigfimp.NewTriggerNode,
	"receive":      trigfimp.NewReceiveNode,
	"if":           ifn.NewNode,
	"action":       actfimp.NewNode,
	"rest_action":  rest.NewNode,
	"wait":         wait.NewWaitNode,
	"set_variable": setvar.NewSetVariableNode,
	"loop":         loop.NewNode,
	"time_trigger": time.NewNode,
	"transform":    transform.NewNode,
	"exec":         exec.NewNode,
}
