package flow

import (
	log "github.com/Sirupsen/logrus"
	"github.com/alivinco/fimpgo"
	"github.com/alivinco/tpflow/connector"
	"github.com/alivinco/tpflow/model"
	flownode "github.com/alivinco/tpflow/node"
	"os"
	"testing"
	"time"
)

var msgChan = make(model.MsgPipeline,10)

func onMsg(topic string, addr *fimpgo.Address, iotMsg *fimpgo.FimpMessage, rawMessage []byte) {
	log.Info("New message from topic = ", topic)

	fMsg := model.Message{AddressStr: topic, Address: *addr, Payload: *iotMsg}
	select {
	case msgChan <- fMsg:
		log.Info("<Test> Message was sent")
	default:
		log.Info("<Test> Message dropped , no receiver ")
	}
}

func sendMsg(mqtt *fimpgo.MqttTransport) {
	msg := fimpgo.NewBoolMessage("evt.binary.report", "out_bin_switch", true, nil, nil, nil)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "test", ResourceAddress: "1", ServiceName: "out_bin_switch", ServiceAddress: "199_0"}
	mqtt.Publish(&adr, msg)
}

func TestWaitFlow(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	mqtt := fimpgo.NewMqttTransport("tcp://localhost:1883", "flow_test", "", "", true, 1, 1)
	err := mqtt.Start()
	t.Log("Connected")
	if err != nil {
		t.Error("Error connecting to broker ", err)
	}

	mqtt.SetMessageHandler(onMsg)
	time.Sleep(time.Second * 1)

	ctx, err := model.NewContextDB("TestWaitFlow.db")

	flowMeta := model.FlowMeta{}
	node := model.MetaNode{Id: "1", Label: "Lux trigger", Type: "trigger", Address: "pt:j1/mt:evt/rt:dev/rn:test/ad:1/sv:out_bin_switch/ad:199_0", Service: "out_bin_switch", ServiceInterface: "evt.binary.report", SuccessTransition: "2"}
	flowMeta.Nodes = append(flowMeta.Nodes, node)
	// Node
	node = model.MetaNode{Id: "2", Label: "Bulb 1", Type: "action", Address: "pt:j1/mt:cmd/rt:dev/rn:test/ad:1/sv:out_bin_switch/ad:200_0", Service: "out_bin_switch", ServiceInterface: "cmd.binary.set", SuccessTransition: "2.1"}
	flowMeta.Nodes = append(flowMeta.Nodes, node)
	// Node
	node = model.MetaNode{Id: "2.1", Label: "Waiting for 500mil", Type: "wait", SuccessTransition: "3", Config: float64(200)}
	flowMeta.Nodes = append(flowMeta.Nodes, node)
	// Node
	node = model.MetaNode{Id: "3", Label: "Bulb 2", Type: "action", Address: "pt:j1/mt:cmd/rt:dev/rn:test/ad:1/sv:out_bin_switch/ad:200_0", Service: "out_bin_switch", ServiceInterface: "cmd.binary.set", SuccessTransition: ""}
	flowMeta.Nodes = append(flowMeta.Nodes, node)
	flow := NewFlow(flowMeta, ctx)
	flow.LoadAndConfigureAllNodes()
	flow.Start()
	time.Sleep(time.Second * 1)
	sendMsg(mqtt)
	time.Sleep(time.Second * 5)
	os.Remove("TestWaitFlow.db")

}

func TestIfFlow(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	mqtt := fimpgo.NewMqttTransport("tcp://localhost:1883", "flow_test", "", "", true, 1, 1)
	err := mqtt.Start()
	t.Log("Connected")
	if err != nil {
		t.Error("Error connecting to broker ", err)
	}

	mqtt.SetMessageHandler(onMsg)
	time.Sleep(time.Second * 1)

	ctx, err := model.NewContextDB("TestIfFlow.db")

	flowMeta := model.FlowMeta{Id: "TestIfFlow", Name: "If flow test"}

	node := model.MetaNode{Id: "1", Label: "Button trigger 1", Type: "trigger", Address: "pt:j1/mt:evt/rt:dev/rn:test/ad:1/sv:sensor_lumin/ad:199_0", Service: "sensor_lumin", ServiceInterface: "evt.sensor.report", SuccessTransition: "1.1"}
	flowMeta.Nodes = append(flowMeta.Nodes, node)
	node = model.MetaNode{Id: "1.2", Label: "Button trigger 1", Type: "trigger", Address: "pt:j1/mt:evt/rt:dev/rn:test/ad:1/sv:sensor_lumin/ad:300_0", Service: "sensor_lumin", ServiceInterface: "evt.sensor.report", SuccessTransition: "1.1"}
	flowMeta.Nodes = append(flowMeta.Nodes, node)

	node = model.MetaNode{Id: "1.1", Label: "IF node", Type: "if", Config: flownode.IFExpressions{TrueTransition: "2", FalseTransition: "3", Expression: []flownode.IFExpression{
		{RightVariable: model.Variable{Value: int64(100), ValueType: "int"}, Operand: "gt", BooleanOperator: "and"},
		{RightVariable: model.Variable{Value: int64(200), ValueType: "int"}, Operand: "lt"}}}}
	flowMeta.Nodes = append(flowMeta.Nodes, node)

	node = model.MetaNode{Id: "2", Label: "Bulb 1.Room light intensity is > 100 lux", Type: "action", Address: "pt:j1/mt:cmd/rt:dev/rn:test/ad:1/sv:out_bin_switch/ad:200_0", Service: "out_bin_switch", ServiceInterface: "cmd.binary.set", SuccessTransition: "4",
		Config: flownode.ActionNodeConfig{DefaultValue:model.Variable{ValueType: "bool", Value: true}}}
	flowMeta.Nodes = append(flowMeta.Nodes, node)

	node = model.MetaNode{Id: "3", Label: "Bulb 2.Room light intensity is < 100 lux", Type: "action", Address: "pt:j1/mt:cmd/rt:dev/rn:test/ad:1/sv:out_bin_switch/ad:200_0", Service: "out_bin_switch", ServiceInterface: "cmd.binary.set", SuccessTransition: "5",
		Config: flownode.ActionNodeConfig{DefaultValue:model.Variable{ValueType: "bool", Value: true}}}
	flowMeta.Nodes = append(flowMeta.Nodes, node)

	node = model.MetaNode{Id: "4", Label: "Set variable", Type: "set_variable", SuccessTransition: "",
		Config: flownode.SetVariableNodeConfig{Name: "mode", UpdateGlobal: false, UpdateInputMsg: false, PersistOnUpdate: true, DefaultValue: model.Variable{Value: "correct", ValueType: "string"}}}
	flowMeta.Nodes = append(flowMeta.Nodes, node)

	node = model.MetaNode{Id: "5", Label: "Set variable", Type: "set_variable", SuccessTransition: "",
		Config: flownode.SetVariableNodeConfig{Name: "mode", UpdateGlobal: false, UpdateInputMsg: false, PersistOnUpdate: true, DefaultValue: model.Variable{Value: "wrong", ValueType: "string"}}}
	flowMeta.Nodes = append(flowMeta.Nodes, node)

	//data, err := json.Marshal(flowMeta)
	//if err == nil {
	//	ioutil.WriteFile("testflow.json", data, 0644)
	//}

	flow := NewFlow(flowMeta, ctx)
	flow.LoadAndConfigureAllNodes()
	flow.Start()
	time.Sleep(time.Second * 1)
	// send msg

	msg := fimpgo.NewIntMessage("evt.sensor.report", "sensor_lumin", 150, nil, nil, nil)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "test", ResourceAddress: "1", ServiceName: "sensor_lumin", ServiceAddress: "199_0"}
	mqtt.Publish(&adr, msg)

	// end
	time.Sleep(time.Second * 1)
	variable, err := flow.GetContext().GetVariable("mode", "TestIfFlow")
	if err != nil {
		t.Error("Variable is not set", err)
	}
	if variable.Value.(string) == "correct" {
		t.Log("Ok , variable is = ", variable.Value.(string))
	} else {
		t.Error("Wrong value")
	}
	flow.Stop()
	os.Remove("TestIfFlow.db")
}

//
func TestNewFlow3(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	mqtt := fimpgo.NewMqttTransport("tcp://localhost:1883", "flow_test", "", "", true, 1, 1)
	err := mqtt.Start()
	t.Log("Connected")
	if err != nil {
		t.Error("Error connecting to broker ", err)
	}

	mqtt.SetMessageHandler(onMsg)
	time.Sleep(time.Second * 1)

	ctx, err := model.NewContextDB("TestNewFlow3.db")
	flowMeta := model.FlowMeta{Id: "TestNewFlow3"}

	node := model.MetaNode{Id: "1", Label: "Button trigger", Type: "trigger", Address: "pt:j1/mt:evt/rt:dev/rn:test/ad:1/sv:out_bin_switch/ad:199_0", Service: "out_bin_switch", ServiceInterface: "evt.binary.report", SuccessTransition: "1.1"}
	flowMeta.Nodes = append(flowMeta.Nodes, node)

	node = model.MetaNode{Id: "1.1", Label: "IF node", Type: "if", Config: flownode.IFExpressions{TrueTransition: "2", FalseTransition: "3", Expression: []flownode.IFExpression{
		{RightVariable: model.Variable{Value: false, ValueType: "bool"}, Operand: "eq"}}}}
	flowMeta.Nodes = append(flowMeta.Nodes, node)

	node = model.MetaNode{Id: "2", Label: "Lights ON", Type: "action", Address: "pt:j1/mt:cmd/rt:dev/rn:test/ad:1/sv:out_bin_switch/ad:200_0", Service: "out_bin_switch", ServiceInterface: "cmd.binary.set", SuccessTransition: "",
		Config: flownode.ActionNodeConfig{DefaultValue:model.Variable{ValueType: "bool", Value: true}}}
	flowMeta.Nodes = append(flowMeta.Nodes, node)

	node = model.MetaNode{Id: "3", Label: "Lights OFF", Type: "action", Address: "pt:j1/mt:cmd/rt:dev/rn:test/ad:1/sv:out_bin_switch/ad:200_0", Service: "out_bin_switch", ServiceInterface: "cmd.binary.set", SuccessTransition: "",
		Config: flownode.ActionNodeConfig{DefaultValue:model.Variable{ValueType: "bool", Value: true}}}
	flowMeta.Nodes = append(flowMeta.Nodes, node)

	conReg := connector.NewRegistry("../testdata/var/connectors")
	if err := conReg.LoadInstancesFromDisk();err != nil {
		t.Error("Failed init registry")
	}
	time.Sleep(time.Second * 1)
	flow := NewFlow(flowMeta, ctx)
	flow.SetConnectorRegistry(conReg)
	flow.Start()
	time.Sleep(time.Second * 1)
	// send msg

	msg := fimpgo.NewBoolMessage("evt.binary.report", "out_bin_switch", true, nil, nil, nil)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "test", ResourceAddress: "1", ServiceName: "out_bin_switch", ServiceAddress: "199_0"}
	mqtt.Publish(&adr, msg)
	time.Sleep(time.Second * 1)
	flow.Stop()
	// end
	time.Sleep(time.Second * 2)
	os.Remove("TestNewFlow3.db")

}

func TestSetVariableFlow(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	mqtt := fimpgo.NewMqttTransport("tcp://localhost:1883", "flow_test", "", "", true, 1, 1)
	err := mqtt.Start()
	t.Log("Connected")
	if err != nil {
		t.Error("Error connecting to broker ", err)
	}

	mqtt.SetMessageHandler(onMsg)
	time.Sleep(time.Second * 1)

	ctx, err := model.NewContextDB("TestSetVariableFlow.db")
	flowMeta := model.FlowMeta{Id: "TestSetVariableFlow"}

	node := model.MetaNode{Id: "1", Label: "Button trigger", Type: "trigger", Address: "pt:j1/mt:evt/rt:dev/rn:test/ad:1/sv:out_bin_switch/ad:199_0", Service: "out_bin_switch", ServiceInterface: "evt.binary.report", SuccessTransition: "2"}
	flowMeta.Nodes = append(flowMeta.Nodes, node)

	node = model.MetaNode{Id: "2", Label: "Set variable", Type: "set_variable", SuccessTransition: "",
		Config: flownode.SetVariableNodeConfig{Name: "volume", UpdateGlobal: false, UpdateInputMsg: false, PersistOnUpdate: true, DefaultValue: model.Variable{Value: 65, ValueType: "int"}}}
	flowMeta.Nodes = append(flowMeta.Nodes, node)
	//data, err := json.Marshal(flowMeta)
	//if err == nil {
	//	ioutil.WriteFile("testflow2.json", data, 0644)
	//}
	flow := NewFlow(flowMeta, ctx)
	flow.LoadAndConfigureAllNodes()
	flow.Start()
	time.Sleep(time.Second * 1)
	// send msg

	msg := fimpgo.NewBoolMessage("evt.binary.report", "out_bin_switch", true, nil, nil, nil)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "test", ResourceAddress: "1", ServiceName: "out_bin_switch", ServiceAddress: "199_0"}
	mqtt.Publish(&adr, msg)
	time.Sleep(time.Second * 1)
	variable, err := flow.GetContext().GetVariable("volume", "TestSetVariableFlow")
	if err != nil {
		t.Error("Variable is not set", err)
	}
	if variable.Value.(int) != 65 {
		t.Error("Wrong value")
	} else {
		t.Log("Ok , variable is = ", variable.Value.(int))
	}
	flow.Stop()
	// end
	time.Sleep(time.Second * 2)
	os.Remove("TestSetVariableFlow.db")

}

func TestTransformFlipFlow(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	mqtt := fimpgo.NewMqttTransport("tcp://localhost:1883", "flow_test", "", "", true, 1, 1)
	err := mqtt.Start()
	t.Log("Connected")
	if err != nil {
		t.Error("Error connecting to broker ", err)
	}

	mqtt.SetMessageHandler(onMsg)
	time.Sleep(time.Second * 1)

	ctx, err := model.NewContextDB("TestTransform.db")
	flowMeta := model.FlowMeta{Id: "TestTransformFlow"}

	node := model.MetaNode{Id: "1", Label: "Button trigger", Type: "trigger", Address: "pt:j1/mt:evt/rt:dev/rn:test/ad:1/sv:out_bin_switch/ad:199_0", Service: "out_bin_switch", ServiceInterface: "evt.binary.report", SuccessTransition: "2"}
	flowMeta.Nodes = append(flowMeta.Nodes, node)

	node = model.MetaNode{Id: "2", Label: "Set variable", Type: "transform", SuccessTransition: "",
		Config:flownode.TransformNodeConfig{Operation:"flip"}}
	flowMeta.Nodes = append(flowMeta.Nodes, node)
	flow := NewFlow(flowMeta, ctx)
	flow.LoadAndConfigureAllNodes()
	flow.Start()
	time.Sleep(time.Second * 1)
	msg := fimpgo.NewBoolMessage("evt.binary.report", "out_bin_switch", false, nil, nil, nil)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "test", ResourceAddress: "1", ServiceName: "out_bin_switch", ServiceAddress: "199_0"}
	mqtt.Publish(&adr, msg)
	time.Sleep(time.Second * 1)
	//variable, err := flow.GetContext().GetVariable("volume", "TestSetVariableFlow")
	//inputMessage := flow.GetCurrentMessage()
	//if inputMessage.Payload.Value.(bool) != true {
	//	t.Error("Wrong value " )
	//}
	flow.Stop()
	// end
	time.Sleep(time.Second * 2)
	os.Remove("TestTransform.db")

}

func TestRestActionFlow(t *testing.T) {
	dbName := "RestActionDb.db"
	log.SetLevel(log.DebugLevel)
	mqtt := fimpgo.NewMqttTransport("tcp://localhost:1883", "flow_test", "", "", true, 1, 1)
	err := mqtt.Start()
	t.Log("Connected")
	if err != nil {
		t.Error("Error connecting to broker ", err)
	}

	mqtt.SetMessageHandler(onMsg)
	time.Sleep(time.Second * 1)

	ctx, err := model.NewContextDB(dbName)
	flowMeta := model.FlowMeta{Id: "TestRestActionFlow"}

	node := model.MetaNode{Id: "1", Label: "Button trigger", Type: "trigger", Address: "pt:j1/mt:evt/rt:dev/rn:test/ad:1/sv:out_bin_switch/ad:199_0", Service: "out_bin_switch", ServiceInterface: "evt.binary.report", SuccessTransition: "2"}
	flowMeta.Nodes = append(flowMeta.Nodes, node)

	//node = model.MetaNode{Id: "2", Label: "Invoke httpbin", Type: "rest_action", SuccessTransition: "",
	//	Config:flownode.RestActionNodeConfig{Method:"POST",Url:"https://httpbin.org/post",RequestTemplate:"{'param1':{{.Variable}} }"}}

	node = model.MetaNode{Id: "2", Label: "Turn off yamaha", Type: "rest_action", SuccessTransition: "",
		Config:flownode.RestActionNodeConfig{Method:"POST",Url:"http://yamaha.st/YamahaRemoteControl/ctrl",
		RequestTemplate:"<YAMAHA_AV cmd=\"PUT\"><Main_Zone><Power_Control><Power>On</Power></Power_Control></Main_Zone></YAMAHA_AV>"}}

	flowMeta.Nodes = append(flowMeta.Nodes, node)
	flow := NewFlow(flowMeta, ctx)
	flow.LoadAndConfigureAllNodes()
	flow.Start()
	time.Sleep(time.Second * 1)
	msg := fimpgo.NewBoolMessage("evt.binary.report", "out_bin_switch", false, nil, nil, nil)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "test", ResourceAddress: "1", ServiceName: "out_bin_switch", ServiceAddress: "199_0"}
	mqtt.Publish(&adr, msg)
	time.Sleep(time.Second * 1)
	//variable, err := flow.GetContext().GetVariable("volume", "TestSetVariableFlow")
	//inputMessage := flow.GetCurrentMessage()
	//if inputMessage.Payload.Value.(bool) != true {
	//	t.Error("Wrong value " )
	//}
	flow.Stop()
	// end
	time.Sleep(time.Second * 2)
	os.Remove(dbName)

}

func TestTransformAddFlow(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	mqtt := fimpgo.NewMqttTransport("tcp://localhost:1883", "flow_test", "", "", true, 1, 1)
	err := mqtt.Start()
	t.Log("Connected")
	if err != nil {
		t.Error("Error connecting to broker ", err)
	}

	mqtt.SetMessageHandler(onMsg)
	time.Sleep(time.Second * 1)

	ctx, err := model.NewContextDB("TestTransform.db")
	flowMeta := model.FlowMeta{Id: "TestTransformFlow"}

	node := model.MetaNode{Id: "1", Label: "Sensor trigger", Type: "trigger", Address: "pt:j1/mt:evt/rt:dev/rn:test/ad:1/sv:sensor_temp/ad:199_0", Service: "sensor_temp", ServiceInterface: "evt.sensor.report", SuccessTransition: "2"}
	flowMeta.Nodes = append(flowMeta.Nodes, node)

	node = model.MetaNode{Id: "2", Label: "Add transform", Type: "transform", SuccessTransition: "",
		Config:flownode.TransformNodeConfig{Operation:"add",RValue:model.Variable{ValueType:"int",Value:int(2)}}}
	flowMeta.Nodes = append(flowMeta.Nodes, node)
	flow := NewFlow(flowMeta, ctx)
	flow.LoadAndConfigureAllNodes()
	flow.Start()
	time.Sleep(time.Second * 1)
	msg := fimpgo.NewFloatMessage("evt.sensor.report", "sensor_temp", 12.5, nil, nil, nil)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "test", ResourceAddress: "1", ServiceName: "sensor_temp", ServiceAddress: "199_0"}
	mqtt.Publish(&adr, msg)
	time.Sleep(time.Second * 1)
	//variable, err := flow.GetContext().GetVariable("volume", "TestSetVariableFlow")
	//inputMessage := flow.GetCurrentMessage()
	//t.Log("Result = ",inputMessage.Payload.Value)
	//if inputMessage.Payload.Value.(float64) != 14.5 {
	//	t.Error("Wrong value " )
	//}
	flow.Stop()
	// end
	time.Sleep(time.Second * 2)
	os.Remove("TestTransform.db")

}

func TestReceiveFlow(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	mqtt := fimpgo.NewMqttTransport("tcp://localhost:1883", "flow_test", "", "", true, 1, 1)
	err := mqtt.Start()
	t.Log("Connected")
	if err != nil {
		t.Error("Error connecting to broker ", err)
	}

	mqtt.SetMessageHandler(onMsg)
	time.Sleep(time.Second * 1)

	ctx, err := model.NewContextDB("TestReceiveFlow.db")
	flowMeta := model.FlowMeta{Id: "TestReceiveFlow"}

	node := model.MetaNode{Id: "1", Label: "Button trigger", Type: "trigger", Address: "pt:j1/mt:evt/rt:dev/rn:test/ad:1/sv:out_bin_switch/ad:199_0", Service: "out_bin_switch", ServiceInterface: "evt.binary.report",
		SuccessTransition: "2"}
	flowMeta.Nodes = append(flowMeta.Nodes, node)

	node = model.MetaNode{Id: "2", Label: "Receive", Type: "receive", Address: "pt:j1/mt:evt/rt:dev/rn:test/ad:1/sv:out_bin_switch/ad:200_0", Service: "out_bin_switch", ServiceInterface: "evt.binary.report",
		SuccessTransition: "3",TimeoutTransition:"5" , Config:flownode.ReceiveConfig{Timeout:5}}
	flowMeta.Nodes = append(flowMeta.Nodes, node)

	node = model.MetaNode{Id: "3", Label: "Receive", Type: "receive", Address: "pt:j1/mt:evt/rt:dev/rn:test/ad:1/sv:out_bin_switch/ad:201_0", Service: "out_bin_switch", ServiceInterface: "evt.binary.report",
		SuccessTransition: "4",TimeoutTransition:"5" , Config:flownode.ReceiveConfig{Timeout:1}}
	flowMeta.Nodes = append(flowMeta.Nodes, node)

	node = model.MetaNode{Id: "4", Label: "Set variable", Type: "set_variable", SuccessTransition: "",
		Config: flownode.SetVariableNodeConfig{Name: "status", UpdateGlobal: false, UpdateInputMsg: false, PersistOnUpdate: true, DefaultValue: model.Variable{Value: "in_time", ValueType: "string"}}}
	flowMeta.Nodes = append(flowMeta.Nodes, node)

	node = model.MetaNode{Id: "5", Label: "Set variable", Type: "set_variable", SuccessTransition: "",
		Config: flownode.SetVariableNodeConfig{Name: "status", UpdateGlobal: false, UpdateInputMsg: false, PersistOnUpdate: true, DefaultValue: model.Variable{Value: "timeout", ValueType: "string"}}}
	flowMeta.Nodes = append(flowMeta.Nodes, node)
	//data, err := json.Marshal(flowMeta)
	//if err == nil {
	//	ioutil.WriteFile("testflow2.json", data, 0644)
	//}
	flow := NewFlow(flowMeta, ctx)
	flow.LoadAndConfigureAllNodes()
	flow.Start()
	time.Sleep(time.Second * 1)
	// send msg

	msg := fimpgo.NewBoolMessage("evt.binary.report", "out_bin_switch", true, nil, nil, nil)
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "test", ResourceAddress: "1", ServiceName: "out_bin_switch", ServiceAddress: "199_0"}
	mqtt.Publish(&adr, msg)
	//time.Sleep(time.Millisecond * 10)
	msg = fimpgo.NewBoolMessage("evt.binary.report", "out_bin_switch", true, nil, nil, nil)
	adr = fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "test", ResourceAddress: "1", ServiceName: "out_bin_switch", ServiceAddress: "200_0"}
	mqtt.Publish(&adr, msg)
	//time.Sleep(time.Second * 2)
	msg = fimpgo.NewBoolMessage("evt.binary.report", "out_bin_switch", true, nil, nil, nil)
	adr = fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "test", ResourceAddress: "1", ServiceName: "out_bin_switch", ServiceAddress: "201_0"}
	mqtt.Publish(&adr, msg)

	time.Sleep(time.Second * 1)
	variable, err := flow.GetContext().GetVariable("status", "TestReceiveFlow")
	if err != nil {
		t.Error("Variable is not set", err)
	}else if variable.Value.(string) == "in_time" {
		t.Log("Ok , result is in time")
	} else {
		t.Error("Error timed out")
	}
	flow.Stop()
	// end
	time.Sleep(time.Second * 2)
	os.Remove("TestReceiveFlow.db")

}

func TestLoopFlow(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	mqtt := fimpgo.NewMqttTransport("tcp://localhost:1883", "flow_test", "", "", true, 1, 1)
	err := mqtt.Start()
	t.Log("Connected")
	if err != nil {
		t.Error("Error connecting to broker ", err)
	}

	mqtt.SetMessageHandler(onMsg)
	time.Sleep(time.Second * 1)

	ctx, err := model.NewContextDB("TestLoopFlow.db")
	flowMeta := model.FlowMeta{Id: "TestLoopFlow"}

	node := model.MetaNode{Id: "1", Label: "Button trigger", Type: "trigger", Address: "pt:j1/mt:evt/rt:dev/rn:test/ad:1/sv:out_bin_switch/ad:199_0", Service: "out_bin_switch", ServiceInterface: "evt.binary.report",
		SuccessTransition: "2"}
	flowMeta.Nodes = append(flowMeta.Nodes, node)

	node = model.MetaNode{Id: "2", Label: "Loop", Type: "loop", SuccessTransition: "4",ErrorTransition:"5",
		Config: flownode.LoopNodeConfig{StartValue:0,EndValue:4}}
	flowMeta.Nodes = append(flowMeta.Nodes, node)

	node = model.MetaNode{Id: "4", Label: "Set variable", Type: "set_variable", SuccessTransition: "",
		Config: flownode.SetVariableNodeConfig{Name: "status", UpdateGlobal: false, UpdateInputMsg: false, PersistOnUpdate: true, DefaultValue: model.Variable{Value: "counting", ValueType: "string"}}}
	flowMeta.Nodes = append(flowMeta.Nodes, node)

	node = model.MetaNode{Id: "5", Label: "Set variable", Type: "set_variable", SuccessTransition: "",
		Config: flownode.SetVariableNodeConfig{Name: "status", UpdateGlobal: false, UpdateInputMsg: false, PersistOnUpdate: true, DefaultValue: model.Variable{Value: "reset", ValueType: "string"}}}
	flowMeta.Nodes = append(flowMeta.Nodes, node)
	//data, err := json.Marshal(flowMeta)
	//if err == nil {
	//	ioutil.WriteFile("testflow2.json", data, 0644)
	//}
	flow := NewFlow(flowMeta, ctx)
	flow.LoadAndConfigureAllNodes()
	flow.Start()
	time.Sleep(time.Second * 1)
	// send msg

	for i:=0;i<4;i++ {
		msg := fimpgo.NewBoolMessage("evt.binary.report", "out_bin_switch", true, nil, nil, nil)
		adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "test", ResourceAddress: "1", ServiceName: "out_bin_switch", ServiceAddress: "199_0"}
		mqtt.Publish(&adr, msg)
		time.Sleep(time.Millisecond * 10)
	}

	//time.Sleep(time.Millisecond * 10)


	time.Sleep(time.Second * 1)
	variable, err := flow.GetContext().GetVariable("status", "TestLoopFlow")
	if err != nil {
		t.Error("Variable is not set", err)
	}else if variable.Value.(string) == "reset" {
		t.Log("Ok ")
	} else {
		t.Error("Error.")
	}
	flow.Stop()
	// end
	time.Sleep(time.Second * 2)
	os.Remove("TestLoopFlow.db")

}

func TestTimeTriggerFlow(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	mqtt := fimpgo.NewMqttTransport("tcp://localhost:1883", "flow_test", "", "", true, 1, 1)
	err := mqtt.Start()
	t.Log("Connected")
	if err != nil {
		t.Error("Error connecting to broker ", err)
	}

	mqtt.SetMessageHandler(onMsg)
	time.Sleep(time.Second * 1)

	ctx, err := model.NewContextDB("TestTimeTriggerFlow.db")
	flowMeta := model.FlowMeta{Id: "TestTimeTriggerFlow"}

	node := model.MetaNode{Id: "1", Label: "Turn ever 1 second", Type: "time_trigger",
		SuccessTransition: "2", Config:flownode.TimeTriggerConfig{Expressions:[]flownode.TimeExpression{ {Name:"every second",Expression:"@every 1s"} } }}
	flowMeta.Nodes = append(flowMeta.Nodes, node)

	node = model.MetaNode{Id: "2", Label: "Loop", Type: "loop", SuccessTransition: "4",ErrorTransition:"5",
		Config: flownode.LoopNodeConfig{StartValue:0,EndValue:3}}
	flowMeta.Nodes = append(flowMeta.Nodes, node)

	node = model.MetaNode{Id: "4", Label: "Set variable", Type: "set_variable", SuccessTransition: "",
		Config: flownode.SetVariableNodeConfig{Name: "status", UpdateGlobal: false, UpdateInputMsg: false, PersistOnUpdate: true, DefaultValue: model.Variable{Value: "counting", ValueType: "string"}}}
	flowMeta.Nodes = append(flowMeta.Nodes, node)

	node = model.MetaNode{Id: "5", Label: "Set variable", Type: "set_variable", SuccessTransition: "",
		Config: flownode.SetVariableNodeConfig{Name: "status", UpdateGlobal: false, UpdateInputMsg: false, PersistOnUpdate: true, DefaultValue: model.Variable{Value: "reset", ValueType: "string"}}}
	flowMeta.Nodes = append(flowMeta.Nodes, node)
	//data, err := json.Marshal(flowMeta)
	//if err == nil {
	//	ioutil.WriteFile("testflow2.json", data, 0644)
	//}
	flow := NewFlow(flowMeta, ctx)
	flow.LoadAndConfigureAllNodes()
	flow.Start()
	time.Sleep(time.Second * 3)
	// send msg
	variable, err := flow.GetContext().GetVariable("status", "TestTimeTriggerFlow")
	if err != nil {
		t.Error("Variable is not set", err)
	}else if variable.Value.(string) == "reset" {
		t.Log("Ok ")
	} else {
		t.Error("Error.")
	}
	flow.Stop()
	// end
	time.Sleep(time.Second * 2)
	os.Remove("TestTimeTriggerFlow.db")

}

