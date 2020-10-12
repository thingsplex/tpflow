package node

import (
	"github.com/thingsplex/tpflow/model"
	"github.com/thingsplex/tpflow/node/action/exec"
	actfimp "github.com/thingsplex/tpflow/node/action/fimp"
	log "github.com/thingsplex/tpflow/node/action/log"
	"github.com/thingsplex/tpflow/node/action/rest"
	"github.com/thingsplex/tpflow/node/control/ifn"
	"github.com/thingsplex/tpflow/node/control/iftime"
	"github.com/thingsplex/tpflow/node/control/loop"
	"github.com/thingsplex/tpflow/node/control/ratelimit"
	"github.com/thingsplex/tpflow/node/control/wait"
	"github.com/thingsplex/tpflow/node/data/setvar"
	"github.com/thingsplex/tpflow/node/data/transform"
	trigfimp "github.com/thingsplex/tpflow/node/trigger/fimp"
	"github.com/thingsplex/tpflow/node/trigger/time"
)

type Constructor func(context *model.FlowOperationalContext, meta model.MetaNode, ctx *model.Context) model.Node

var Registry = map[string]Constructor{
	"trigger":      trigfimp.NewTriggerNode,
	"vinc_trigger": trigfimp.NewVincTriggerNode,
	"receive":      trigfimp.NewReceiveNode,
	"if":           ifn.NewNode,
	"iftime":		iftime.NewNode,
	"rate_limit":	ratelimit.NewNode,
	"action":       actfimp.NewNode,
	"log_action":   log.NewNode,
	"rest_action":  rest.NewNode,
	"wait":         wait.NewWaitNode,
	"set_variable": setvar.NewSetVariableNode,
	"loop":         loop.NewNode,
	"time_trigger": time.NewNode,
	"transform":    transform.NewNode,
	"exec":         exec.NewNode,
}
