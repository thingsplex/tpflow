package flow

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/alivinco/fimpgo"
	"github.com/alivinco/tpflow"
	"testing"
	"time"
)

func TestManager_LoadFlowFromFile(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	config := tpflow.Configs{FlowStorageDir: "./testdata/var/flow_storage", ConnectorStorageDir: "./testdata/var/connectors"}
	man, err := NewManager(config)
	if err != nil {
		t.Error(err)
	}
	man.LoadFlowFromFile("testflow.json")

	mqtt := fimpgo.NewMqttTransport("tcp://localhost:1883", "flow_test", "", "", true, 1, 1)
	err = mqtt.Start()
	t.Log("Connected")
	if err != nil {
		t.Error("Error while connecting to broker ", err)
	}
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "test", ResourceAddress: "1", ServiceName: "sensor_lumin", ServiceAddress: "199_0"}

	msg := fimpgo.NewIntMessage("evt.sensor.report", "sensor_lumin", 50, nil, nil, nil)
	mqtt.Publish(&adr, msg)

	msg = fimpgo.NewIntMessage("evt.sensor.report", "sensor_lumin", 100, nil, nil, nil)
	mqtt.Publish(&adr, msg)
	time.Sleep(time.Second * 1)
	//man.DeleteFlow("123")
	//man.LoadFlowFromFile("testflow.json")
	msg = fimpgo.NewIntMessage("evt.sensor.report", "sensor_lumin", 150, nil, nil, nil)
	mqtt.Publish(&adr, msg)

	// end
	time.Sleep(time.Second * 5)
}

func TestManager_LoadAllFlowsFromStorage(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	config := tpflow.Configs{FlowStorageDir: "./flows"}
	man, err := NewManager(config)
	if err != nil {
		t.Error(err)
	}
	man.LoadAllFlowsFromStorage()

	mqtt := fimpgo.NewMqttTransport("tcp://localhost:1883", "flow_test", "", "", true, 1, 1)
	err = mqtt.Start()
	t.Log("Connected")
	if err != nil {
		t.Error("Error while connecting to broker ", err)
	}
	adr := fimpgo.Address{MsgType: fimpgo.MsgTypeEvt, ResourceType: fimpgo.ResourceTypeDevice, ResourceName: "test", ResourceAddress: "1", ServiceName: "sensor_lumin", ServiceAddress: "199_0"}

	msg := fimpgo.NewIntMessage("evt.sensor.report", "sensor_lumin", 50, nil, nil, nil)
	mqtt.Publish(&adr, msg)

	msg = fimpgo.NewIntMessage("evt.sensor.report", "sensor_lumin", 100, nil, nil, nil)
	mqtt.Publish(&adr, msg)
	time.Sleep(time.Second * 1)
	//man.DeleteFlow("123")
	//man.LoadFlowFromFile("testflow.json")
	msg = fimpgo.NewIntMessage("evt.sensor.report", "sensor_lumin", 150, nil, nil, nil)
	mqtt.Publish(&adr, msg)

	// end
	time.Sleep(time.Second * 5)
}

func TestManager_GenerateNewFlow(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	config := tpflow.Configs{FlowStorageDir: "../var/flow_storage"}
	man, err := NewManager(config)
	if err != nil {
		t.Error(err)
	}
	flow := man.GenerateNewFlow()
	data, _ := json.Marshal(flow)
	man.UpdateFlowFromJsonAndSaveToStorage(flow.Id, data)
}

func TestManager_GetFlowList(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	config := tpflow.Configs{FlowStorageDir: "./flows"}
	man, err := NewManager(config)
	if err != nil {
		t.Error(err)
	}
	man.LoadAllFlowsFromStorage()
	flows := man.GetFlowList()
	t.Log(flows)
	if len(flows) == 0 {
		t.Error("List is empty.")
	}

}
