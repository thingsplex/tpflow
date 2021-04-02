package exec

import (
	"github.com/futurehomeno/fimpgo"
	"os"
	"testing"
	"time"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/tpflow/model"
)

func TestExecNode_OnInput_Python(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	flowId := "ExecTest"
	ctx, err := model.NewContextDB(flowId+".db")
	ctx.RegisterFlow(flowId)
	meta := model.MetaNode{
		Id:                "1",
		Type:              "exec_action",
		Label:             "Test python",
		SuccessTransition: "2",
		TimeoutTransition: "3",
		ErrorTransition:   "",
		Address:           "",
		Service:           "",
		ServiceInterface:  "",
		Config:NodeConfig{
			ExecType:               "python",
			Command:                "",
			ScriptBody:             "import fimp; print('99')",
			InputVariableName:      "",
			IsInputVariableGlobal:  false,
			OutputVariableName:     "",
			OutputVariableType:     "int",
			IsOutputVariableGlobal: false,
			IsOutputJson:           false,
			IsInputJson:            false,
		},
		Ui:                nil,
	}


	opCtx := model.FlowOperationalContext{
		FlowId:                      flowId,
		IsFlowRunning:               true,
		State:                       "",
		TriggerControlSignalChannel: nil,
		NodeIsReady:                 nil,
		StoragePath:                 "",
		ExtLibsDir:                  "../../../extlibs",
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


	nodes, err := node.OnInput(&msg)

	if err != nil {
		t.Error(err)
	}

	if nodes[0] == "3" {
		t.Error("Error transition")
	}

	time.Sleep(time.Second * 2)

	ctx.Close()
	os.Remove(flowId+".db")
	os.Remove(flowId+"_1.py")

}


