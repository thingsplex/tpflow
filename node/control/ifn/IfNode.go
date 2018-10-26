package ifn

import (
	"errors"
	"github.com/alivinco/tpflow/model"
	"github.com/alivinco/tpflow/node/base"
	"github.com/alivinco/tpflow/utils"
	//"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/mapstructure"
)

type IFExpressions struct {
	Expression      []IFExpression
	TrueTransition  model.NodeID
	FalseTransition model.NodeID
}

type IFExpression struct {
	LeftVariableName     string         // Left variable of expression . If empty , Message value will be used .
	LeftVariableIsGlobal bool           // true - if left variable is global
	LeftVariable         model.Variable `json:"-"` // Right variable of expression . Have to be defined , empty value will generate error .
	RightVariable        model.Variable // Right variable of expression . Have to be defined , empty value will generate error .
	Operand              string         // eq , gr , lt
	BooleanOperator      string         // and , or , not
}

// IF node
type Node struct {
	base.BaseNode
	config IFExpressions
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
	exp := IFExpressions{}
	err := mapstructure.Decode(node.Meta().Config, &exp)
	if err != nil {
		node.GetLog().Error("Failed to load node config", err)
	} else {
		node.config = exp
	}
	return nil
}

func (node *Node) WaitForEvent(responseChannel chan model.ReactorEvent) {

}

func (node *Node) OnInput(msg *model.Message) ([]model.NodeID, error) {
	var leftNumericValue, rightNumericValue float64
	var err error
	conf := node.config
	booleanOperator := ""
	var finalResult bool
	for i := range conf.Expression {

		if conf.Expression[i].RightVariable.ValueType == "" {
			return nil, errors.New(node.FlowOpCtx().FlowId + "Right variable is not defined. Node is skipped.")
		}
		if conf.Expression[i].LeftVariableName == "" {
			conf.Expression[i].LeftVariable = model.Variable{ValueType: msg.Payload.ValueType, Value: msg.Payload.Value}
		} else {
			flowId := node.FlowOpCtx().FlowId
			if conf.Expression[i].LeftVariableIsGlobal {
				flowId = "global"
			}
			conf.Expression[i].LeftVariable, err = node.ctx.GetVariable(conf.Expression[i].LeftVariableName, flowId)
			if err != nil {
				node.GetLog().Error("Can't get variable from context.Error : ", err)
				return nil, err
			}
		}
		if conf.Expression[i].LeftVariable.ValueType != conf.Expression[i].RightVariable.ValueType {
			return nil, errors.New(node.FlowOpCtx().FlowId + "<Node> Right and left of expression have different types ")
		}

		var result bool
		node.GetLog().Debug("Operand = ", conf.Expression[i].Operand)

		if conf.Expression[i].Operand == "gt" || conf.Expression[i].Operand == "lt" {
			if conf.Expression[i].LeftVariable.ValueType == "int" || conf.Expression[i].LeftVariable.ValueType == "float" {
				leftNumericValue, err = utils.ConfigValueToNumber(conf.Expression[i].LeftVariable.ValueType, conf.Expression[i].LeftVariable.Value)
				if err != nil {
					node.GetLog().Error("Error while converting left variable to number.Error : ", err)
					return nil, err
				}
				rightNumericValue, err = utils.ConfigValueToNumber(conf.Expression[i].RightVariable.ValueType, conf.Expression[i].RightVariable.Value)
				if err != nil {
					node.GetLog().Error("Error while converting right variable to number.Error : ", err)
					return nil, err
				}

			} else {
				return nil, errors.New("Incompatible value type . gt and lt can be used only with numeric types")
			}
		}
		//node.GetLog().Debug(node.flowOpCtx.FlowId+"<Node> Left numeric value = ", leftNumericValue)
		//node.GetLog().Debug(node.flowOpCtx.FlowId+"<Node> Right numeric value = ", rightNumericValue)
		switch conf.Expression[i].Operand {
		case "eq":
			result = conf.Expression[i].LeftVariable.Value == conf.Expression[i].RightVariable.Value
		case "gt":
			result = leftNumericValue > rightNumericValue
		case "lt":
			result = leftNumericValue < rightNumericValue
		}
		if len(conf.Expression) > 1 {
			if i > 0 {
				// boolean operator between current and previous element
				booleanOperator = conf.Expression[i-1].BooleanOperator
				switch booleanOperator {
				case "":
					// empty = and
					finalResult = finalResult && result
				case "and":
					finalResult = finalResult && result
				case "or":
					finalResult = finalResult || result
				case "not":
					finalResult = !result
				}
			} else {
				// first element
				finalResult = result
			}

		} else {
			finalResult = result
		}
	}
	if finalResult {
		return []model.NodeID{conf.TrueTransition}, nil
	} else {
		return []model.NodeID{conf.FalseTransition}, nil
	}
	return nil, nil

	return []model.NodeID{node.Meta().SuccessTransition}, nil
}
