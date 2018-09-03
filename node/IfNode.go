package node

import (
	"errors"
	"github.com/alivinco/fimpgo"
	"github.com/alivinco/tpflow/model"
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
	LeftVariableName    string  // Left variable of expression . If empty , Message value will be used .
	LeftVariableIsGlobal bool  // true - if left variable is global
	LeftVariable    model.Variable  `json:"-"` // Right variable of expression . Have to be defined , empty value will generate error .
	RightVariable   model.Variable // Right variable of expression . Have to be defined , empty value will generate error .
	Operand         string // eq , gr , lt
	BooleanOperator string // and , or , not
}

type IfNode struct {
	BaseNode
	ctx *model.Context
	transport *fimpgo.MqttTransport
}

func NewIfNode(flowOpCtx *model.FlowOperationalContext,meta model.MetaNode,ctx *model.Context,transport *fimpgo.MqttTransport) model.Node {
	node := IfNode{ctx:ctx,transport:transport}
	node.meta = meta
	node.flowOpCtx = flowOpCtx
	node.SetupBaseNode()
	return &node
}

func (node *IfNode) LoadNodeConfig() error {
	exp := IFExpressions{}
	err := mapstructure.Decode(node.meta.Config,&exp)
	if err != nil{
		node.getLog().Error("Failed to load node config",err)
	}else {
		node.meta.Config = exp
	}
	return nil
}

func (node *IfNode) WaitForEvent(responseChannel chan model.ReactorEvent) {

}

func (node *IfNode) OnInput( msg *model.Message) ([]model.NodeID,error) {
	var leftNumericValue , rightNumericValue float64
	var err error
	conf, ok := node.meta.Config.(IFExpressions)
	if ok {
		booleanOperator := ""
		var finalResult bool
		for i := range conf.Expression {

			if conf.Expression[i].RightVariable.ValueType == "" {
				return nil,errors.New(node.flowOpCtx.FlowId+"Right variable is not defined. IfNode is skipped.")
			}
			if conf.Expression[i].LeftVariableName == ""{
				conf.Expression[i].LeftVariable = model.Variable{ValueType:msg.Payload.ValueType,Value:msg.Payload.Value}
			}else {
				flowId := node.flowOpCtx.FlowId
				if conf.Expression[i].LeftVariableIsGlobal {
					flowId = "global"
				}
				conf.Expression[i].LeftVariable ,err = node.ctx.GetVariable(conf.Expression[i].LeftVariableName,flowId)
				if err != nil {
					node.getLog().Error("Can't get variable from context.Error : ",err)
					return nil,err
				}
			}
			if conf.Expression[i].LeftVariable.ValueType != conf.Expression[i].RightVariable.ValueType {
				return nil,errors.New(node.flowOpCtx.FlowId+"<IfNode> Right and left of expression have different types ")
			}


			var result bool
			node.getLog().Debug("Operand = ", conf.Expression[i].Operand)

			if conf.Expression[i].Operand == "gt" || conf.Expression[i].Operand == "lt"  {
				if conf.Expression[i].LeftVariable.ValueType == "int" || conf.Expression[i].LeftVariable.ValueType  == "float" {
					leftNumericValue , err = utils.ConfigValueToNumber(conf.Expression[i].LeftVariable.ValueType,conf.Expression[i].LeftVariable.Value)
					if err != nil {
						node.getLog().Error("Error while converting left variable to number.Error : ",err)
						return nil,err
					}
					rightNumericValue , err = utils.ConfigValueToNumber(conf.Expression[i].RightVariable.ValueType,conf.Expression[i].RightVariable.Value)
					if err != nil {
						node.getLog().Error("Error while converting right variable to number.Error : ",err)
						return nil,err
					}

				}else {
					return nil,errors.New("Incompatible value type . gt and lt can be used only with numeric types")
				}
			}
			//node.getLog().Debug(node.flowOpCtx.FlowId+"<IfNode> Left numeric value = ", leftNumericValue)
			//node.getLog().Debug(node.flowOpCtx.FlowId+"<IfNode> Right numeric value = ", rightNumericValue)
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
				}else {
					// first element
					finalResult = result
				}

			}else {
				finalResult = result
			}
		}
		if finalResult {
			return []model.NodeID{conf.TrueTransition},nil
		} else {
			return []model.NodeID{conf.FalseTransition},nil
		}
		return nil,nil
	} else {
		node.getLog().Error("Failed to cast IF expression")
		return nil, errors.New(node.flowOpCtx.FlowId+"Incompatible node configuration format")
	}

	return []model.NodeID{node.meta.SuccessTransition},nil
}


