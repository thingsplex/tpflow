package node

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/alivinco/fimpgo"
	"github.com/alivinco/tpflow/model"
	"github.com/mitchellh/mapstructure"
	"text/template"
)

type TransformNode struct {
	BaseNode
	ctx *model.Context
	nodeConfig TransformNodeConfig
	transport *fimpgo.MqttTransport
	template *template.Template
}

type TransformNodeConfig struct {
	TargetVariableName string  // Variable
	TargetVariableType string
	IsTargetVariableGlobal bool
	TransformType string       // map , calc , str-to-json ,json-to-str , jpath , xpath , template
	IsRVariableGlobal bool                    // true - update global variable ; false - update local variable
	IsLVariableGlobal bool                    // true - update global variable ; false - update local variable
	Operation string 			// type of transform operation , flip , add , subtract , multiply , divide , to_bool
	RType     string            // var , const
	RValue model.Variable 		// Constant Right variable value .
	RVariableName string 		// Right variable name , if empty , RValue will be used instead
 	LVariableName string  		// Update input message if LVariable is empty
 	ValueMapping []ValueMappingRecord // ["LValue":1,"RValue":"mode-1"]
 	XPathMapping []TransformXPathRecord
 	Template string // template used in jtemplate transformation

 	//value mapping
}

type ValueMappingRecord struct {
	LValue model.Variable
	RValue model.Variable
}

type TransformXPathRecord struct {
	Name string
	Path string
	TargetVariableName string
	TargetVariableType string
	IsTargetVariableGlobal bool
	UpdateInputVariable bool
}

func NewTransformNode(flowOpCtx *model.FlowOperationalContext,meta model.MetaNode,ctx *model.Context,transport *fimpgo.MqttTransport) model.Node {
	node := TransformNode{ctx:ctx,transport:transport}
	node.meta = meta
	node.flowOpCtx = flowOpCtx
	node.SetupBaseNode()
	return &node
}

func (node *TransformNode) LoadNodeConfig() error {
	defValue := TransformNodeConfig{}
	err := mapstructure.Decode(node.meta.Config,&defValue)
	if err != nil{
		node.getLog().Error(" Can't decode configuration",err)
		return err
	}else {
		node.nodeConfig = defValue
		node.meta.Config = defValue
	}
	if node.nodeConfig.TransformType == "template" {
		funcMap := template.FuncMap{
			"variable": func(varName string,isGlobal bool)(interface{},error) {
				//node.getLog().Debug("Getting variable by name ",varName)
				var vari model.Variable
				var err error
				if isGlobal {
					vari , err = node.ctx.GetVariable(varName,"global")
				}else {
					vari , err = node.ctx.GetVariable(varName,node.flowOpCtx.FlowId)
				}

				if vari.IsNumber() {
					return vari.ToNumber()
				}
				vstr , ok := vari.Value.(string)
				if ok {
					return vstr,err
				}else {
					return "",errors.New("Only simple types are supported ")
				}

			},
		}
		node.template,err = template.New("transform").Funcs(funcMap).Parse(node.nodeConfig.Template)
		if err != nil {
			node.getLog().Error(" Failed while parsing request template.Error:",err)
			return err
		}
	}

	return nil
}

func (node *TransformNode) OnInput( msg *model.Message) ([]model.NodeID,error) {
	node.getLog().Info(" Executing TransformNode . Name = ", node.meta.Label)
	node.getLog().Info(" Transform type  = ", node.nodeConfig.TransformType)

	// There are 3 possible sources for RVariable : default value , inputMessage , variable from context
	// There are 2 possible destinations for LVariable : inputMessage , variable from context
	var lValue model.Variable
	var rValue model.Variable
	var result model.Variable
	var err error

	if node.nodeConfig.LVariableName == "" {
		// Use input message
		lValue.Value = msg.Payload.Value
		lValue.ValueType = msg.Payload.ValueType
	} else {
		// Use variable
		if node.nodeConfig.IsLVariableGlobal {
			lValue,err = node.ctx.GetVariable(node.nodeConfig.LVariableName,"global")
		}else {
			lValue,err = node.ctx.GetVariable(node.nodeConfig.LVariableName,node.flowOpCtx.FlowId)
		}
	}

	if err != nil {
		node.getLog().Warn(" Error 1 : ",err)
		return []model.NodeID{node.meta.ErrorTransition} , err
	}

    if node.nodeConfig.RType == "var" {
		if node.nodeConfig.RVariableName == "" {
			rValue.Value = msg.Payload.Value
			rValue.ValueType = msg.Payload.ValueType
		}else{
			// Use variable
			if node.nodeConfig.IsRVariableGlobal {
				rValue,err = node.ctx.GetVariable(node.nodeConfig.RVariableName,"global")
			}else {
				rValue,err = node.ctx.GetVariable(node.nodeConfig.RVariableName,node.flowOpCtx.FlowId)
			}
		}
	}else {
		rValue = node.nodeConfig.RValue
	}


	if err != nil {
		node.getLog().Warn(" Error 2 : ",err)
		return []model.NodeID{node.meta.ErrorTransition} , err
	}

    if lValue.ValueType == rValue.ValueType || (lValue.IsNumber() && rValue.IsNumber()) ||
		(node.nodeConfig.TransformType == "xpath" || node.nodeConfig.TransformType == "jpath"||
			node.nodeConfig.TransformType == "template" || node.nodeConfig.TransformType == "map" )  {

    	if node.nodeConfig.TransformType == "calc" {
			switch node.nodeConfig.Operation {
			case "flip":
				if lValue.ValueType == "bool" {
					val,ok := rValue.Value.(bool)
					if ok {
						result.Value = !val
						result.ValueType = rValue.ValueType
					}else {
						node.getLog().Error(" Value type is not bool. Has to bool")
						return []model.NodeID{node.meta.ErrorTransition},err
					}
				}else {
					node.getLog().Warn(" Only bool variable can be flipped")
					return []model.NodeID{node.meta.ErrorTransition},err
				}
			case "to_bool":
				if lValue.IsNumber() {
					val,err := lValue.ToNumber()
					if err == nil {
						if val == 0 {
							result.Value = false
						} else {
							result.Value = true
						}
						result.ValueType = "bool"
					}else {
						node.getLog().Error(" Value type is not number.")
					}
				}else {
					node.getLog().Warn(" Only numeric value can be converted into bool")
					return []model.NodeID{node.meta.ErrorTransition},err
				}
			case "add","subtract","multiply","divide":
				if lValue.IsNumber(){
					rval,err := rValue.ToNumber()
					lval,err := lValue.ToNumber()
					var calcResult float64
					if err == nil {
						switch node.nodeConfig.Operation {
						case "add":
							calcResult = lval + rval
						case "subtract":
							calcResult = lval - rval
						case "multiply":
							calcResult = lval * rval
						case "divide":
							calcResult = lval / rval
						default:
							node.getLog().Warn(" Unknown arithmetic operator")
						}
						if rValue.ValueType == "float" {
							result.Value = calcResult
						}else {
							result.Value = int64(calcResult)
						}
						result.ValueType = lValue.ValueType

					}else {
						node.getLog().Error(" Value type is not number.")
						return []model.NodeID{node.meta.ErrorTransition},err
					}
				}else {
					node.getLog().Warn(" Only numeric value can be used for arithmetic operations")
					return []model.NodeID{node.meta.ErrorTransition},err
				}

			}
		}else if node.nodeConfig.TransformType == "map" {
			for i := range node.nodeConfig.ValueMapping {
				//node.getLog().Debug(" record Value ",node.nodeConfig.ValueMapping[i].LValue.Value)
				//node.getLog().Debug(" record input Value = ",lValue.Value )
				if lValue.ValueType == node.nodeConfig.ValueMapping[i].LValue.ValueType {
					varsAreEqual , err :=  lValue.IsEqual(&node.nodeConfig.ValueMapping[i].LValue)
					if err != nil {
						node.getLog().Warn(" Error while comparing map vars : ",err)
						return []model.NodeID{node.meta.ErrorTransition},err
					}
					if varsAreEqual {
						result = node.nodeConfig.ValueMapping[i].RValue
						break
					}
				}
			}
		}else if node.nodeConfig.TransformType == "jpath" || node.nodeConfig.TransformType == "xpath" {
			node.getLog().Info(" Doing XPATH transformation ")
			for i := range node.nodeConfig.XPathMapping {
				result.Value,err = model.GetValueByPath(msg,node.nodeConfig.TransformType,node.nodeConfig.XPathMapping[i].Path,node.nodeConfig.XPathMapping[i].TargetVariableType)
				result.ValueType = node.nodeConfig.XPathMapping[i].TargetVariableType
				node.getLog().Info(" Extracted value : ",result.Value)
				if err != nil {
					node.getLog().Warn(" Error while processing path in variable : ",err)
					return []model.NodeID{node.meta.ErrorTransition},err
				}
				if node.nodeConfig.XPathMapping[i].TargetVariableName == "" {
					// Update input message
					msg.Payload.Value = result.Value
					msg.Payload.ValueType = result.ValueType
				}else {
					// Save value into variable
					// Save default value from node config to variable
					node.getLog().Info(" Setting transformed variable : ")
					if node.nodeConfig.XPathMapping[i].IsTargetVariableGlobal {
						node.ctx.SetVariable(node.nodeConfig.XPathMapping[i].TargetVariableName, result.ValueType, result.Value, "", "global", false)
					} else {
						node.ctx.SetVariable(node.nodeConfig.XPathMapping[i].TargetVariableName, result.ValueType, result.Value, "", node.flowOpCtx.FlowId, false)
					}

				}
			}
			return []model.NodeID{node.meta.SuccessTransition},nil
		}else if node.nodeConfig.TransformType == "template" {
			node.getLog().Info(" Doing template transformation ")
			var templateBuffer bytes.Buffer
			var template = struct {
				Variable interface{}
			}{Variable:lValue.Value}
			node.template.Execute(&templateBuffer,template)

			err = json.Unmarshal(templateBuffer.Bytes(),&result.Value)
			node.getLog().Debug("Template output:",templateBuffer.String())
			if err != nil {
				node.getLog().Warn("Error Unmarshaling template output",err)
				return []model.NodeID{node.meta.ErrorTransition},err
			}

			result.ValueType = node.nodeConfig.TargetVariableType
		}
	}else {
		node.getLog().Warn("Transformation can't be executed , one or several parameters are wrong. ")
		return []model.NodeID{node.meta.ErrorTransition},err
	}

	if node.nodeConfig.TargetVariableName == "" {
		node.getLog().Debug("Updating input variable.")
		// Update input message
		msg.Payload.Value = result.Value
		msg.Payload.ValueType = result.ValueType
	}else {
		// Save value into variable
		// Save default value from node config to variable
		if node.nodeConfig.IsTargetVariableGlobal {
			    node.ctx.SetVariable(node.nodeConfig.TargetVariableName, result.ValueType, result.Value, "", "global", false)
		} else {
				node.ctx.SetVariable(node.nodeConfig.TargetVariableName, result.ValueType, result.Value, "", node.flowOpCtx.FlowId, false)
		}

	}
	return []model.NodeID{node.meta.SuccessTransition},nil
}


func (node *TransformNode) WaitForEvent(responseChannel chan model.ReactorEvent) {

}
