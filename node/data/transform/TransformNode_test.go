package transform

import (
	"encoding/json"
	"github.com/futurehomeno/fimpgo"
	"github.com/thingsplex/tpflow/model"
	"os"
	"testing"
)

func TestNode_OnInput_Calc(t *testing.T) {

	ctx, err := model.NewContextDB("TestTransform.db")
	ctx.SetVariable("temp","float",20.0,"","test",true)
	ctx.SetVariable("var2","float",10.0,"","test",true)
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
			TargetVariableName:     "result_var",
			TargetVariableType:     "bool",
			IsTargetVariableGlobal: false,
			IsTargetVariableInMemory:true,
			TransformType:          "calc",
			IsRVariableGlobal:      false,
			IsLVariableGlobal:      false,
			Expression:             "(temp+var2+input)==35",
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
		FlowId:                      "test",
		IsFlowRunning:               true,
		State:                       "",
		TriggerControlSignalChannel: nil,
		NodeIsReady:                 nil,
		StoragePath:                 "",
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
		UID:           "",
	}}
	_, err = node.OnInput(&msg)

	if err != nil {
		t.Error(err)
	}


	//if msg.Payload.Value.(int64) != 7 {
	//	t.Error("Wrong result = ",msg.Payload.Value.(int))
	//}
	rvar,err := ctx.GetVariable("result_var","test")
	if !rvar.Value.(bool) {
		t.Error("Wrong result = ",rvar.Value)
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
			Expression:              "",
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
			},{
				Name:                   "",
				Path:                   "$.roomId",
				TargetVariableName:     "roomId",
				TargetVariableType:     "int",
				IsTargetVariableGlobal: false,
				UpdateInputVariable:    false,
			},{
				Name:                   "",
				Path:                   "$.devices[?(@.val==9.0)]",
				TargetVariableName:     "devTemp",
				TargetVariableType:     "object",
				IsTargetVariableGlobal: false,
				UpdateInputVariable:    false,
			}},
			Template:               "",
		},
		Ui:                nil,
	}

	opCtx := model.FlowOperationalContext{
		FlowId:                      "test",
		IsFlowRunning:               true,
		State:                       "",
		TriggerControlSignalChannel: nil,
		NodeIsReady:                 nil,
		StoragePath:                 "",
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
		UID:           "",
	}}

	var jvar interface{}
	if err := json.Unmarshal([]byte("{\"temp\":23.5,\"roomId\":201701142110080000,\"devices\":[{\"name\":\"temp\",\"val\":5},{\"name\":\"humid\",\"val\":9}]}"), &jvar); err != nil {
		t.Error(err)
	}

	ctx.SetVariable("temp_report","object",jvar,"","test",true)
	_, err = node.OnInput(&msg)

	if err != nil {
		t.Error(err)
	}

	tempSensorValue,err := ctx.GetVariable("temp_value","test")

	temp := tempSensorValue.Value.(float64)

	roomIdValue,err := ctx.GetVariable("roomId","test")

	roomId := roomIdValue.Value.(int)

	devTempVal,err := ctx.GetVariable("devTemp","test")

	//devTempVal.ToNumber()
	devTemp,_ := devTempVal.Value.(int)



	if temp != 23.5 {
		t.Error("Wrong result = ",msg.Payload.Value.(int))
	}

	if roomId != 201701142110080000 {
		t.Error("Wrong result = ",msg.Payload.Value.(int))
	}

	if devTemp != 5 {
		t.Error("Wrong devTemp result = ",devTempVal)
	}


	t.Log("Temperature = ",temp)
	t.Logf("Room id  = %d",roomId)

	ctx.Close()
	os.Remove("TestTransform.db")

}

