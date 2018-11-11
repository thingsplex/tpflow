package rest

import (
	"github.com/alivinco/fimpgo"
	"os"
	"testing"
	"time"

	"github.com/alivinco/tpflow/model"
)

func TestNode_OnInput_Jpath(t *testing.T) {
	flowId := "RestTest"
	ctx, err := model.NewContextDB(flowId+".db")
	ctx.RegisterFlow(flowId)
	meta := model.MetaNode{Id: "2", Label: "Invoke httpbin", Type: "rest_action", SuccessTransition: "",
		Config:NodeConfig{
			Url:                  "https://httpbin.org/post",
			Method:               "POST",
			TemplateVariableName: "",
			IsVariableGlobal:     false,
			RequestPayloadType:   "",
			RequestTemplate:      "{'param1':{{.Variable}} }",
			Headers:              nil,
			HeadersVariableName:"customHeaders",
			ResponseMapping: []ResponseToVariableMap{{
				Name:                 "Host",
				Path:                 "$.headers.Host",
				PathType:             "json",
				TargetVariableName:   "host",
				IsVariableGlobal:     false,
				TargetVariableType:   "string",
				UpdateTriggerMessage: false,
			},{
				Name:                 "Connection",
				Path:                 "$.headers.Connection",
				PathType:             "json",
				TargetVariableName:   "connection",
				IsVariableGlobal:     false,
				TargetVariableType:   "string",
				UpdateTriggerMessage: false,
			}},
			LogResponse:          true,
			Auth: OAuth{
				Enabled:      false,
				GrantType:    "",
				Url:          "",
				ClientID:     "",
				ClientSecret: "",
				Scope:        "",
				Username:     "",
				Password:     "",
			},
		}}


	opCtx := model.FlowOperationalContext{
		FlowId:                   flowId,
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

	customHeaders := make(map[string]string)
	customHeaders["Object"]= "one"

	ctx.SetVariable("customHeaders","map_str",customHeaders,"",flowId,true)
	_, err = node.OnInput(&msg)

	if err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second * 2)
	variable1, err := ctx.GetVariable("connection", flowId)
	variable2, err := ctx.GetVariable("host", flowId)
	//inputMessage := flow.GetCurrentMessage()
	if variable1.Value.(string) == "close" && variable2.Value.(string) == "httpbin.org" {
		t.Log("All good.")
	}else {
		t.Error("Wrong value " )
	}

	ctx.Close()
	os.Remove(flowId+".db")

}


func TestNode_OnInput_Xpath(t *testing.T) {
	flowId := "RestTest"
	ctx, err := model.NewContextDB(flowId+".db")
	ctx.RegisterFlow(flowId)
	meta := model.MetaNode{Id: "2", Label: "Invoke httpbin", Type: "rest_action", SuccessTransition: "",
		Config:NodeConfig{
			Url:                  "https://httpbin.org/xml",
			Method:               "GET",
			TemplateVariableName: "",
			IsVariableGlobal:     false,
			RequestPayloadType:   "",
			RequestTemplate:      "",
			Headers:              nil,
			ResponseMapping: []ResponseToVariableMap{{
				Name:                 "title",
				Path:                 "/slideshow/slide[2]/title",
				PathType:             "xml",
				TargetVariableName:   "title",
				IsVariableGlobal:     false,
				TargetVariableType:   "string",
				UpdateTriggerMessage: false,
			},{
				Name:                 "author",
				Path:                 "/slideshow/@author",
				PathType:             "xml",
				TargetVariableName:   "author",
				IsVariableGlobal:     false,
				TargetVariableType:   "string",
				UpdateTriggerMessage: false,
			}},
			LogResponse:          true,
			Auth: OAuth{
				Enabled:      false,
				GrantType:    "",
				Url:          "",
				ClientID:     "",
				ClientSecret: "",
				Scope:        "",
				Username:     "",
				Password:     "",
			},
		}}


	opCtx := model.FlowOperationalContext{
		FlowId:                   flowId,
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
	time.Sleep(time.Second * 2)
	variable1, err := ctx.GetVariable("title", flowId)
	variable2, err := ctx.GetVariable("author", flowId)
	if variable1.Value.(string) == "Overview" && variable2.Value.(string) == "Yours Truly" {
		t.Log("All good.")
	}else {
		t.Error("Wrong value " )
	}

	ctx.Close()
	os.Remove(flowId+".db")

}
