package iftime

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/thingsplex/tpflow/model"
	"github.com/thingsplex/tpflow/node/base"
	"strconv"
	"strings"
	"time"
)

type TExpressions struct {
	Expression      []TExpression
	TrueTransition  model.NodeID
	FalseTransition model.NodeID
}

// [{"wday":"1","from":"12:00", "to":"13:00" ,"action":"allow"},
//  {"wday":"*","from":"12:00", "to":"13:00" ,"action":"allow"}]
// "from" - can be time or "sunrise" and "sunset"

type TExpression struct {
	Weekday  string   // 0 - Sunday , 1 - Monday
	From     string  // time in format 12:00 or "sunrise" or "sunset"
	To       string  // time in format 12:00 or "sunrise" or "sunset"
	Action   string  // a - allow or d - deny
}

// IF node
type Node struct {
	base.BaseNode
	config TExpressions
	ctx    *model.Context
}

func NewNode(flowOpCtx *model.FlowOperationalContext, meta model.MetaNode, ctx *model.Context) model.Node {
	node := Node{ctx: ctx}
	node.SetMeta(meta)
	node.SetFlowOpCtx(flowOpCtx)
	node.SetupBaseNode()
	return &node
}

func (node *Node) LoadNodeConfig() error {
	exp := TExpressions{}
	err := mapstructure.Decode(node.Meta().Config, &exp)
	if err != nil {
		node.GetLog().Error("Failed to load node config", err)
	} else {
		node.config = exp
	}
	return nil
}

func parseTime(timeStr string)(minutes int,err error) {
	var hours int
	fComp := strings.Split(timeStr,":")
	if len(fComp)==2 {
		var errh,errm error
		hours,errh = strconv.Atoi(fComp[0])
		minutes,errm = strconv.Atoi(fComp[1])
		if errh != nil || errm != nil {
			err = fmt.Errorf("wrong format")
		}
		minutes += hours * 60
		return
	}
	return 0,fmt.Errorf("wrong format")
}


func(node *Node) isNowInRange(from,to string) bool {
	fromMinutesSinceMidnight,err1 := parseTime(from)
	toMinutesSinceMidnight,err2 := parseTime(to)
	if err1 != nil || err2 != nil  {
		node.GetLog().Error("Time range parse error . Err:",err1)
		return false
	}
	var nowMinutesSinceMidnight int
	now := time.Now()
	nowMinutesSinceMidnight = now.Hour() * 60
	nowMinutesSinceMidnight+= now.Minute()
	if fromMinutesSinceMidnight < nowMinutesSinceMidnight && nowMinutesSinceMidnight < toMinutesSinceMidnight {
		//node.GetLog().Debug("Time is in range")
		return true
	}
	//node.GetLog().Debug("Time is NOT in range")
	return false
}

func (node *Node) WaitForEvent(responseChannel chan model.ReactorEvent) {

}

func (node *Node) OnInput(msg *model.Message) ([]model.NodeID, error) {
	conf := node.config
	now := time.Now()
	for _,exp := range conf.Expression {
		wday,err := strconv.Atoi(exp.Weekday)
		if err != nil {
			node.GetLog().Errorf("Incorrect Weekday format:%s",exp.Weekday)
			continue
		}
		if now.Weekday() == time.Weekday(wday) && node.isNowInRange(exp.From,exp.To) {
			if exp.Action == "a" {
				return []model.NodeID{node.Meta().SuccessTransition}, nil
			}else {
				return []model.NodeID{node.Meta().ErrorTransition}, nil
			}

		}
	}
	return []model.NodeID{node.Meta().ErrorTransition}, nil
}
