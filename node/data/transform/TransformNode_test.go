package transform

import (
	"encoding/json"
	"github.com/alivinco/fimpgo"
	"os"
	"testing"
	"time"

	"github.com/alivinco/tpflow/model"
)

func TestNode_OnInput_Calc(t *testing.T) {

	ctx, err := model.NewContextDB("TestTransform.db")
	meta  := model.MetaNode{
		Id:                "2",
		Type:              "transform",
		Label:             "Add transform",
		SuccessTransition: "",
		TimeoutTransition: "",
		ErrorTransition:   "",
		Address:           "",
		Service:           "",
		ServiceInterface:  "",
		Config:            NodeConfig{
			TargetVariableName:     "",
			TargetVariableType:     "",
			IsTargetVariableGlobal: false,
			TransformType:          "calc",
			IsRVariableGlobal:      false,
			IsLVariableGlobal:      false,
			Operation:              "add",
			RType:                  "",
			RValue:                 model.Variable{ValueType: "int", Value: int(2)},
			RVariableName:          "",
			LVariableName:          "",
			ValueMapping:           nil,
			XPathMapping:           nil,
			Template:               "",
		},
		Ui:                nil,
	}

	opCtx := model.FlowOperationalContext{
		FlowId:                   "test",
		IsFlowRunning:            true,
		State:                    "",
		NodeControlSignalChannel: nil,
		NodeIsReady:              nil,
		StoragePath:              "",
	}

	node := NewNode(&opCtx,meta,ctx)
	node.LoadNodeConfig()

	msg := model.Message{Payload:fimpgo.FimpMessage{
		Type:          "",
		Service:       "",
		ValueType:     "int",
		Value:         5,
		ValueObj:      nil,
		Tags:          nil,
		Properties:    nil,
		Version:       "",
		CorrelationID: "",
		CreationTime:  time.Time{},
		UID:           "",
	}}
	_, err = node.OnInput(&msg)

	if err != nil {
		t.Error(err)
	}


	if msg.Payload.Value.(int64) != 7 {
		t.Error("Wrong result = ",msg.Payload.Value.(int))
	}

	ctx.Close()
	os.Remove("TestTransform.db")

}


func TestNode_OnInput_Jpath(t *testing.T) {

	ctx, err := model.NewContextDB("TestTransform.db")
	ctx.RegisterFlow("test")
	meta  := model.MetaNode{
		Id:                "2",
		Type:              "transform",
		Label:             "Add transform",
		SuccessTransition: "",
		TimeoutTransition: "",
		ErrorTransition:   "",
		Address:           "",
		Service:           "",
		ServiceInterface:  "",
		Config:            NodeConfig{
			TargetVariableName:     "",
			TargetVariableType:     "",
			IsTargetVariableGlobal: false,
			TransformType:          "jpath",
			IsRVariableGlobal:      false,
			IsLVariableGlobal:      false,
			Operation:              "",
			RType:                  "",
			RVariableName:          "",
			LVariableName:          "temp_report",
			ValueMapping:           nil,
			XPathMapping: []TransformXPathRecord{{
				Name:                   "",
				Path:                   "$.temp",
				TargetVariableName:     "temp_value",
				TargetVariableType:     "float",
				IsTargetVariableGlobal: false,
				UpdateInputVariable:    false,
			}},
			Template:               "",
		},
		Ui:                nil,
	}

	opCtx := model.FlowOperationalContext{
		FlowId:                   "test",
		IsFlowRunning:            true,
		State:                    "",
		NodeControlSignalChannel: nil,
		NodeIsReady:              nil,
		StoragePath:              "",
	}

	node := NewNode(&opCtx,meta,ctx)
	node.LoadNodeConfig()

	msg := model.Message{Payload:fimpgo.FimpMessage{
		Type:          "",
		Service:       "",
		ValueType:     "int",
		Value:         5,
		ValueObj:      nil,
		Tags:          nil,
		Properties:    nil,
		Version:       "",
		CorrelationID: "",
		CreationTime:  time.Time{},
		UID:           "",
	}}

	var jvar interface{}
	if err := json.Unmarshal([]byte("{\"temp\":23.5}"), &jvar); err != nil {
		t.Error(err)
	}

	ctx.SetVariable("temp_report","object",jvar,"","test",true)
	_, err = node.OnInput(&msg)

	if err != nil {
		t.Error(err)
	}

	tempSensorValue,err := ctx.GetVariable("temp_value","test")

	temp := tempSensorValue.Value.(float64)

	if temp != 23.5 {
		t.Error("Wrong result = ",msg.Payload.Value.(int))
	}

	t.Log("Temperature = ",temp)

	ctx.Close()
	os.Remove("TestTransform.db")

}

